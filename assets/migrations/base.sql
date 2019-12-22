


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


CREATE TABLE IF NOT EXISTS users_activity (
	id bigint unsigned NOT NULL AUTO_INCREMENT,
	device_id varchar(255),
	activity_type ENUM('open_app', 'dashboard','setting') NOT NULL,
	created_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted boolean NOT NULL DEFAULT FALSE,
	PRIMARY KEY(id)
) ENGINE=InnoDB CHARSET=utf8;





