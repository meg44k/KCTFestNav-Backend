-- 開発用の初期データ。座標は北九州高専の敷地内に収まる値にしてある。
-- 混雑度は MySQL ではなく Redis (congestion_status:{id}) で管理しているため、
-- ここでは投入しない。未登録のブースは「空いている」として扱われる。
SET NAMES utf8mb4;

USE kctfestnav;

-- 位置と階はフロントの地図(src/data/campus.json)の棟に合わせてある。floor 0 = 屋外
INSERT INTO booths (name, organizer, detail, location, image_url, x, y, z, latitude, longitude, floor) VALUES
  ('たこ焼き',       '1-1',    'アツアツのたこ焼きを焼いています',       '中庭',       NULL, 0, 0, 1, 33.8163292, 130.87196, 0),
  ('焼きそば',       '1-2',    '大盛り無料の焼きそばです',               '中庭',       NULL, 1, 0, 1, 33.8163713, 130.8723101, 0),
  ('クレープ',       '2-1',    '甘いクレープを用意しています',           '2号館1階',     NULL, 2, 0, 1, 33.8163292, 130.8721222, 1),
  ('わたあめ',       '2-2',    'ふわふわのわたあめ',                     '1号館1階',     NULL, 3, 0, 1, 33.8164888, 130.8717485, 1),
  ('お化け屋敷',     '3-1',    '本気で怖いお化け屋敷です',               '3号館2階', NULL, 0, 1, 1, 33.8160767, 130.8723894, 2),
  ('脱出ゲーム',     '3-2',    '制限時間30分の脱出ゲーム',               '3号館3階', NULL, 1, 1, 1, 33.8162644, 130.8725332, 3),
  ('プラネタリウム', '天文部', '暗室に星空を投影します',                 '7号館3階 視聴覚室',   NULL, 0, 2, 2, 33.8159838, 130.871574, 3),
  ('VR体験',         'PC部',   '自作のVRコンテンツを体験できます',       '6号館2階 情報処理室', NULL, 1, 2, 2, 33.8161949, 130.8713045, 2);

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
