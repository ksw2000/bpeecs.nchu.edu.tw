# bpeecs.nchu.edu.tw

__NCHU BPEECS__ [https://bpeecs.nchu.edu.tw/](https://bpeecs.nchu.edu.tw/)

![](https://imgur.com/OUv4VWm.png)

## Quick start

1. Create database `./db/main.db`

2. Create a new account

```go
package main

import (
    "bpeecs.nchu.edu.tw/handler"
    "log"
)

func main() {
    if err := handler.NewAcount("user_id", "password", "userName"); err != nil{
        log.Println(err)
    }
}
```

3. run
```sh
# build
$ go build main.go

# run at port 9000
$ ./main

# run at port 443
$ ./main -p 443
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
    + assets/  (static files)
        + fonts/
        + img/
        + json/
        + js/
        + style/
        + upload/ (client upload files)
    + db/ (sqlite database)
        + main.db
    + handler/
        + basic.go (`/*`)
        + error.go (`/error/*` HTTP403 & 404)
        + api.go (`/api/*` for Ajax)
        + manage.go (`/manage/*`)
        + syllabus.go (`/syllabus/*`)
        + article.go
        + calendar.go
        + login.go
        + renderer.go
    + include/  (html files & gohtml layout files)
    + files/ (manage the uploaded files)
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