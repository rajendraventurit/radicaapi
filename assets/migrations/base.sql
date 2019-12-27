


CREATE TABLE IF NOT EXISTS users (
	user_id bigint unsigned NOT NULL AUTO_INCREMENT,
	first_name varchar(255),
	last_name varchar(255),
	email varchar(255) NOT NULL,
	password varchar(255),
	created_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted boolean NOT NULL DEFAULT FALSE,
	UNIQUE(email),
	PRIMARY KEY(user_id)
) ENGINE=InnoDB CHARSET=utf8;


CREATE TABLE IF NOT EXISTS `users_activity` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `device_id` varchar(255) DEFAULT NULL,
  `activity_type` enum('open_app','dashboard','setting') NOT NULL,
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8



CREATE TABLE IF NOT EXISTS `disease_by_radiation` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `disease` varchar(255) DEFAULT NULL,
  `symtoms` varchar(255) DEFAULT NULL,
  `disease_date` timestamp NOT NULL,
  `dbm`  integer,
  `onscreen_time` integer,
  `user_id` bigint unsigned NOT NULL,
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
   CONSTRAINT disease_by_radiation_fk1 FOREIGN KEY (user_id)
		REFERENCES users (user_id) ON DELETE CASCADE,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8






