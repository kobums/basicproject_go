-- basicproject_go.sql
-- basicproject_springboot 와 동일한 스키마(user_tb / board_tb).
-- 모델 매니저가 사용하는 컬럼명(u_*, b_*)에 맞춘다.

CREATE TABLE IF NOT EXISTS `user_tb` (
  `u_id` bigint NOT NULL AUTO_INCREMENT,
  `u_email` varchar(255) NOT NULL,
  `u_name` varchar(100) DEFAULT NULL,
  `u_passwd` varchar(255) NOT NULL,
  `u_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`u_id`),
  UNIQUE KEY `uk_user_email` (`u_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `board_tb` (
  `b_id` bigint NOT NULL AUTO_INCREMENT,
  `b_title` varchar(200) NOT NULL,
  `b_content` text,
  `b_img` varchar(255) NOT NULL DEFAULT '',
  `b_user` bigint NOT NULL,
  `b_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`b_id`),
  KEY `idx_board_user` (`b_user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
