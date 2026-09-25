# KCTFestNav-Backend

高専祭ナビゲーションアプリ「KCTFestNav」のバックエンド (Go / Echo)。
API仕様は [docs/openapi.yaml](docs/openapi.yaml) を参照。

## ローカル環境の立ち上げ

MySQL と Redis を Docker で用意する。

```sh
docker compose up -d      # MySQL(3306) と Redis(6379) を起動
go run ./cmd              # APIサーバーを http://localhost:1323 で起動
```

初回起動時に `db/init/` 以下が名前順で実行され、次の状態になる。

| DB | 用途 | 初期データ |
|---|---|---|
| `kctfestnav` | 開発用 | `db/init/03-seed.sql` のブース8件 |
| `kctfest_test_handler` | `internal/handler` のe2eテスト用 | なし(テストが毎回TRUNCATEする) |
| `kctfest_test_repository` | `internal/repository` のテスト用 | なし(同上) |

テストDBをパッケージごとに分けているのは、`go test ./...` がパッケージを
並列に実行するため。同じDBを共有すると TRUNCATE とデータ投入が競合して
テストが不安定になる。Redis も同じ理由で論理DBを分けている。

| Redis の論理DB | 用途 |
|---|---|
| 0 | アプリ本体 |
| 1 | `internal/handler` のテスト |
| 2 | `internal/repository` のテスト |

動作確認:

```sh
curl -s http://localhost:1323/booths | python3 -m json.tool
```

### 注意点

- 開発用のため root はパスワード無し。`.env` の `DB_USER=root` / `DB_PASS=`(空) と、
  e2eテストがハードコードしている DSN に合わせている
- `db/init/` はボリュームが空のときだけ実行される。スキーマやシードを変えたら
  `docker compose down -v && docker compose up -d` で作り直す
- 混雑度は MySQL ではなく Redis (`congestion_status:{id}`) で管理している。
  未登録のブースは「空いている」扱いになる
- コンテナ内の mysql クライアントは既定で `character_set_client=latin1` になり、
  UTF-8のSQLを流すと二重エンコードで文字化けする。
  `docker/mysql-charset.cnf` で utf8mb4 に固定している

## テスト

```sh
go test ./...
```

`internal/handler/*_e2e_test.go` は実際のMySQLとRedisに接続する。
**未起動の場合は自動でスキップされる**ため、テストが緑でも
e2eが走っているとは限らない。実行されたか確かめるには:

```sh
go test -v -count=1 ./internal/handler/ -run E2E
```

## スキーマ変更の手順

1. `db/schema.sql` と `db/query.sql` を編集
2. `sqlc generate` で `internal/database/` を再生成
3. 既存DBには手動でマイグレーションを当てる(マイグレーションツールは未導入)
