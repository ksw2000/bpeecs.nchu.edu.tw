# bpeecs.nchu.edu.tw

__NCHU BPEECS__ [https://bpeecs.nchu.edu.tw/](https://bpeecs.nchu.edu.tw/)

## Dependencies

__Go__

Powered by Golang 1.12 (need go mod)

![](https://golang.org/doc/gopher/pkg.png)

__SQLite3__

![](https://www.sqlite.org/images/sqlite370_banner.gif)

1. Install

    > Today, almost all the flavours of Linux OS are being shipped with SQLite. So you just issue the following command to check if you already have SQLite installed on your machine.

2. Create database
```sh
$ sqlite3 main.db
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

    + db/ (handle database)
        + db.go
        + main.db

    + include/  (html files & gohtml files)

    + article/ (handle article/(news) add, update, delete)
    + files/ (manage the uploaded files)

    + function/ (some func we usually use)

    + login/ (handle login)
    + render/
        + dynamic.go (render some pages when requesting)
        + static.go (render some pages before requesting)
    + web/
        + basic.go (handle: ./xxx)
        + error.go (hanle HTTP403 & 404)
        + function.go (handle ./function/xxx for Ajax)

    + go.mod

    + go.sum

    + __main.go__ (main program)

    + newAccount.go `private` (regist a new user)

## Database
__main.db__

```sql
CREATE TABLE "article" (
	"id"	INTEGER,
	"user"	TEXT,
	"type"	TEXT DEFAULT 'normal',
	"create_time"	INTEGER,
	"publish_time"	INTEGER,
	"last_modified"	INTEGER,
	"title"	TEXT,
	"content"	TEXT,
	PRIMARY KEY("id")
);

CREATE TABLE "files" (
	"upload_time"	INTEGER,
	"client_name"	TEXT,
	"server_name"	TEXT,
    "mime"  TEXT,
	"path"	TEXT,
	"article_id"	INTEGER,
	FOREIGN KEY("article_id") REFERENCES "article"("id"),
	PRIMARY KEY("server_name")
);

CREATE TABLE "user" (
	"id"	TEXT,
	"password"	TEXT,
	"salt"	TEXT,
	"name"	TEXT,
	PRIMARY KEY("id")
);
```

## Quick run

1. Create database `./db/main.db`

2. implemented pwdHash() in `package login`
```go
func pwdHash(pwd string, salt string)
```

3. use `newAccount.go` to create a new account

4. go build
```sh
$ go build main.go
```

5. run
```sh
# run at port 9000
./main
# run at port 8080 and render static page
./main -r -p 8080
# run at port 443
./main -r -p 443
```
`-p` can specify port.
`-r` can render static page
