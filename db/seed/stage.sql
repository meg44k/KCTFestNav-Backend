-- 既存の開発用DBにステージのモックだけを足す(流すたびに増えるので 1 回だけ)
-- docker compose exec -T mysql mysql --default-character-set=utf8mb4 -uroot kctfestnav < db/seed/stage.sql
SET NAMES utf8mb4;

USE kctfestnav;

-- ステージイベントのモック(時刻は UTC。コメントは日本時間)
-- 10/31: Live1 13:00-13:50 8組 / 癒し系ミュージシャン 14:00-14:40, 14:50-15:30 各3組
-- 11/1 : Live2 12:00-12:45 6組 / 弾き語り 13:30-13:50, 14:00-14:20, 14:30-14:50 各2組
INSERT INTO stage_sections (name, location, sort_order) VALUES ('Live1', '第一体育館', 1);
SET @section = LAST_INSERT_ID();
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-10-31 04:00:00', '2026-10-31 04:50:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'ROCKETS', 'ROCKETSの演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'ねこじゃらし', 'ねこじゃらしの演奏です。ぜひ聴きに来てください!', '', 2),
  (@block, 'Blue Monday', 'Blue Mondayの演奏です。ぜひ聴きに来てください!', '', 3),
  (@block, '電子工学科バンド', '電子工学科バンドの演奏です。ぜひ聴きに来てください!', '', 4),
  (@block, '夜更かし同好会', '夜更かし同好会の演奏です。ぜひ聴きに来てください!', '', 5),
  (@block, 'The Kosen', 'The Kosenの演奏です。ぜひ聴きに来てください!', '', 6),
  (@block, 'スリーコード', 'スリーコードの演奏です。ぜひ聴きに来てください!', '', 7),
  (@block, 'ラストオーダー', 'ラストオーダーの演奏です。ぜひ聴きに来てください!', '', 8);
INSERT INTO stage_sections (name, location, sort_order) VALUES ('癒し系ミュージシャン', '中庭', 2);
SET @section = LAST_INSERT_ID();
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-10-31 05:00:00', '2026-10-31 05:40:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'アコースティック部', 'アコースティック部の演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'ひだまり', 'ひだまりの演奏です。ぜひ聴きに来てください!', '', 2),
  (@block, '木漏れ日トリオ', '木漏れ日トリオの演奏です。ぜひ聴きに来てください!', '', 3);
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-10-31 05:50:00', '2026-10-31 06:30:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'ゆうなぎ', 'ゆうなぎの演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'ピアノとうた', 'ピアノとうたの演奏です。ぜひ聴きに来てください!', '', 2),
  (@block, 'そよかぜ', 'そよかぜの演奏です。ぜひ聴きに来てください!', '', 3);
INSERT INTO stage_sections (name, location, sort_order) VALUES ('Live2', '第一体育館', 3);
SET @section = LAST_INSERT_ID();
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-11-01 03:00:00', '2026-11-01 03:45:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'Midnight Owls', 'Midnight Owlsの演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, '情報工学科有志', '情報工学科有志の演奏です。ぜひ聴きに来てください!', '', 2),
  (@block, 'アンプ壊れた', 'アンプ壊れたの演奏です。ぜひ聴きに来てください!', '', 3),
  (@block, 'Sunday Morning', 'Sunday Morningの演奏です。ぜひ聴きに来てください!', '', 4),
  (@block, '五年生バンド', '五年生バンドの演奏です。ぜひ聴きに来てください!', '', 5),
  (@block, 'フィナーレ', 'フィナーレの演奏です。ぜひ聴きに来てください!', '', 6);
INSERT INTO stage_sections (name, location, sort_order) VALUES ('弾き語り', '中庭', 4);
SET @section = LAST_INSERT_ID();
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-11-01 04:30:00', '2026-11-01 04:50:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'ひとりぼっち', 'ひとりぼっちの演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'うたうたい', 'うたうたいの演奏です。ぜひ聴きに来てください!', '', 2);
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-11-01 05:00:00', '2026-11-01 05:20:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, 'ギターと僕', 'ギターと僕の演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'ゆめうつつ', 'ゆめうつつの演奏です。ぜひ聴きに来てください!', '', 2);
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (@section, '2026-11-01 05:30:00', '2026-11-01 05:50:00');
SET @block = LAST_INSERT_ID();
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES
  (@block, '夕焼け', '夕焼けの演奏です。ぜひ聴きに来てください!', '', 1),
  (@block, 'またね', 'またねの演奏です。ぜひ聴きに来てください!', '', 2);
