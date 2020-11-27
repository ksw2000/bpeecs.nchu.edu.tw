package article
import(
    "database/sql"
    "fmt"
    "time"
    "log"
    _"github.com/mattn/go-sqlite3"
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
    stmt, _ := a.db.Prepare("INSERT INTO article(id, user, create_time, publish_time, last_modified, title, content, attachment) values(?, ?, ?, ?, ?, ?, ?, ?)")
    now := time.Now().Unix()
    serial_num := a.serialNumber(user);
    if _, err := stmt.Exec(serial_num, user, now, 0, 0, "", "", ""); err != nil{
        log.Println(err)
        return 0, err
    }

    return serial_num, nil
}

// Save article (do not change scope)
func (a *Article) Save(artfmt Article_Format) error{
    stmt, _ := a.db.Prepare("UPDATE article SET title=?, type=?, content=?, last_modified=?, attachment=?  WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, artfmt.Attachment, artfmt.Id, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

func (a *Article) UpdateAttachment(id uint32, user string, attachment string) error{
    stmt, _ := a.db.Prepare("UPDATE article SET last_modified=?, attachment=?  WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(now, attachment, id, user); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

// Publish an article (update content and change scope)
func (a *Article) Publish(artfmt Article_Format) error{
    stmt, _ := a.db.Prepare("UPDATE article SET title=?, type=?, content=?, publish_time=?, last_modified=? , attachment=?  WHERE id=? and user=?")
    now := time.Now().Unix()
    if _, err := stmt.Exec(artfmt.Title, artfmt.Type, artfmt.Content, now, now, artfmt.Attachment, artfmt.Id, artfmt.User); err != nil{
        log.Println(err)
        return err
    }
    return nil
}

// Delete an article
func (a *Article) Del(serial uint32, user string) error{
    stmt, _ := a.db.Prepare("DELETE from article WHERE id=? and user=?")
    if _, err := stmt.Exec(serial, user); err != nil{
        return err
    }
    return nil
}

// Get the lastest article
func (a *Article) GetLatest(scope string, artType string, user string, from int32, to int32) (list []Article_Format, hasNext bool){
    var db_query_str string

    switch scope {
    // All of article that the user have
    case "all":
        db_query_str =  "SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` "
        db_query_str += "FROM article WHERE user = ? "
        db_query_str += "ORDER BY `last_modified` DESC, `create_time` DESC, `publish_time` DESC"

    // The user's article and these articles have not been published (still in draft box)
    case "draft":
        db_query_str =  "SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` "
        db_query_str += "FROM article WHERE publish_time = 0 and user = ? "
        db_query_str += "ORDER BY `last_modified` DESC, `create_time` DESC"

    // The user's article and these articles have been published
    case "published":
        db_query_str =  "SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` "
        db_query_str += "FROM article WHERE publish_time <> 0 and user = ? "
        db_query_str += "ORDER BY `last_modified` DESC, `create_time` DESC, `publish_time` DESC"

    // All of specifying type article that has been published in the database
    case "public-with-type":
        db_query_str =  "SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` "
        db_query_str += "FROM article WHERE publish_time <> 0 and type = ? "
        db_query_str += "ORDER BY `publish_time` DESC"

    // All of article that has been published in the database
    case "public":
        db_query_str =  "SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` "
        db_query_str += "FROM article WHERE publish_time <> 0 "
        db_query_str += "ORDER BY `publish_time` DESC"
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
    for i:= int32(0); rows.Next()  ; i++ {
        var r Article_Format
        rows.Scan(&r.Id, &r.User, &r.Type, &r.Create_time, &r.Publish_time, &r.Last_modified, &r.Title, &r.Content, &r.Attachment)
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
    row := a.db.QueryRow("SELECT `id`, `user`, `type`, `create_time`, `publish_time`, `last_modified`, `title`, `content`, `attachment` FROM article WHERE `id` = ?", serial)

    r := new(Article_Format)
    err := row.Scan(&r.Id, &r.User, &r.Type, &r.Create_time, &r.Publish_time, &r.Last_modified, &r.Title, &r.Content, &r.Attachment)

    if err == sql.ErrNoRows{
        return nil
    }

    // The news has not been published
    // Permission denied
    if r.Publish_time == 0 && r.User != user {
        return nil
    }

    return r
}

func (a *Article) serialNumber(user string) uint32{
    rows := a.db.QueryRow("SELECT `user`, `id`, `title`, `content`, `attachment` FROM article ORDER BY `id` DESC")

    var num uint32
    var title, content, attachment, user_in_db string
    err := rows.Scan(&user_in_db, &num, &title, &content, &attachment)
    if err == sql.ErrNoRows{
        return 1
    }

    /* 如果流水號最大的那個消息是空消息且該消息屬於你自己則刪除該消息，並回傳該消息序號 */
    if title == "" && content == "" && attachment == "" && user_in_db == user{
        stmt, _ := a.db.Prepare("DELETE FROM article WHERE `id` = ?")
        stmt.Exec(num)
        return num
    }
    return num + 1
}
