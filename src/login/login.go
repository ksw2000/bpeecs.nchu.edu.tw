package login

import(
    "context"
    "fmt"
    "github.com/go-session/session"
    "net/http"
)

type Login struct{
    IsLogin bool
    UserID string
    UserName string
}

func CheckLogin(w http.ResponseWriter, r *http.Request) *Login{
    store, err := session.Start(context.Background(), w, r)
    if err != nil {
        fmt.Fprint(w, err)
        return nil
    }

    _, ok := store.Get("isLogin")
    if !ok {
        return nil
    }

    ret := new(Login)
    ret.IsLogin = true
    userID, ok := store.Get("loginID")
    if !ok{
        return nil
    }
    ret.UserID = userID.(string)
    ret.UserName = "無名氏"
    return ret
}
