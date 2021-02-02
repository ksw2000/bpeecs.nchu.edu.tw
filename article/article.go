package article
import(
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
    "log"
    "bpeecs.nchu.edu.tw/files"
    "bpeecs.nchu.edu.tw/db"
)

// Article handles manipulations about article
type Article struct{
    artList []Format
}

// Format records article's information
type Format struct{
    ID uint32
    User string
    Type string
    CreateTime uint64
    PublishTime uint64
    LastModified uint64
    Title string
    Content string
    Attachment string
}

// AttachmentFormat records attachment's information
type AttachmentFormat struct{
    ClientName []string `json:"client_name"`
    Path []string `json:"path"`
    ServerName []string `json:"server_name"`
}

// New returns new instance of Article
func New() *Article{
    return new(Article)
}

// NewArticle is used to initialize an editor when user want to add new article
func (a *Article) NewArticle(user string) (uint32, error){
    d, err := db.Connect(db.Main)
    if err != nil {
        log.Println(err)
        return 0, err
    }
    defer d.Close()
    stmt, _ := d.Prepare("INSERT INTO article(id, user, create_time, publish_time, last_modified, title, content) values(?, ?, ?, ?, ?, ?, ?)")
    now := time.Now().Unix()
    serialNum, err := a.serialNumber(user);
    if err != nil {
        log.Println(err)
        return 0, err
    }
    if _, err = stmt.Exec(serialNum, user, now, 0, 0, "", ""); err != nil{
        log.Println(err)
        return 0, err
    }

    return serialNum, nil
}

// Save article (do not change scope)
func (a *Article) Save(artfmt Format) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    stmt, _ := d.Prepare("UPDATE article SET title=?, type=?, content=?, last_modified=? WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err = stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, artfmt.ID, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    if err = a.UpdateAttachment(artfmt.ID, artfmt.Attachment); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

// UpdateAttachment handles attachment information when users publishing or saving an article
func (a *Article) UpdateAttachment(aid uint32, attachment string) error{
    // Parse JSON
    format := new(AttachmentFormat)
    json.Unmarshal([]byte(attachment), format)
    for _, name := range format.ServerName{
        a.LinkAttachment(name, aid)
    }

    return nil
}

// LinkAttachment links temporary files information in article table
// We handle files by two steps
// 1. record information on file table
// 2. article table link the file information
func (a *Article) LinkAttachment(serverName string, aid uint32) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    stmt, _ := d.Prepare("UPDATE files SET article_id=? WHERE server_name=?")
    _, err = stmt.Exec(aid, serverName)
    defer stmt.Close()
    if err != nil{
        log.Println(err)
        return err
    }

    return nil
}

// Publish an article (update content and change scope)
func (a *Article) Publish(artfmt Format) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    stmt, _ := d.Prepare("UPDATE article SET title=?, type=?, content=?, publish_time=?, last_modified=?  WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, now, artfmt.ID, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    if err := a.UpdateAttachment(artfmt.ID, artfmt.Attachment); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

// Del deletes an article
func (a *Article) Del(serial uint32, user string) error{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    stmt, _ := d.Prepare("DELETE from article WHERE id=? and user=?")
    if _, err := stmt.Exec(serial, user); err != nil{
        log.Println(err, "article.go Del() DELETE from article")
        return err
    }

    // remove attachment
    f := files.New()
    rows, _ := d.Query("SELECT path FROM files WHERE article_id=?", serial)
    path := ""
    pathList := []string{}
    for rows.Next(){
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
func (a *Article) GetLatest(scope string, artType string, user string, from int32, to int32) (list []Format, hasNext bool){
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
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return nil, false
    }
    defer d.Close()

    if scope == "all" || scope == "draft" || scope == "published" {
        rows, err = d.Query(queryString, user)
    }else if scope == "public-with-type"{
        if artType == ""{
            artType = "normal"
        }
        rows, err = d.Query(queryString, artType)
    }else{
        rows, err = d.Query(queryString)
    }
    if err != nil{
        log.Println(err)
        return nil, false
    }

    defer rows.Close()
    for i:= int32(0); rows.Next(); i++ {
        var r Format
        rows.Scan(&r.ID, &r.User, &r.Type, &r.CreateTime, &r.PublishTime, &r.LastModified, &r.Title, &r.Content)

        // Load attachment list
        r.Attachment = a.getAttachmentByArticleID(r.ID)

        if i == to - from + 1 {
            hasNext = true;
        }else{
            list = append(list, r)
        }
    }

    return list, hasNext
}

// GetArticleBySerial gets article information by article serial
func (a *Article) GetArticleBySerial(serial uint32, user string) *Format{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return nil
    }
    defer d.Close()
    row := d.QueryRow("SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content` FROM article WHERE `id` = ?", serial)

    r := new(Format)
    err = row.Scan(&r.ID, &r.User, &r.Type, &r.CreateTime, &r.PublishTime, &r.LastModified, &r.Title, &r.Content)

    if err == sql.ErrNoRows{
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

func (a *Article) serialNumber(user string) (uint32, error){
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return 0, err
    }
    defer d.Close()
    rows := d.QueryRow("SELECT `user`, `id`, `title`, `content` FROM article ORDER BY `id` DESC")

    var num uint32
    var title, content, dbUser string
    err = rows.Scan(&dbUser, &num, &title, &content)
    if err == sql.ErrNoRows{
        return 1, nil
    }

    /* 如果流水號最大的那個消息是空消息且該消息屬於你自己則刪除該消息，並回傳該消息序號 */
    if title == "" && content == "" && a.getAttachmentByArticleID(num) == "" && dbUser == user{
        stmt, _ := d.Prepare("DELETE FROM article WHERE `id` = ?")
        stmt.Exec(num)
        return num, nil
    }
    return num + 1, nil
}

// getAttachmentByArticleID gets attachment info and returns JSON
func (a *Article) getAttachmentByArticleID(id uint32) string{
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return ""
    }
    defer d.Close()
    rows, err := d.Query("SELECT client_name, server_name, path FROM files WHERE article_id=?", id)
    clientName := ""
    serverName := ""
    path := ""
    format := new(AttachmentFormat)
    defer rows.Close()
    for rows.Next(){
        rows.Scan(&clientName, &serverName, &path)
        format.ClientName = append(format.ClientName, clientName)
        format.ServerName = append(format.ServerName, serverName)
        format.Path = append(format.Path, path)
    }

    if len(format.ClientName) == 0{
        return ""
    }

    ret, err := json.Marshal(format)
    if err != nil{
        log.Println(err, "article.go getAttachmentByArticleID() json.Marshal()")
        return ""
    }
    return string(ret)
}
