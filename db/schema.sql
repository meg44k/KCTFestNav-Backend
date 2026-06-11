CREATE TABLE booths(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  organizer VARCHAR(255) NOT NULL,
  detail TEXT NOT NULL,
  congestion_status TINYINT NOT NULL DEFAULT 0,
  x FLOAT NOT NULL,
  y FLOAT NOT NULL,
  z FLOAT NOT NULL
);

CREATE TABLE users(
  id VARCHAR(36) PRIMARY KEY, /*UUID*/  
  name VARCHAR(255) NOT NULL,
  assigned_booth_id INT,
  password TEXT NOT NULL,
  role VARCHAR(20) NOT NULL,
  FOREIGN KEY (assigned_booth_id) REFERENCES booths(id)
);


CREATE TABLE lives(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  detail TEXT,
  thumbnailURL TEXT,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  session_number TINYINT,
  status TINYINT NOT NULL
);
