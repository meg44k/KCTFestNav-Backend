-- 既存のDBをステージイベントの形にする(マイグレーションツールは未導入のため手動で流す)
-- docker compose exec -T mysql mysql --default-character-set=utf8mb4 -uroot kctfestnav < db/migrations/2026-10-05-stage.sql
SET NAMES utf8mb4;

DROP TABLE IF EXISTS lives;

-- ステージイベント。セクション(Live1 など) → ブロック(時間帯) → 出演者(順番だけ)
CREATE TABLE IF NOT EXISTS stage_sections(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  location VARCHAR(255) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0
);

-- current_order: 0 = まだ始まっていない、1〜出演者数 = その順番の出演者が演奏中、出演者数 + 1 = 終了
CREATE TABLE IF NOT EXISTS stage_blocks(
  id INT AUTO_INCREMENT PRIMARY KEY,
  section_id INT NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  current_order INT NOT NULL DEFAULT 0,
  FOREIGN KEY (section_id) REFERENCES stage_sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS performers(
  id INT AUTO_INCREMENT PRIMARY KEY,
  block_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  detail TEXT NOT NULL,
  thumbnail_url TEXT NOT NULL,
  perform_order INT NOT NULL,
  FOREIGN KEY (block_id) REFERENCES stage_blocks(id) ON DELETE CASCADE
);
