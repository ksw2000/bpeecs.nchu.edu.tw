# bpeecs.nchu.edu.tw

__NCHU BPEECS__ [https://bpeecs.nchu.edu.tw/](https://bpeecs.nchu.edu.tw/)

![](https://imgur.com/OUv4VWm.png)


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
    + config/
    + db/ (sqlite database)
        + calendar.db
        + calendar.db.sql (*only schema*)
        + main.db
        + main.db.sql (*only schema*)
    + handler/
        + basic.go (`/*`)
        + error.go (`/error/*` HTTP403 & 404)
        + api.go (`/api/*` for Ajax)
        + manage.go (`/manage/*`)
        + syllabus.go (`/syllabus/*`)
        + article.go
        + calendar.go
        + files.go
        + login.go
        + renderer.go
    + include/  (html files & gohtml layout files)
    + files/ (manage the uploaded files)
    + go.mod
    + go.sum
    + __main.go__ (main program)

## Quick start

1. Create database from /db/*.sql
2. Create a new account

    ```go
    handler.NewAcount("user_id", "password", "userName")
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