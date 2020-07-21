package login

import(
    "context"
    "database/sql"
    "errors"
    "fmt"
    "github.com/go-session/session"
    _"github.com/mattn/go-sqlite3"
    "net/http"
    "bpeecs.nchu.edu.tw/function"
)

type Login struct{
    IsLogin bool
    UserID string
    UserName string
    db *sql.DB
}

func New() (l *Login){
    l = new(Login)
    return
}

func (l *Login) Connect(path string) (err error){
    l.db, err = sql.Open("sqlite3", path)
    return
}

func (l *Login) Login(w http.ResponseWriter, r *http.Request) (err error){
    id, pwd := function.GET("id", r), function.GET("pwd", r)
    row := l.db.QueryRow("SELECT `password`, `name`, `salt` FROM user WHERE `id` = ?", id)

    var pwd_in_db, name, salt string
    err = row.Scan(&pwd_in_db, &name, &salt)
    defer l.db.Close()
    // Check account
    if err == sql.ErrNoRows{
        l   = nil
        err = errors.New(`{"err" : true , "msg" : "Accound not found"}`)
        return
    }

    // Check password
    if pwdHash(pwd, salt) != pwd_in_db{
        l   = nil
        err = errors.New(`{"err" : true , "msg" : "Password is wrong"}`)
        return
    }

    l.IsLogin = true
    l.UserID = id
    l.UserName = name

    //Session srart
    store, err := session.Start(context.Background(), w, r)
    if err != nil {
        err = errors.New(`{"err" : true , "msg" : "Session start error"}`)
        return
    }

    store.Set("isLogin", "yes")
    store.Set("userID", l.UserID)
    store.Set("userName", l.UserName)
    err = store.Save()
    if  err != nil {
        err = errors.New(`{"err" : true , "msg" : "Session store error"}`)
        return
    }

    err = nil

    return
}

func (l *Login) Logout(w http.ResponseWriter, r *http.Request) (err error){
    store, err := session.Start(context.Background(), w, r)

    store.Set("isLogin", "no")
    store.Set("userID", "")
    store.Set("userName", "")

    return err
}

func (l *Login) NewAcount(id string, pwd string, name string) error{
    salt := function.RandomString(64);
    pwd = pwdHash(pwd, salt)
    stmt, err := l.db.Prepare("INSERT INTO user(id, password, salt, name) values(?, ?, ?, ?)")
    if err != nil{
        return err
    }

    stmt.Exec(id, pwd, salt, name)

    return nil
}

func CheckLogin(w http.ResponseWriter, r *http.Request) *Login{
    store, err := session.Start(context.Background(), w, r)
    if err != nil {
        fmt.Fprint(w, err)
        return nil
    }

    isLogin, ok1 := store.Get("isLogin")
    userID, ok2 := store.Get("userID")
    userName, ok3 := store.Get("userName")

    if !(ok1 && ok2 && ok3){
        return nil
    }

    if isLogin != "yes"{
        return nil
    }

    l := new(Login)
    l.IsLogin = true
    l.UserID = userID.(string)
    l.UserName = userName.(string)

    return l
}
