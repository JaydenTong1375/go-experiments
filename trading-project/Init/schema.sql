CREATE TABLE IF NOT EXISTS users (
  user_id VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  hashed_password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS users_inventory_transactions (
  transactions_ID VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  item_id VARCHAR(255) NOT NULL,
  value INT unsigned NOT NULL,
  status ENUM('CREDIT', 'DEBIT') NOT NULL,
  reason TEXT NOT NULL DEFAULT "",
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_id (user_id),
  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(user_id)
);

