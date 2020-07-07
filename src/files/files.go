package files

import(
    "database/sql"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "strings"
    "time"
    "path/filepath"
    _"github.com/mattn/go-sqlite3"
    //--------------
    "function"
)

type Files struct{
    db *sql.DB
    Upload_time uint64
    Client_name string
    Server_name string
    Path string
    Hash string
    err error
}

type IFiles interface{
    Connect(path string) *sql.DB
    NewFile(fh *multipart.FileHeader) *Files
}

func (f *Files) errProcess(err error){
    if err!=nil{
        fmt.Println(err)
        f.err = err
        return
    }
}

func (f *Files) GetErr() error{
    return f.err
}

func (f *Files) Connect(path string) *sql.DB{
    db, err := sql.Open("sqlite3", path)
    f.db = db
    f.errProcess(err)
    return f.db
}

func (f *Files) NewFile(fh *multipart.FileHeader) *Files{
    filePath := "../assets/upload/"
    fileExt  := filepath.Ext(fh.Filename)
    fileName := strings.TrimRight(fh.Filename, fileExt)
    for fileExists(filePath + fileName + fileExt){
        fileName = function.RandomString(10)
    }

    newFile, err := os.OpenFile(filePath + string(fileName) + fileExt, os.O_WRONLY | os.O_CREATE, 0666)
    f.errProcess(err);

    oriFile, _ := fh.Open()
    defer oriFile.Close()

    _, err = io.Copy(newFile, oriFile)
    f.errProcess(err);

    stmt, err := f.db.Prepare("INSERT INTO files(upload_time, client_name, server_name, path) values(?, ?, ?, ?)")
    f.errProcess(err)
    now := time.Now().Unix()

    f.Upload_time = uint64(now)
    f.Client_name = fh.Filename
    f.Server_name = fileName
    f.Path = "/assets/upload/" + fileName + fileExt

    stmt.Exec(now, f.Client_name, f.Server_name, f.Path)

    return f
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
