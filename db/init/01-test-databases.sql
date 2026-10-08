-- e2e / リポジトリテスト用のDB。
-- go test ./... はパッケージを並列に実行するため、同じDBを共有すると
-- TRUNCATE とデータ投入が競合してテストが不安定になる。
-- そのためテストパッケージごとにDBを分けている。
SET NAMES utf8mb4;

CREATE DATABASE IF NOT EXISTS kctfest_test_handler;    -- internal/handler
CREATE DATABASE IF NOT EXISTS kctfest_test_repository; -- internal/repository
