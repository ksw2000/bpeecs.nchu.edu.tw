package article
import(
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "time"
    "log"
    _"github.com/mattn/go-sqlite3"
    "bpeecs.nchu.edu.tw/files"
)

type Article struct{
    db *sql.DB
    artList []Article_Format
}

type Article_Format struct{
    Id uint32
    User string
    Type string
    Create_time uint64
    Publish_time uint64
    Last_modified uint64
    Title string
    Content string
    Attachment string
}

type Attachment_Format struct{
    Client_name []string `json:"client_name"`
    Path []string `json:"path"`
    Server_name []string `json:"server_name"`
}

func New() *Article{
    return new(Article)
}

func (a *Article) GetArtList() []Article_Format{
    return a.artList
}

func (a *Article) Connect(path string) *sql.DB{
    db, err := sql.Open("sqlite3", path)
    a.db = db
    if err != nil{
        log.Println(err)
        return nil
    }
    return a.db
}

func (a *Article) NewArticle(user string) (uint32, error){
    stmt, _ := a.db.Prepare("INSERT INTO article(id, user, create_time, publish_time, last_modified, title, content) values(?, ?, ?, ?, ?, ?, ?)")
    now := time.Now().Unix()
    serial_num := a.serialNumber(user);
    if _, err := stmt.Exec(serial_num, user, now, 0, 0, "", ""); err != nil{
        log.Println(err)
        return 0, err
    }

    return serial_num, nil
}

// Save article (do not change scope)
func (a *Article) Save(artfmt Article_Format) error{
    stmt, _ := a.db.Prepare("UPDATE article SET title=?, type=?, content=?, last_modified=? WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, artfmt.Id, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    if err := a.UpdateAttachment(artfmt.Id, artfmt.Attachment); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

func (a *Article) UpdateAttachment(article_id uint32, attachment string) error{
    // Parse JSON
    format := new(Attachment_Format)
    json.Unmarshal([]byte(attachment), format)
    for _, name := range format.Server_name{
        a.LinkAttachment(name, article_id)
    }

    return nil
}

func (a *Article) LinkAttachment(server_name string, article_id uint32) error{
    stmt, _ := a.db.Prepare("UPDATE files SET article_id=? WHERE server_name=?")
    _, err := stmt.Exec(article_id, server_name)
    defer stmt.Close()
    if err != nil{
        log.Println(err)
        return err
    }

    return nil
}

// Publish an article (update content and change scope)
func (a *Article) Publish(artfmt Article_Format) error{
    stmt, _ := a.db.Prepare("UPDATE article SET title=?, type=?, content=?, publish_time=?, last_modified=?  WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, now, artfmt.Id, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    if err := a.UpdateAttachment(artfmt.Id, artfmt.Attachment); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

// Delete an article
func (a *Article) Del(serial uint32, user string) error{
    stmt, _ := a.db.Prepare("DELETE from article WHERE id=? and user=?")
    if _, err := stmt.Exec(serial, user); err != nil{
        log.Println(err, "article.go Del() DELETE from article")
        return err
    }

    // remove attachment
    f := files.New()
    if db := f.Connect("./db/main.db"); db == nil{
        log.Println("error article.go Del()")
        return errors.New("Database connect error")
    }

    rows, _ := a.db.Query("SELECT path FROM files WHERE article_id=?", serial)
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

// Get the lastest article
func (a *Article) GetLatest(scope string, artType string, user string, from int32, to int32) (list []Article_Format, hasNext bool){
    var db_query_str string

    switch scope {
    // All of article that the user have
    case "all":
        db_query_str = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE user = ?
        ORDER BY last_modified DESC, create_time DESC, publish_time DESC`

    // The user's article and these articles have not been published (still in draft box)
    case "draft":
        db_query_str = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time = 0 and user = ?
        ORDER BY last_modified DESC, create_time DESC`

    // The user's article and these articles have been published
    case "published":
        db_query_str = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0 and user = ?
        ORDER BY last_modified DESC, create_time DESC, publish_time DESC`

    // All of specifying type article that has been published in the database
    case "public-with-type":
        db_query_str = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0 and type = ?
        ORDER BY publish_time DESC`

    // All of article that has been published in the database
    case "public":
        db_query_str = `
        SELECT id, user, type, create_time, publish_time, last_modified,
        title, content
        FROM article WHERE publish_time <> 0
        ORDER BY publish_time DESC`

    default:
        return nil, false
    }

    // query more than one to decide [hasNext]
    db_query_str += fmt.Sprintf(" limit %d, %d", from, to-from+2)

    var rows *sql.Rows
    var err error
    if scope == "all" || scope == "draft" || scope == "published" {
        rows, err = a.db.Query(db_query_str, user)
    }else if scope == "public-with-type"{
        if artType == ""{
            artType = "normal"
        }
        rows, err = a.db.Query(db_query_str, artType)
    }else{
        rows, err = a.db.Query(db_query_str)
    }
    if err != nil{
        log.Println(err)
        return nil, false
    }

    defer rows.Close()
    for i:= int32(0); rows.Next(); i++ {
        var r Article_Format
        rows.Scan(&r.Id, &r.User, &r.Type, &r.Create_time, &r.Publish_time, &r.Last_modified, &r.Title, &r.Content)

        // Load attachment list
        r.Attachment = a.getAttachmentByArticleID(r.Id)

        if i == to - from + 1 {
            hasNext = true;
        }else{
            list = append(list, r)
        }
    }

    a.artList = list

    return list, hasNext
}

func (a *Article) GetArticleBySerial(serial uint32, user string) *Article_Format{
    row := a.db.QueryRow("SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content` FROM article WHERE `id` = ?", serial)

    r := new(Article_Format)
    err := row.Scan(&r.Id, &r.User, &r.Type, &r.Create_time, &r.Publish_time, &r.Last_modified, &r.Title, &r.Content)

    if err == sql.ErrNoRows{
        return nil
    }

    // The news has not been published
    // Permission denied
    if r.Publish_time == 0 && r.User != user {
        return nil
    }

    // Load attachment list
    r.Attachment = a.getAttachmentByArticleID(r.Id)

    return r
}

func (a *Article) serialNumber(user string) uint32{
    rows := a.db.QueryRow("SELECT `user`, `id`, `title`, `content` FROM article ORDER BY `id` DESC")

    var num uint32
    var title, content, user_in_db string
    err := rows.Scan(&user_in_db, &num, &title, &content)
    if err == sql.ErrNoRows{
        return 1
    }

    /* 如果流水號最大的那個消息是空消息且該消息屬於你自己則刪除該消息，並回傳該消息序號 */
    if title == "" && content == "" && a.getAttachmentByArticleID(num) == "" && user_in_db == user{
        stmt, _ := a.db.Prepare("DELETE FROM article WHERE `id` = ?")
        stmt.Exec(num)
        return num
    }
    return num + 1
}

// return JSON
func (a *Article) getAttachmentByArticleID(id uint32) string{
    rows, err := a.db.Query("SELECT client_name, server_name, path FROM files WHERE article_id=?", id)
    client_name := ""
    server_name := ""
    path := ""
    f := new(Attachment_Format)
    defer rows.Close()
    for rows.Next(){
        rows.Scan(&client_name, &server_name, &path)
        f.Client_name = append(f.Client_name, client_name)
        f.Server_name = append(f.Server_name, server_name)
        f.Path = append(f.Path, path)
    }

    if len(f.Client_name) == 0{
        return ""
    }

    ret, err := json.Marshal(f)
    if err != nil{
        log.Println(err, "article.go getAttachmentByArticleID() json.Marshal()")
        return ""
    }
    return string(ret)
}
