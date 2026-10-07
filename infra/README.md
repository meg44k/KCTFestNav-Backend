# 本番の作り方(infra)

| 役割 | 使うもの |
|---|---|
| フロント | Vercel(東京 `hnd1`)。設定はフロントの README |
| API | Cloud Run `kctfestnav-api`(`asia-northeast1`)→ `https://api.kctfest.jp` |
| DB | Cloud SQL for MySQL 8.4(`db-f1-micro`) |
| Redis | Upstash(東京・TLS) |
| 写真 | Cloudflare R2 `kctfestnav-photos` → `https://img.kctfest.jp` |
| DNS | Cloudflare(ドメインは .jp を扱う登録業者で取り、ネームサーバーを Cloudflare に向ける) |

Terraform が作るのは GCP と Cloudflare(DNS・R2)。Vercel と Upstash は画面で作る。
状態ファイルは GCS のバケット(Terraform の外で作る)に置く。

## 1. 用意するもの

```sh
brew install --cask google-cloud-sdk
brew install terraform cloud-sql-proxy mysql-client
```

- GCP のプロジェクトと請求先アカウント
- Cloudflare のアカウント
- ドメイン `kctfest.jp`

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
6. `api.kctfest.jp` を Cloud Run にひも付けるため、ドメインの持ち主を確かめる

   ```sh
   gcloud domains verify kctfest.jp
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
DBPASS="$(gcloud secrets versions access latest --secret=db-pass)"

# 最初の 1 回だけ: 最新の形
mysql --default-character-set=utf8mb4 -h127.0.0.1 -P3307 -ukctfestnav -p"$DBPASS" kctfestnav < ../db/schema.sql

# 以後の変更: db/migrations/ のファイルを同じ要領で
unset DBPASS
kill %1
```

- 開発用の仮のデータ `db/init/03-seed.sql` は流さない
- `db/seed/*.sql` は本物のデータか確かめてから流す
- 最初の管理者は Cloud Run の起動時に `INIT_ADMIN_ID` と `init-admin-password` から作られる

## 5. 出す

develop を main にマージすると、GitHub Actions(`.github/workflows/deploy.yml`)がテストして Cloud Run に出す。

```sh
curl -s https://api.kctfest.jp/booths
```

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

2. 片付ける(状態ファイルのバケットとドメインは残る)

   ```sh
   terraform destroy
   ```

## 費用の目安

動かす 1 か月で 1,500〜2,000 円(ほぼ Cloud SQL)+ ドメイン年 3,000〜4,000 円。
予算アラート(月 3,000 円の 50%・90%・100%)が請求先アカウントの管理者にメールで届く。
