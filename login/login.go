package login

import(
    "context"
    "database/sql"
    "errors"
    "fmt"
    "log"
    "github.com/go-session/session"
    "net/http"
    "bpeecs.nchu.edu.tw/function"
    "bpeecs.nchu.edu.tw/db"
)

// Login handles manipulations about login
type Login struct{
    IsLogin bool
    UserID string
    UserName string
}

// ErrorReapeatID is returned when the user want to sign up an account which has already existed
var ErrorReapeatID error

// New returns new instance of Login
func New() (l *Login){
    l = new(Login)
    ErrorReapeatID = errors.New("ID-Repeat")
    return
}

// Login is a function handle login
func (l *Login) Login(w http.ResponseWriter, r *http.Request) (err error){
    id, pwd := function.GET("id", r), function.GET("pwd", r)

    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()
    row := d.QueryRow("SELECT `password`, `name`, `salt` FROM user WHERE `id` = ?", id)

    var enryptedPwd, name, salt string
    err = row.Scan(&enryptedPwd, &name, &salt)

    // Check account
    if err == sql.ErrNoRows{
        l   = nil
        err = errors.New(`{"err" : true , "msg" : "Accound not found"}`)
        return
    }

    // Check password
    if pwdHash(pwd, salt) != enryptedPwd{
        l   = nil
        err = errors.New(`{"err" : true , "msg" : "Password is wrong"}`)
        return
    }

    l.IsLogin = true
    l.UserID = id
    l.UserName = name

    // Session srart
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

// NewAcount creates a new account
func (l *Login) NewAcount(id string, pwd string, name string) error{
    // check if there are the same id in db
    d, err := db.Connect(db.Main)
    if err != nil{
        log.Println(err)
        return err
    }
    defer d.Close()

    row := d.QueryRow("SELECT COUNT(*) FROM user WHERE `id` = ?", id)

    count := 0
    if err := row.Scan(&count); err != nil {
        fmt.Println(err)
        return err
    }

    // Check account
    if(count == 0){
        salt := function.RandomString(64);
        pwd = pwdHash(pwd, salt)

        stmt, err := d.Prepare("INSERT INTO user(id, password, salt, name) values(?, ?, ?, ?)")
        if err != nil{
            return err
        }

        stmt.Exec(id, pwd, salt, name)

        return nil
    }

    return ErrorReapeatID
}

// Logout is a function deal with sign out
func (l *Login) Logout(w http.ResponseWriter, r *http.Request) (err error){
    store, err := session.Start(context.Background(), w, r)
    store.Set("isLogin", "no")
    store.Set("userID", "")
    store.Set("userName", "")
    return err
}

// CheckLogin checks if ID and Password is match
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

    return &Login{
        IsLogin: true,
        UserID: userID.(string),
        UserName: userName.(string),
    }
}
