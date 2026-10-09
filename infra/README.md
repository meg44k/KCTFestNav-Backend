# 本番の作り方(infra)

| 役割 | 使うもの |
|---|---|
| フロント | Vercel(東京 `hnd1`)。設定はフロントの README |
| API | Cloud Run `kctfestnav-api`(`asia-northeast1`)→ `https://api.kctfes.app` |
| DB | Cloud SQL for MySQL 8.4(`db-f1-micro`) |
| Redis | Upstash(東京・TLS) |
| 写真 | Cloudflare R2 `kctfestnav-photos` → `https://img.kctfes.app` |
| DNS | Cloudflare(ドメインはお名前.com で取り、ネームサーバーを Cloudflare に向ける) |

Terraform が作るのは GCP と Cloudflare(DNS・R2)。Vercel と Upstash は画面で作る。
状態ファイルは GCS のバケット(Terraform の外で作る)に置く。

## 1. 用意するもの

```sh
brew install --cask google-cloud-sdk
brew install terraform cloud-sql-proxy mysql-client
# mysql-client は PATH に入らないので: export PATH=/opt/homebrew/opt/mysql-client/bin:$PATH
```

- GCP のプロジェクトと請求先アカウント
- Cloudflare のアカウント
- ドメイン `kctfes.app`(お名前.com)

## 2. 最初の 1 回だけ

1. ドメインを取り、Cloudflare に「サイトを追加」して、登録業者でネームサーバーを Cloudflare のものに変える。
   Cloudflare の **zone ID** と **account ID** を控える
2. GCP にログインする

   ```sh
   gcloud auth login
   gcloud auth application-default login
   gcloud config set project <project>
   ```

3. 状態ファイルのバケットを作る(Terraform の外。`terraform destroy` で消えない)

   ```sh
   gcloud storage buckets create gs://<project>-tfstate --location=asia-northeast1 --uniform-bucket-level-access
   gcloud storage buckets update gs://<project>-tfstate --versioning
   ```

4. Cloudflare で API トークンを作る(権限: Zone → DNS → 編集、Account → Workers R2 Storage → 編集)。
   ファイルには書かず、使うシェルで `export CLOUDFLARE_API_TOKEN=...`
5. Upstash で Redis を作る(リージョンは東京、TLS あり)。Endpoint と Port を `redis_addr`(`xxxx.upstash.io:6379`)に使う
6. `api.kctfes.app` を Cloud Run にひも付けるため、ドメインの持ち主を確かめる

   ```sh
   gcloud domains verify kctfes.app
   ```

7. 変数のファイルを作る

   ```sh
   cd infra
   cp terraform.tfvars.example terraform.tfvars   # 値を入れる。git には入らない
   ```

## 3. 作る(順番が大事)

Cloud Run は秘密の値の「版」が無いと起動できない。入れ物 → 値 → 全部 の順に作る。

```sh
terraform init -backend-config="bucket=<project>-tfstate"

# 1) 秘密の値の入れ物だけ
terraform apply -target=google_secret_manager_secret.s

# 2) 手で値を入れる(画面にもシェルの履歴にも残さない)
read -s P && printf %s "$P" | gcloud secrets versions add redis-password --data-file=- ; unset P
read -s P && printf %s "$P" | gcloud secrets versions add init-admin-password --data-file=- ; unset P

# 3) 残り全部(plan を読んでから)
terraform plan
terraform apply
```

`db-pass` と `jwt-secret` は Terraform が作って入れる(値は状態ファイルにだけある)。

GitHub Actions が使う値を、リポジトリの変数に入れる:

```sh
gh variable set GCP_PROJECT_ID -b <project>
gh variable set GCP_WIF_PROVIDER -b "$(terraform output -raw wif_provider)"
gh variable set GCP_DEPLOY_SA -b "$(terraform output -raw deploy_service_account)"
```

## 4. DB の形を入れる・変える

