package main

import(
    "context"
    "fmt"
    "github.com/go-session/session"
    "html/template"
    "io/ioutil"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "function"
    "time"
)

type PageData struct{
    Title string
    Isindex bool
    MAIN_ID string
    Main interface{}
    Time int64
}

type Login struct{
    IsLogin bool
    UserID string
    UserName string
}

func getContent(fileName string) interface{}{
    file, err := os.Open("../include" + fileName + ".html")
    if err != nil{
        log.Fatal(err)
    }
    defer file.Close()
    content, err := ioutil.ReadAll(file)

    return template.HTML(content);
}

func checkLogin(w http.ResponseWriter, r *http.Request) *Login{
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

func GET(key string, r *http.Request) string{
    return strings.Join(r.Form[key], "")
}

func basicWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    fmt.Println(r.Form)
    path := r.URL.Path
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k := range r.Form{
        fmt.Println("key:", k)
        fmt.Println("val:", GET(k, r))
    }

    t, _ := template.ParseFiles("layout.html")

    var data PageData
    switch path {
    case "/":
        data = PageData{
            Title : "國立中興大學電機資訊學院學士班",
            Isindex : true,
            MAIN_ID : "main-for-index",
            Main :getContent("/index"),
        }
    case "/news":
        data = PageData{
            Title : "最新消息",
        }
    case "/about":
        data = PageData{
            Title : "本系簡介",
        }
    case "/course":
        data = PageData{
            Title : "課程內容",
        }
    case "/member":
        data = PageData{
            Title : "系上成員",
        }
    case "/recruit":
        data = PageData{
            Title : "招生資訊",
        }
    case "/login":
        if checkLogin(w, r) != nil{
            http.Redirect(w, r, "/manage", 302)
            return
        }else{
            data = PageData{
                Title : "登入",
            }
        }

    case "/manage":
        if checkLogin(w, r) != nil{
            data = PageData{
                Title : "管理模式",
            }
        }else{
            http.Redirect(w, r, "/?notlogin", 302)
            return
        }
    default:
        fmt.Println("未預期的路徑")
        fmt.Println(path)
        http.Redirect(w, r, "/error/404", 302)
        return
    }

    if(path != "/"){
        data.Title += " | 國立中興大學電機資訊學院學士班"
        data.Isindex = false
        data.Main = getContent(path)
        data.MAIN_ID = "main"
    }

    data.Time = time.Now().Unix();

    t.Execute(w, data);
}

func functionWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)

    if path == "/function/login" {
        if GET("id", r) != "root"{
            fmt.Fprint(w, `{"err" : true , "msg" : "無此帳號"}`)
            return;
        }else if GET("pwd", r) != "00000000"{
            fmt.Fprint(w, `{"err" : true , "msg" : "密碼錯誤"}`)
            return
        }else{
            //Session srart
            store, err := session.Start(context.Background(), w, r)
            if err != nil {
                fmt.Fprint(w, err)
                return
            }

            store.Set("isLogin", "true")
            store.Set("loginID", "root")

            err = store.Save()
            if err != nil {
                fmt.Fprint(w, err)
                return
            }

            fmt.Fprint(w, `{"err" : false}`)
        }
    }else if path == "/function/add_news" {
        // is login？
        if checkLogin(w, r) == nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
            return
        }

        user := "root"

        // step1: connect to database
        art := new(function.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step2: get serial number
        serial := art.NewArticle(user)
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結成功但新增文章失敗", "code": 2}`)
            return
        }
        ret := fmt.Sprintf(`{"err" : false, "msg" : %d}`, serial)
        fmt.Fprint(w, ret)
    }else if path == "/function/save_news" || path == "/function/publish_news" || path == "/function/del_news" {
        // is login？
        if checkLogin(w, r) == nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
            return
        }

        // write to database
        // step1: fetch http POST
        num, err := strconv.Atoi(GET("serial", r))
        if err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "文章代碼錯誤 (POST參數錯誤)", "code": 3}`)
            return
        }
        serial := uint32(num)
        user := "root"
        title := GET("title", r)
        content := GET("content", r)

        // step2: connect to database
        art := new(function.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step3: call SaveArticle() or PublishArticle()
        if path == "/function/save_news" {
            art.SaveArticle(serial, user, title, content)
        }else if path == "/function/publish_news" {
            art.PublishArticle(serial, user, title, content)
        }else if path == "/function/del_news" {
            art.DelArticle(serial, user)
        }

        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結成功但操作文章失敗", "code": 2}`)
            return
        }
        fmt.Fprint(w, `{"err" : false}`)
    }else if path == "/function/get_news"{
        // read news from database
        // step1: read GET
        t := GET("type", r)
        n := GET("id", r)
        var serial uint32
        from, to := 0, 19   // Default from = 0, to = 19

        if t != "public" && t != "all" && t != "draft"{
            if n == ""{
                fmt.Fprint(w, `{"err" : true , "msg" : "錯誤的請求 (GET 參數錯誤)", "code": 3}`)
                return;
            }else{
                num, err := strconv.Atoi(n)
                if err != nil{
                    fmt.Fprint(w, `{"err" : true , "msg" : "文章代碼錯誤 (GET 參數錯誤)", "code": 3}`)
                    return
                }
                serial = uint32(num)
            }
        }else{
            if f, t := GET("from", r), GET("to", r); f != "" && t != ""{
                var err error
                from, err = strconv.Atoi(f)
                to, err = strconv.Atoi(t)
                if err != nil{
                    fmt.Fprint(w, `{"err" : true , "msg" : "from to 代碼錯誤 (GET 參數錯誤)", "code": 3}`)
                    return
                }
            }
        }

        // step2: some request need user id
        user := ""
        if loginInfo := checkLogin(w, r); loginInfo != nil{
            user = loginInfo.UserID
        }

        // step3: connect to database
        art := new(function.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step4: call GetLatestArticle(whatType, from, to)
        if t!=""{
            ret := new(struct{
                NewsList []function.Article_Format
                HasNext bool
                Err error
            })
            ret.NewsList, ret.HasNext = art.GetLatestArticle(t, user, int32(from), int32(to))
            ret.Err = nil;

            // step5: encode to json
            // art.GetArtList()
            json.NewEncoder(w).Encode(ret)
        }else if n!=""{
            if ret := art.GetArticleBySerial(serial, user); ret != nil{
                json.NewEncoder(w).Encode(ret)
            }else{
                fmt.Fprint(w,`{}`)
            }
        }
    }
}

func errorWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path
    fmt.Println("path", path)
    fmt.Println("scheme", r.URL.Scheme)
}

func main(){
    fs := http.FileServer(http.Dir("../assets/"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

    http.HandleFunc("/function/", functionWeb)
    http.HandleFunc("/error/", errorWeb)
    http.HandleFunc("/", basicWeb)
    err := http.ListenAndServe(":9000", nil)
    if err != nil{
        log.Fatal("ListenAndServe: ", err)
    }
}
