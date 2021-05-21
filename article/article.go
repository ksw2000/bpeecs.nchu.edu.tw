package article

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bpeecs.nchu.edu.tw/config"
	"bpeecs.nchu.edu.tw/files"

	_ "github.com/mattn/go-sqlite3"
)

// Article handles manipulations about article
type Article struct {
	artList []Format
}

// Format records article's information
type Format struct {
	ID           uint32        `json:"id"`
	User         string        `json:"user"`
	Type         string        `json:"type"`
	CreateTime   uint64        `json:"create"`
	PublishTime  uint64        `json:"publish"`
	LastModified uint64        `json:"lastModified"`
	Title        string        `json:"title"`
	Content      string        `json:"content"`
	Attachment   []files.Files `json:"attachment"`
}

// New returns new instance of Article
func New() *Article {
	return new(Article)
}

// NewArticle is used to initialize an editor when user want to add new article
func (a *Article) NewArticle(user string) (uint32, error) {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer d.Close()
	stmt, err := d.Prepare("INSERT INTO article(user, create_time, publish_time, last_modified, title, content) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("d.Prepare() error %v", err)
	}

	res, err := stmt.Exec(user, time.Now().Unix(), 0, 0, "", "")
	if err != nil {
		return 0, fmt.Errorf("stmt.Exec() error %v", err)
	}
	aid, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("res.LastInsertId() error %v", err)
	}
	return uint32(aid), nil
}

// Save article (do not change scope)
func (a *Article) Save(artfmt Format, serverNameList []string) error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	stmt, _ := d.Prepare("UPDATE article SET title=?, type=?, content=?, last_modified=? WHERE id=? and user=?")
	now := time.Now().Unix()
	if _, err = stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, artfmt.ID, artfmt.User); err != nil {
		log.Println(err)
		return err
	}
	if err = a.UpdateAttachment(artfmt.ID, serverNameList); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// UpdateAttachment handles attachment information when users publishing or saving an article
func (a *Article) UpdateAttachment(aid uint32, serverNameList []string) error {
	for _, serverName := range serverNameList {
		err := a.LinkAttachment(serverName, aid)
		if err != nil {
			return err
		}
	}
	return nil
}

// LinkAttachment links temporary files information in article table
// We handle files by two steps
// 1. record information on file table
// 2. article table link the file information
func (a *Article) LinkAttachment(serverName string, aid uint32) error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	stmt, _ := d.Prepare("UPDATE files SET article_id = ? WHERE server_name = ?")
	_, err = stmt.Exec(aid, serverName)
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Publish an article (update content and change scope)
func (a *Article) Publish(artfmt Format, serverNameList []string) error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	stmt, _ := d.Prepare("UPDATE article SET title=?, type=?, content=?, publish_time=?, last_modified=?  WHERE id=? and user=?")
	now := time.Now().Unix()
	if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, now, artfmt.ID, artfmt.User); err != nil {
		log.Println(err)
		return err
	}
	if err := a.UpdateAttachment(artfmt.ID, serverNameList); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Del deletes an article
func (a *Article) Del(aid uint32, user string) error {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	stmt, _ := d.Prepare("DELETE from article WHERE id=? and user=?")
	if _, err := stmt.Exec(aid, user); err != nil {
		log.Println(err, "article.go Del() DELETE from article")
		return err
	}

	// remove attachment
	f := files.New()
	rows, _ := d.Query("SELECT path FROM files WHERE article_id=?", aid)
	path := ""
	pathList := []string{}
	for rows.Next() {
		rows.Scan(&path)
		pathList = append(pathList, path)
	}
	rows.Close()
	f.DelByPathList(pathList)

	// auto remove
	f.AutoDel()
	return nil
}

// GetLatest will get the lastest article
func (a *Article) GetLatest(scope string, artType string, user string, from int32, to int32) (list []Format, hasNext bool) {
	var queryString string

	switch scope {
	// All of article that the user have
	case "all":
		queryString = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE user = ?
        ORDER BY last_modified DESC, create_time DESC, publish_time DESC`

	// The user's article and these articles have not been published (still in draft box)
	case "draft":
		queryString = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time = 0 and user = ?
        ORDER BY last_modified DESC, create_time DESC`

	// The user's article and these articles have been published
	case "published":
		queryString = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0 and user = ?
        ORDER BY last_modified DESC, create_time DESC, publish_time DESC`

	// All of specifying type article that has been published in the database
	case "public-with-type":
		queryString = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0 and type = ?
        ORDER BY publish_time DESC`

	// All of article that has been published in the database
	case "public":
		queryString = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0
        ORDER BY publish_time DESC`

	default:
		return nil, false
	}

	// query more than one to decide [hasNext]
	queryString += fmt.Sprintf(" limit %d, %d", from, to-from+2)

	var rows *sql.Rows
	var err error
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return nil, false
	}
	defer d.Close()

	if scope == "all" || scope == "draft" || scope == "published" {
		rows, err = d.Query(queryString, user)
	} else if scope == "public-with-type" {
		if artType == "" {
			artType = "normal"
		}
		rows, err = d.Query(queryString, artType)
	} else {
		rows, err = d.Query(queryString)
	}
	if err != nil {
		log.Println(err)
		return nil, false
	}

	defer rows.Close()
	for i := int32(0); rows.Next(); i++ {
		var r Format
		rows.Scan(&r.ID, &r.User, &r.Type, &r.CreateTime, &r.PublishTime, &r.LastModified, &r.Title, &r.Content)

		// Load attachment list
		r.Attachment = a.getAttachmentByArticleID(r.ID)

		if i == to-from+1 {
			hasNext = true
		} else {
			list = append(list, r)
		}
	}

	return list, hasNext
}

// GetArticleByAid gets article information by aid
func (a *Article) GetArticleByAid(aid uint32, user string) *Format {
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer d.Close()
	row := d.QueryRow(`SELECT id, user, type, create_time, publish_time, last_modified, title, content
                       FROM article WHERE id = ?`, aid)

	r := new(Format)
	err = row.Scan(&r.ID, &r.User, &r.Type, &r.CreateTime, &r.PublishTime, &r.LastModified, &r.Title, &r.Content)

	if err == sql.ErrNoRows {
		return nil
	}

	// The news has not been published
	// Permission denied
	if r.PublishTime == 0 && r.User != user {
		return nil
	}

	// Load attachment list
	r.Attachment = a.getAttachmentByArticleID(r.ID)

	return r
}

// getAttachmentByArticleID gets attachment info and returns Files list
func (a *Article) getAttachmentByArticleID(id uint32) []files.Files {
	d, err := sql.Open("sqlite3", config.MainDB)
	fileList := []files.Files{}
	if err != nil {
		log.Println(err)
		return fileList
	}
	defer d.Close()
	rows, err := d.Query(`SELECT client_name, server_name,
                         IFNULL(mime,"") as mime, path
                         FROM files WHERE article_id=?`, id)
	/*
	   type Files struct{
	       UploadTime uint64
	       ClientName string
	       ServerName string
	       Mime string
	       Path string
	   }
	*/
	defer rows.Close()
	for rows.Next() {
		f := new(files.Files)
		rows.Scan(&f.ClientName, &f.ServerName, &f.Mime, &f.Path)
		fileList = append(fileList, *f)
	}
	return fileList
}
