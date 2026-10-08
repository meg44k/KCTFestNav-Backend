-- ブースに階を足す(0 = 屋外/未設定)。マイグレーションツールは未導入のため手動で流す
-- docker compose exec -T mysql mysql --default-character-set=utf8mb4 -uroot kctfestnav < db/migrations/2026-10-07-booth-floor.sql
ALTER TABLE booths ADD COLUMN floor INT NOT NULL DEFAULT 0 AFTER longitude;
