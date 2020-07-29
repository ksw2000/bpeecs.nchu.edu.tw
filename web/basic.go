package web
import(
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"
    "bpeecs.nchu.edu.tw/login"
)

type PageData struct{
    Title string
    Isindex bool
    IsLogin bool
    Main interface{}
    Time int64
}

func getContent(fileName string) interface{}{
    file, err := os.Open("./include" + fileName + ".html")
    if err != nil{
        log.Fatal(err)
    }
    defer file.Close()
    content, err := ioutil.ReadAll(file)

    return template.HTML(content);
}

func BasicWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path

    if path == "/favicon.ico" {
        http.Redirect(w, r, "/assets/img/favicon.ico", 301)
        return
    }
    // TEMPLATE
    t, _ := template.ParseFiles("./include/layout.html")

    data := new(PageData)

    // Is login?
    if login.CheckLogin(w, r) != nil{
        data.IsLogin = true
    }else{
        data.IsLogin = false
    }

    var simpleWeb = map[string]string{
        "/news" : "最新消息",
        "/about/education-goal-and-core-ability" : "教育目標及核心能力",
        "/about/enrollment" : "招生方式",
        "/about/feature" : "特色",
        "/about/future-development-direction" : "學生未來發展方向",
        "/about/why-establish" : "創系緣由",
        "/course" : "課程內容",
        "/course/graduation-conditions" : "畢業條件",
        "/member/admin-staff" : "行政人員",
        "/member/faculty" : "師資陣容",
        "/member/class-teacher" : "班主任",
    }

    title, ok := simpleWeb[path]

    if ok {
        data.Title = title
    }else{
        switch path {
        case "/":
            data.Title = "國立中興大學電機資訊學院學士班"
            data.Isindex = true
            data.Main = getContent("/index")
        case "/login":
            if login.CheckLogin(w, r) != nil{
                http.Redirect(w, r, "/manage", 302)
                return
            }else{
                data.Title = "登入"
            }
        case "/logout":
            l := login.New()
            l.Connect("./sql/user.db")
            if err := l.Logout(w, r); err!=nil {
                fmt.Fprint(w, `{"err" : true, "msg" : "登出失敗"}`)
            }else{
                http.Redirect(w, r, "/", 302)
                return
            }
        case "/manage":
            if data.IsLogin{
                data.Title = "管理模式"
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
    }

    if(path != "/"){
        data.Title += " | 國立中興大學電機資訊學院學士班"
        data.Isindex = false
        data.Main = getContent(path)
    }

    data.Time = time.Now().Unix()

    t.Execute(w, data)
}
