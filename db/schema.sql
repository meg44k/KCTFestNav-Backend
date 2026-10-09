CREATE TABLE booths(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  organizer VARCHAR(255) NOT NULL,
  detail TEXT NOT NULL,
  location TEXT,
  image_url TEXT,
  x FLOAT NOT NULL,
  y FLOAT NOT NULL,
  z FLOAT NOT NULL,
  latitude DOUBLE,
  longitude DOUBLE,
  floor INT NOT NULL DEFAULT 0 /* 階。0 = 屋外(または未設定) */
);

CREATE TABLE users(
  id VARCHAR(36) PRIMARY KEY, /*UUID*/  
  login_id VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  assigned_booth_id INT,
  password TEXT NOT NULL,
  role VARCHAR(20) NOT NULL,
  FOREIGN KEY (assigned_booth_id) REFERENCES booths(id) ON DELETE SET NULL
);


-- ステージイベント。セクション(Live1 など) → ブロック(時間帯) → 出演者(順番だけ)
CREATE TABLE stage_sections(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  location VARCHAR(255) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0
);

-- current_order: 0 = まだ始まっていない、1〜出演者数 = その順番の出演者が演奏中、出演者数 + 1 = 終了
CREATE TABLE stage_blocks(
  id INT AUTO_INCREMENT PRIMARY KEY,
  section_id INT NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  current_order INT NOT NULL DEFAULT 0,
  FOREIGN KEY (section_id) REFERENCES stage_sections(id) ON DELETE CASCADE
);

CREATE TABLE performers(
  id INT AUTO_INCREMENT PRIMARY KEY,
  block_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  detail TEXT NOT NULL,
  thumbnail_url TEXT NOT NULL,
  perform_order INT NOT NULL,
  FOREIGN KEY (block_id) REFERENCES stage_blocks(id) ON DELETE CASCADE
);

-- クラス展示のいいね。1 つの投票者番号で同じブースは 1 件
CREATE TABLE likes(
  booth_id INT NOT NULL,
  voter_id CHAR(36) NOT NULL,
  created_at DATETIME NOT NULL,
  PRIMARY KEY (booth_id, voter_id),
  INDEX likes_booth_time (booth_id, created_at),
  INDEX likes_voter (voter_id),
  FOREIGN KEY (booth_id) REFERENCES booths(id) ON DELETE CASCADE
);

-- いいねを取り消した記録。booth_id が NULL は「全部消す」
CREATE TABLE like_removals(
  id INT AUTO_INCREMENT PRIMARY KEY,
  booth_id INT,
  from_at DATETIME,
  to_at DATETIME,
  removed INT NOT NULL,
  removed_by VARCHAR(36) NOT NULL,
  created_at DATETIME NOT NULL
);
