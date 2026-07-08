CREATE TABLE booths(
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  organizer VARCHAR(255) NOT NULL,
  detail TEXT NOT NULL,
  location TEXT,
  x FLOAT NOT NULL,
  y FLOAT NOT NULL,
  z FLOAT NOT NULL,
  latitude DOUBLE,
  longitude DOUBLE
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
