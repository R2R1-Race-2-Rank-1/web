CREATE TABLE sets(
	id INT NOT NULL,
	ranked_status TINYINT NOT NULL,
	approved_date DATETIME NOT NULL,
	last_update DATETIME NOT NULL,
	last_checked DATETIME NOT NULL,
	artist VARCHAR(1000) NOT NULL,
	title VARCHAR(1000) NOT NULL,
	creator VARCHAR(1000) NOT NULL,
	source VARCHAR(1000) NOT NULL,
	tags VARCHAR(1000) NOT NULL,
	has_video TINYINT NOT NULL,
	genre TINYINT NOT NULL,
	language TINYINT NOT NULL,
	favourites INT NOT NULL,
	set_modes TINYINT NOT NULL,
	PRIMARY KEY(id)
);
