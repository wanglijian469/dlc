CREATE DATABASE IF NOT EXISTS dl_nongji_parts
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'dlc_app'@'%' IDENTIFIED BY 'change_me_please';
GRANT ALL PRIVILEGES ON dl_nongji_parts.* TO 'dlc_app'@'%';
FLUSH PRIVILEGES;
