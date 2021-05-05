# bpeecs.nchu.edu.tw

__NCHU BPEECS__ [https://bpeecs.nchu.edu.tw/](https://bpeecs.nchu.edu.tw/)

![](https://imgur.com/OUv4VWm.png)

## Quick start

1. Create database `./db/main.db`

2. Create a new account

```go
package main

import "bpeecs.nchu.edu.tw/login"

func main() {
	l := new(login.Login)
	l.Connect("./db/main.db")
	l.NewAcount("user_id", "password", "userName")
}
```

3. run
```sh
# build
$ go build main.go

# run at port 9000
$ ./main

# run at port 8080 and render static page
$ ./main -r -p 8080

# run at port 443
$ ./main -r -p 443

# -p for specifing port. (default: 9000)
# -r for rendering static pages
```

## Dependencies

+ Go 1.12
+ SQLite3
    ```sh
    $ sqlite3 main.db
    ```
+ Front-end javascript

    + jQuery (v3.5.1)

    + Text editor: [CkEditor](https://ckeditor.com/)

    + Date format (jQuery dependency): [jquery-dateFormat](https://github.com/phstc/jquery-dateFormat)

## Files
+ beepcs.nchu.edu.tw/
    + .git/
    + assets/  (static file)
        + fonts/
        + img/
        + json/
        + js/
        + style/
        + upload/ (client upload files)
    + db/ (sqlite database)
        + main.db
    + handler/
        + basic.go (`./*`)
        + error.go (`/error/*` HTTP403 & 404)
        + function.go (`./function/*xxx*` for Ajax)
        + manage.go (h`/manage/*`)
        + syllabus.go (`/syllabus/*`)
    + include/  (html files & gohtml layout files)
    + article/ (handle `article/(news)` add, update, delete)
    + files/ (manage the uploaded files)
    + login/ (handle login)
    + render/
        + dynamic.go (render some pages when requesting)
        + static.go (render some pages before requesting)
    + go.mod
    + go.sum
    + __main.go__ (main program)

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



