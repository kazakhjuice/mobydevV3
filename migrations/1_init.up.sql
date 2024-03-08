CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    isAdmin BOOL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS films (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	category TEXT,
	project_type TEXT,
	year TEXT,
	duration INTEGER,
	tags TEXT,
	description TEXT,
	director TEXT,
	producer TEXT
);