Cloud SQL Auth Proxy で手元からつなぐ。

```sh
cloud-sql-proxy "$(terraform output -raw sql_connection_name)" --port 3307 &
# パスワードはコマンドの引数に書かない(ps で見える)
export MYSQL_PWD="$(gcloud secrets versions access latest --secret=db-pass)"
# MySQL 8.4 の caching_sha2_password は、Proxy 越し(TLS なし)だとサーバーの公開鍵が要る
m() { mysql --get-server-public-key --default-character-set=utf8mb4 -h127.0.0.1 -P3307 -ukctfestnav kctfestnav "$@"; }

# 最初の 1 回だけ: 最新の形
m < ../db/schema.sql

# 以後の変更: db/migrations/ のファイルを同じ要領で(m < ../db/migrations/xxxx.sql)
unset MYSQL_PWD
kill %1
```

- 開発用の仮のデータ `db/init/03-seed.sql` は流さない
- `db/seed/*.sql` は本物のデータか確かめてから流す
- 最初の管理者は Cloud Run の起動時に `INIT_ADMIN_ID` と `init-admin-password` から作られる

### いいねの合言葉を Vercel に入れる

`internal-api-key` は Terraform が作る。Vercel の Settings → Environment Variables に
`INTERNAL_API_KEY`(Production だけ、`NEXT_PUBLIC_` は付けない)として入れる。値は次でクリップボードに写す(画面には出さない):

```sh
gcloud secrets versions access latest --secret=internal-api-key | pbcopy 
```

いいねの表は `m < ../db/migrations/2026-10-09-likes.sql` で本番の DB に足す(上の Proxy の手順で)。

## 5. 出す

develop を main にマージすると、GitHub Actions(`.github/workflows/deploy.yml`)がテストして Cloud Run に出す。

```sh
curl -s "$(terraform output -raw api_url)/booths"
curl -s https://api.kctfes.app/booths   # 証明書ができるまで(数十分)はつながらない
```

フロント(Vercel)の `NEXT_PUBLIC_API_BASE_URL` には `terraform output -raw api_url`(`https://kctfestnav-api-....run.app`)を入れる。
API を呼ぶのは Vercel のサーバーだけなので、ブラウザ向けの名前は要らない。
`api.kctfes.app`(Cloud Run のドメインのひも付け)は Google がまだ「プレビュー」としている機能なので、本番の通り道には使わず、確かめる用に残す。

## 6. 当日(10/31・11/1)

最初の 1 台が起きるまでの数秒の待ちを無くすため、最低 1 台にする。

```sh
terraform plan -var min_instances=1    # 変わるのが scaling だけなのを確かめる
terraform apply -var min_instances=1
```

終わったら `-var min_instances=0` で戻す。

## 7. 文化祭のあと

1. DB の中身を書き出す(来年の参考)

   ```sh
   SA="$(gcloud sql instances describe kctfestnav --format='value(serviceAccountEmailAddress)')"
   gcloud storage buckets add-iam-policy-binding gs://<project>-tfstate --member="serviceAccount:$SA" --role=roles/storage.objectCreator
   gcloud sql export sql kctfestnav gs://<project>-tfstate/backup/kctfestnav-$(date +%F).sql.gz --database=kctfestnav
   ```

2. 写真のバケットを空にする(中身があると R2 のバケットは消せない)。残したい写真は先に手元へ落とす。
   Cloudflare の画面の R2 → `kctfestnav-photos` → 全部選んで削除

3. 片付ける(状態ファイルのバケットとドメインは残る)

   ```sh
   terraform destroy
   ```

## 費用の目安

動かす 1 か月で 1,500〜2,000 円(ほぼ Cloud SQL)+ ドメイン(お名前.com、初年度 1 円。2 年目からは更新料がかかる)。
予算アラート(月 3,000 円の 50%・90%・100%)が請求先アカウントの管理者にメールで届く。
