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
$ go get github.com/mattn/go-sqlite3
```

__SQlite3__

1. Install

    > Today, almost all the flavours of Linux OS are being shipped with SQLite. So you just issue the following command to check if you already have SQLite installed on your machine.

2. Create tables
```sh
$ sqlite3 tableName
```

## Files

+ .git/

+ assests/  (static file)
    + fonts/
    + img/
    + js/
    + style/
    + upload/ (client upload files)

+ include/  (html files)

+ pkg/ (golang package)
    + linux_amd64
        + github.com
            + go-session (for session)
            + mattn (for sqlite3)

+ sql/ (store database)

+ src/ (golang source code)
    + article/ (process article/(news) add, update, delte)
    + files/ (manage the file which clients uploaded)
    + function/ (some func that golang often use)
    + github.com/
        + go-session (for session)
        + mattn (for sqlite3)
    + login/ (process login)
    + web/
        + basic.go (process: ./xxx)
        + error.go (process error url)
        + function.go (process ./function/xxx)
    + __index.go__ (main program)
    + newAccount.go `private` (regist a new user)


## Database

### article.db
```
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
```
CREATE TABLE files (
    id INTEGER PRIMARY KEY UNIQUE,
    client_name TEXT,
    server_name TEXT,
    path TEXT,
    hash TEXT
);
```

### user.db
```
CREATE TABLE user (
    num INTEGER PRIMARY KEY UNIQUE,
    id TEXT UNIQUE,
    password TEXT,
    salt TEXT,
    name TEXT
);
```

## Quick run

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
