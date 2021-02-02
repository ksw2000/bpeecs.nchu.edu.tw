package files

import(
    "io"
    "mime/multipart"
    "os"
    "time"
    "log"
    "path/filepath"
    "bpeecs.nchu.edu.tw/function"
    "bpeecs.nchu.edu.tw/db"
)

// Files handles manipulations about files
type Files struct{
    UploadTime uint64
    ClientName string
    ServerName string
    Path string
    Hash string
}

// New returns new instance of Files
func New() *Files{
    return new(Files)
}

// NewFile creates new file by FileHeader
func (f *Files) NewFile(fh *multipart.FileHeader) error{
    filePath := "./assets/upload/"
    fileExt  := filepath.Ext(fh.Filename)

    // Generate new file name on server (do not use client name)
    fileName := function.RandomString(10)
    for fileExists(filePath + fileName + fileExt){
        fileName = function.RandomString(10)
    }

    newFile, err := os.OpenFile(filePath + string(fileName) + fileExt, os.O_WRONLY | os.O_CREATE, 0666)
    defer newFile.Close()
    if err != nil{
        log.Println(err, "files.go NewFile() os.OpenFile() failed")
        return err
    }

    oriFile, _ := fh.Open()
    defer oriFile.Close()

    _, err = io.Copy(newFile, oriFile)
    if err != nil{
        log.Println(err, "files.go NewFile() io.Copy() failed")
        return err
    }

    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    stmt, _ := d.Prepare("INSERT INTO files(upload_time, client_name, server_name, path) values(?, ?, ?, ?)")
    now := time.Now().Unix()

    f.UploadTime = uint64(now)
    f.ClientName = fh.Filename
    f.ServerName = fileName
    f.Path = "/assets/upload/" + fileName + fileExt

    _, err = stmt.Exec(now, f.ClientName, f.ServerName, f.Path)
    if err != nil{
        log.Println(err, "files.go NewFile() stmt.Exec() failed")
        return err
    }

    return nil
}

// Del deletes a file by server_name
func (f *Files) Del(server_name string) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    rows := d.QueryRow("SELECT path FROM files WHERE server_name = ?", server_name)

    var path string
    if err := rows.Scan(&path); err != nil{
        log.Println(err)
        return err
    }

    return f.DelByPathList([]string{path})
}

// DelByPathList deletes files by path list
func (f *Files) DelByPathList(pathList []string) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()

    for _, v := range pathList{
        err := os.Remove("." + v)
        if err != nil{
            log.Println(err, "files.go os.Remove()")
        }

        stmt, _ := d.Prepare("DELETE FROM files WHERE path=?")
        _, err = stmt.Exec(v)
        if err != nil{
            log.Println(err, "files.go Remove() stmt.Exec() failed")
            return err
        }
    }
    return nil
}

// AutoDel deletes files do not be used anymore
func (f *Files) AutoDel(){
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return
    }
    defer d.Close()

    rows, err := d.Query(`
        SELECT path FROM files
        WHERE article_id is null and upload_time < ?` ,
        time.Now().Unix() - 12*60*60)
    defer rows.Close()

    if err != nil{
        log.Println(err, "files.go Automremove() db.Query failed")
        return
    }

    path := ""
    pathList := []string{}
    for rows.Next(){
        rows.Scan(&path)
        pathList = append(pathList, path)
    }

    f.DelByPathList(pathList)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
