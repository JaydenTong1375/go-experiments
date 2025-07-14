CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(255) UNIQUE,
  username VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transaction_users_inventory (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  item_id VARCHAR(255) DEFAULT "",
  quantity INT UNSIGNED NOT NULL DEFAULT 0,
  transaction_type enum('CREDIT','DEBIT','WIPE') NOT NULL,
  reason TEXT,
  FOREIGN KEY (user_id) REFERENCES users(user_id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transaction_users_credit (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  user_id varchar(100) NOT NULL,
  transaction_id varchar(100) NOT NULL UNIQUE,
  amount INT unsigned NOT NULL DEFAULT 0,
  transaction_type enum('CREDIT','DEBIT') NOT NULL,
  reason TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  KEY transaction_user_id (user_id),
  CONSTRAINT fk_transaction_credit_user FOREIGN KEY (user_id) REFERENCES users (user_id)
);

CREATE TABLE IF NOT EXISTS transaction_gacha_result (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(100) NOT NULL,
  item_id VARCHAR(100) NOT NULL,
  transaction_id VARCHAR(100) NOT NULL,
  quantity INT UNSIGNED NOT NULL DEFAULT 0,
  rarity VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,

  -- Indexes
  KEY transaction_user_id (user_id),
  KEY transaction_id_idx (transaction_id),

  -- Foreign Keys
  CONSTRAINT fk_transaction_user FOREIGN KEY (user_id) REFERENCES users(user_id),
  CONSTRAINT fk_transaction_credit FOREIGN KEY (transaction_id) REFERENCES transaction_users_credit(transaction_id)
);

-- Tracks user progress and metadata on each mission.
CREATE TABLE IF NOT EXISTS users_mission_tracker (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  user_id varchar(100) NOT NULL,
  mission_id varchar(100) NOT NULL UNIQUE,
  mission_name varchar(100) NOT NULL,
  mission_type enum('DAILY', 'WEEKLY') NOT NULL,
  mission_description TEXT DEFAULT "",
  mission_progress INT unsigned DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tracks individual objectives within a mission.
CREATE TABLE IF NOT EXISTS mission_objectives_tracker (
  id INT unsigned AUTO_INCREMENT PRIMARY KEY,
  mission_id varchar(100) NOT NULL,
  objective_id varchar(100) NOT NULL,
  objective_order INT unsigned DEFAULT 0, 
  objective_name varchar(100) NOT NULL,
  objective_description TEXT DEFAULT "",
  objective_progress INT unsigned DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_objectives_tracker FOREIGN KEY (mission_id) REFERENCES users_mission_tracker(mission_id)
);