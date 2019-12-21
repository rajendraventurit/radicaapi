


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

