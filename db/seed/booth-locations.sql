-- 既存のDBのモックのブースを、地図の棟の中(と中庭)に置き直す。
-- 先に db/migrations/2026-10-07-booth-floor.sql を流しておくこと
-- docker compose exec -T mysql mysql --default-character-set=utf8mb4 -uroot kctfestnav < db/seed/booth-locations.sql
SET NAMES utf8mb4;

UPDATE booths SET latitude = 33.8163292, longitude = 130.87196, floor = 0, location = '中庭' WHERE name = 'たこ焼き';
UPDATE booths SET latitude = 33.8163713, longitude = 130.8723101, floor = 0, location = '中庭' WHERE name = '焼きそば';
UPDATE booths SET latitude = 33.8163292, longitude = 130.8721222, floor = 1, location = '2号館1階' WHERE name = 'クレープ';
UPDATE booths SET latitude = 33.8164888, longitude = 130.8717485, floor = 1, location = '1号館1階' WHERE name = 'わたあめ';
UPDATE booths SET latitude = 33.8160767, longitude = 130.8723894, floor = 2, location = '3号館2階' WHERE name = 'お化け屋敷';
UPDATE booths SET latitude = 33.8162644, longitude = 130.8725332, floor = 3, location = '3号館3階' WHERE name = '脱出ゲーム';
UPDATE booths SET latitude = 33.8159838, longitude = 130.871574, floor = 3, location = '7号館3階 視聴覚室' WHERE name = 'プラネタリウム';
UPDATE booths SET latitude = 33.8161949, longitude = 130.8713045, floor = 2, location = '6号館2階 情報処理室' WHERE name = 'VR体験';
