# bpeecs.nchu.edu.tw

> A website for Banchelor Program of Electrical Engineering and Computer Science

## Dependencies

__GoLang__

Powered by Golang 1.12 (need go mod)

__SQlite3__

1. Install

    > Today, almost all the flavours of Linux OS are being shipped with SQLite. So you just issue the following command to check if you already have SQLite installed on your machine.

2. Create tables
```sh
$ sqlite3 tableName
```

__Front-end js dependencies__

All of the javascript dependencies are embedded by the online resource links. If these links are lost, replace a new one, or implemented it by yourself.

1. jQuery (v3.5.1)

2. Text editor: [CkEditor](https://ckeditor.com/)

3. Date foramt (jQuery dependency): [jquery-dateFormat](https://github.com/phstc/jquery-dateFormat)

4. Promise() for ES5: [ES6-promise](https://github.com/stefanpenner/es6-promise)

### IE

>
> 1. Transfer ES6 to ES5 at [Babel](https://babeljs.io/)
>
> 2. ES5 promise() support [ES6-promise](https://github.com/stefanpenner/es6-promise)
>

## Files
+ beepcs.nchu.edu.tw/
    + .git/

    + assests/  (static file)
        + fonts/
        + img/
        + json/
        + js/
        + style/
        + upload/ (client upload files)

    + include/  (html files)

    + sql/ (store database)

    + article/ (process article/(news) add, update, delte)
    + files/ (manage the file which clients uploaded)

    + function/ (some func that golang often use)

    + login/ (process login)
    + render/
        + dynamic.go (render in runtime)
        + static.go (render in compile-time)
    + web/
        + basic.go (process: ./xxx)
        + error.go (process error url)
        + function.go (process ./function/xxx)

    + go.mod

    + go.sum

    + __main.go__ (main program)

    + newAccount.go `private` (regist a new user)

## Database

### article.db
```sql
CREATE TABLE "article" (
	"id"	INTEGER UNIQUE,
	"user"	TEXT,
	"type"	TEXT DEFAULT 'normal',
	"create_time"	INTEGER,
	"publish_time"	INTEGER,
	"last_modified"	INTEGER,
	"title"	TEXT,
	"content"	INTEGER,
	"attachment"	INTEGER,
	PRIMARY KEY("id")
);
```

### files.db
```sql
CREATE TABLE "files" (
	"id"	INTEGER NOT NULL UNIQUE,
	"upload_time"	INTEGER,
	"client_name"	TEXT,
	"server_name"	TEXT,
	"path"	TEXT,
	PRIMARY KEY("id")
);
```

### user.db
```sql
CREATE TABLE "user" (
	"num"	INTEGER UNIQUE,
	"id"	TEXT UNIQUE,
	"password"	TEXT,
	"salt"	TEXT,
	"name"	TEXT,
	PRIMARY KEY("num")
);
```

## Quick run

1. create database

2. implemented pwdHash() in `package login`
```go
func pwdHash(pwd string, salt string)
```

3. use `newAccount.go` to create a new account

4. go build
```sh
$ go build main.go
```
5. run in port 8080
```sh
./main -r -p 8080
# or https (443)
./main -r -p 443
```
`-p` can specify port.
`-r` can render static page
