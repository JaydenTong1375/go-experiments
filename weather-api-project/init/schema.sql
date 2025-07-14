CREATE TABLE IF NOT EXISTS weather_days_record (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  day_id VARCHAR(255) NOT NULL UNIQUE,
  country VARCHAR(255) NOT NULL,
  timezone VARCHAR(255) NOT NULL,
  `date` DATE NOT NULL,
  conditions VARCHAR(255),
  `description` VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  key idx_day_id (day_id)
);

CREATE TABLE IF NOT EXISTS weather_hours_record (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  day_id VARCHAR(255) NOT NULL,
  `time` TIME,
  conditions VARCHAR(255),
  stations VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  KEY idx_day_id (day_id),
  CONSTRAINT FK_ FOREIGN key (day_id) REFERENCES weather_days_record(day_id)
);
