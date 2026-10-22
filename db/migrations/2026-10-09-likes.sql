-- クラス展示への投票。1 つの投票者番号で 1 票(別の展示に入れ直すと移る)
CREATE TABLE likes(
  voter_id CHAR(36) NOT NULL PRIMARY KEY,
  booth_id INT NOT NULL,
  created_at DATETIME NOT NULL,
  INDEX likes_booth_time (booth_id, created_at),
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
