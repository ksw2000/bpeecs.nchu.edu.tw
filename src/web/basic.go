package web
import(
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"
    //---------------
    "function"
    "login"
)

type PageData struct{
    Title string
    Isindex bool
    MAIN_ID string
    Main interface{}
    Time int64
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

func BasicWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    fmt.Println(r.Form)
    path := r.URL.Path
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k := range r.Form{
        fmt.Println("key:", k)
        fmt.Println("val:", function.GET(k, r))
    }

    t, _ := template.ParseFiles("../include/layout.html")

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
            Title : "關於本系",
        }
    case "/course":
        data = PageData{
            Title : "課程內容",
        }
    case "/member":
        data = PageData{
            Title : "系上成員",
        }
    case "/login":
        if login.CheckLogin(w, r) != nil{
            http.Redirect(w, r, "/manage", 302)
            return
        }else{
            data = PageData{
                Title : "登入",
            }
        }

    case "/manage":
        if login.CheckLogin(w, r) != nil{
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
