package handler

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"bpeecs.nchu.edu.tw/config"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Files handles manipulations about files
type Files struct {
	UploadTime uint64
	ClientName string `json:"client_name"`
	ServerName string `json:"server_name"`
	Mime       string
	Path       string `json:"path"`
}

// NewFile creates new file by FileHeader
func NewFile(fh *multipart.FileHeader) (*Files, error) {
	filePath := "./assets/upload/"
	fileExt := filepath.Ext(fh.Filename)

	// Generate new file name on server (do not use client name)
	fileName := randomString(10)
	for fileExists(filePath + fileName + fileExt) {
		fileName = randomString(10)
	}

	destFile, err := os.OpenFile(filePath+string(fileName)+fileExt, os.O_WRONLY|os.O_CREATE, 0666)
	defer destFile.Close()
	if err != nil {
		log.Println(err, "files.go NewFile() os.OpenFile() failed")
		return nil, err
	}

	srcFile, _ := fh.Open()
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Println(err, "files.go NewFile() io.Copy() failed")
		return nil, err
	}

	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer d.Close()
	stmt, _ := d.Prepare(`INSERT INTO files(upload_time, client_name, server_name, mime, path)
                          values(?, ?, ?, ?, ?)`)
	now := time.Now().Unix()

	f := Files{
		UploadTime: uint64(now),
		ClientName: fh.Filename,
		ServerName: fileName,
		Mime:       fh.Header.Get("Content-Type"),
		Path:       "/assets/upload/" + fileName + fileExt,
	}

	_, err = stmt.Exec(now, f.ClientName, f.ServerName, f.Mime, f.Path)
	if err != nil {
		log.Println(err, "files.go NewFile() stmt.Exec() failed")
		return nil, err
	}

	return &f, nil
}

// Del deletes a file by server_name
func (f *Files) Del() error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	rows := d.QueryRow("SELECT path FROM files WHERE server_name = ?", f.ServerName)

	var path string
	if err := rows.Scan(&path); err != nil {
		log.Println(err)
		return err
	}

	return DelFilesByPathList([]string{path})
}

// AutoCleanFiles deletes files do not be used anymore
// i.e., delete the file where it's aid is null
func AutoCleanFiles() {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return
	}
	defer d.Close()

	rows, err := d.Query(`
        SELECT path FROM files
        WHERE article_id is null and upload_time < ?`,
		time.Now().Unix()-12*60*60)
	defer rows.Close()

	if err != nil {
		log.Println(err, "files.go AutoCleanFiles() db.Query failed")
		return
	}

	path := ""
	pathList := []string{}
	for rows.Next() {
		rows.Scan(&path)
		pathList = append(pathList, path)
	}

	DelFilesByPathList(pathList)
}

func DelFilesByPathList(pathList []string) error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()

	for _, v := range pathList {
		err := os.Remove("." + v)
		if err != nil {
			log.Println(err, "files.go os.Remove()")
		}

		stmt, _ := d.Prepare("DELETE FROM files WHERE path=?")
		_, err = stmt.Exec(v)
		if err != nil {
			log.Println(err, "files.go DelFilesByPathList() stmt.Exec() failed")
			return err
		}
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
