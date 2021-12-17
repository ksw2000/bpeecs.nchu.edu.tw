BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "article" (
	"id"	INTEGER,
	"user"	TEXT NOT NULL,
	"type"	TEXT NOT NULL DEFAULT 'normal',
	"create_time"	INTEGER,
	"publish_time"	INTEGER,
	"last_modified"	INTEGER,
	"title"	TEXT,
	"content"	TEXT,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "user" (
	"id"	TEXT,
	"password"	TEXT,
	"salt"	TEXT,
	"name"	TEXT,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "files" (
	"upload_time"	INTEGER,
	"client_name"	TEXT,
	"server_name"	TEXT,
	"mime"	TEXT,
	"path"	TEXT,
	"article_id"	INTEGER,
	PRIMARY KEY("server_name"),
	FOREIGN KEY("article_id") REFERENCES "article"("id")
);
COMMIT;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "user" (
	"id"	TEXT,
	"password"	TEXT,
	"salt"	TEXT,
	"name"	TEXT,
	PRIMARY KEY("id")
);
INSERT INTO "user" ("id","password","salt","name") VALUES ('root','24aaa75951fb0b7ec740e834cfadea6e374204e0539aaf7c39c6932d03d7ce0a','PMSLiYE@!LEMtshZVuD75k83QLC4E9goSermiLNnRdZ65MqMKcx!EPlStRKPv6lf','Root');
COMMIT;

