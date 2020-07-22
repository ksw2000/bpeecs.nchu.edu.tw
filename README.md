# bpeecs.nchu.edu.tw

> A website for Banchelor Program of Electrical Engineering and Computer Science

## Dependencies

__GOlang__
1. Install go-session
```sh
$ go get -v github.com/go-session/session
```

2. Install go-sqlite3
```sh
$ go get -v github.com/mattn/go-sqlite3
```

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

2. Text editor: [simplemde](https://simplemde.com/)

3. Date foramte (jQuery dependency): [jquery-dateFormat](https://github.com/phstc/jquery-dateFormat)

## Files
+ bin/ (golang bin)

+ pkg/ (golang package)

+ src/ (golang source code)

    + beepcs.nchu.edu.tw/
        + .git/

        + assests/  (static file)
            + fonts/
            + img/
            + js/
            + style/
            + upload/ (client upload files)

        + include/  (html files)

        + sql/ (store database)

        + article/ (process article/(news) add, update, delte)
        + files/ (manage the file which clients uploaded)

        + function/ (some func that golang often use)

        + login/ (process login)

        + web/
            + basic.go (process: ./xxx)
            + error.go (process error url)
            + function.go (process ./function/xxx)

        + __index.go__ (main program)

        + newAccount.go `private` (regist a new user)

    + github.com/
        + go-session (for session)
        + mattn (for sqlite3)

## Database

### article.db
```sql
CREATE TABLE article (
    id PRIMARY KEY UNIQUE,
    user,
    create_time,
    publish_time,
    last_modified,
    title,
    content,
    attachment
);
```

### files.db
```sql
CREATE TABLE files (
    id INTEGER PRIMARY KEY UNIQUE,
    client_name TEXT,
    server_name TEXT,
    path TEXT,
    hash TEXT
);
```

### user.db
```sql
CREATE TABLE user (
    num INTEGER PRIMARY KEY UNIQUE,
    id TEXT UNIQUE,
    password TEXT,
    salt TEXT,
    name TEXT
);
```

## Quick run

1. create database

2. implemented pwdHash() in `package login`
```go
func pwdHash(pwd string, salt string)
```

3. use `newAccount.go` to create a new account

3. go run

```sh
$ cd ./src
$ go run index.go
```

## Build

```sh
$ cd ./src
$ go build index.go
$ ./index
```
