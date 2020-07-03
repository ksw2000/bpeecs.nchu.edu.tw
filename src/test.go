package main
import(
    "database/sql"
    "fmt"
    "time"
    _"github.com/mattn/go-sqlite3"
)

type Article struct{
    db *sql.DB
    artList []Article_Format
    err error
}

type Article_Format struct{
    id uint32
    user string
    create_time uint64
    publish_time uint64
    last_modified uint64
    title string
    content string
}

type ArticleCtrl interface{
    // connect to detabase
    Connect(path string) *sql.DB
    // Create a new article and return last_insert_id
    AddArticle(user string, title string, content string) int64
    // Get the latest article
    GetLatestArticle(whatType string, from int32, to int32) []Article_Format
    // Get private variables
    GetErr() error
    GetArtList() []Article_Format

    // Generate serial number
    serialNumber() uint32

    // Error process
    errProcess(err error)
}

func (a *Article) GetErr() error{
    return a.err
}

func (a *Article) GetArtList() []Article_Format{
    return a.artList
}

func (a *Article) errProcess(err error){
    if err!=nil{
        fmt.Println(err)
        a.err = err
        return
    }
}

func (a *Article) Connect(path string) *sql.DB{
    db, err := sql.Open("sqlite3", path)
    a.db = db
    a.errProcess(err)
    return a.db
}

func (a *Article) AddArticle(user string, title string, content string) int64{
    stmt, err := a.db.Prepare("INSERT INTO article(id, user, create_time, publish_time, last_modified, title, content) values(?, ?, ?, ?, ?, ?, ?)")
    a.errProcess(err)
    now := time.Now().Unix()
    res, err := stmt.Exec(a.serialNumber(), user, now, 0, 0, title, content)
    a.errProcess(err)
    id, err := res.LastInsertId()
    a.errProcess(err)
    return id
}

func (a *Article) GetLatestArticle(whatType string, from int32, to int32) []Article_Format{
    var db_query_str string

    switch whatType {
    case "all":
        db_query_str = "SELECT * FROM article ORDER BY `last_modified` DESC, `create_time` DESC, `publish_time` DESC"
    case "private":
        db_query_str = "SELECT * FROM article WHERE publish_time = 0 ORDER BY `last_modified` DESC, `create_time` DESC"
    case "public":
        db_query_str = "SELECT * FROM article WHERE publish_time <> 0 ORDER BY `publish_time` DESC"
    default:
        return nil
    }

    db_query_str += fmt.Sprintf(" limit %d, %d", from, to-from+1)

    rows, err := a.db.Query(db_query_str)
    a.errProcess(err)

    list := []Article_Format{}

    for rows.Next() {
        var r Article_Format
        err = rows.Scan(&r.id, &r.user, &r.create_time, &r.publish_time, &r.last_modified, &r.title, &r.content)
        a.errProcess(err)
        list = append(list, r)
    }

    a.artList = list

    return list
}

func (a *Article) serialNumber() uint32{
    rows, err := a.db.Query("SELECT `id` FROM article ORDER BY `id` DESC limit 0, 1")
    a.errProcess(err)

    if !rows.Next(){
        return 0
    }

    var num uint32
    err = rows.Scan(&num)
    a.errProcess(err)

    return num+1
}

func main(){
    art := new(Article)
    art.Connect("./sql/article.db")
    num := art.serialNumber()
    fmt.Println(num)
}
