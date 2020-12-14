package web
import(
    "bytes"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"
    "bpeecs.nchu.edu.tw/article"
    "bpeecs.nchu.edu.tw/function"
    "bpeecs.nchu.edu.tw/login"
    "bpeecs.nchu.edu.tw/renderer"
)

type PageData struct{
    Title   string
    Isindex bool
    IsLogin bool
    Main    template.HTML
    Time    int64
}

func getContent(fileName string) template.HTML{
    file, err := os.Open("./include" + fileName + ".html")
    defer file.Close()
    if err != nil{
        log.Fatal(err)
    }
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
    t, _ := template.ParseFiles("./include/layout.gohtml")

    data := new(PageData)
    data.Isindex = false    // default value

    var simpleWeb = map[string]string{
        "/about/education-goal-and-core-ability" : "教育目標及核心能力",
        "/about/enrollment" : "招生方式",
        "/about/feature" : "特色",
        "/about/future-development-direction" : "學生未來發展方向",
        "/about/why-establish" : "創系緣由",
        "/course" : "課程內容",
        "/course/graduation-conditions" : "畢業條件",
        "/course/109" : "109學年度課程內容",
        "/member/admin-staff" : "行政人員",
        "/member/faculty" : "師資陣容",
        "/member/class-teacher" : "班主任",
    }

    // Is login?
    loginInfo := login.CheckLogin(w, r)
    if loginInfo != nil{
        data.IsLogin = true
    }else{
        data.IsLogin = false
    }

    var ok bool
    data.Title, ok = simpleWeb[path]

    if !ok {
        var manageWeb = map[string]string{
            "/manage" : "歡迎進入後台管理系統",
            "/manage/article" : "文章管理",
            "/manage/reg" : "註冊新用戶",
            "/manage/reg-done" : "新用戶註冊成功",
        }

        data.Title, ok = manageWeb[path]
        if ok{
            if data.IsLogin{
                manageTemplate, _ := template.ParseFiles("./include/manage.html")
                var manageTemplateByte bytes.Buffer
                manageTemplateData := struct{
                    UserID      string
                    UserName    string
                }{
                    UserID: loginInfo.UserID,
                    UserName: loginInfo.UserName,
                }
                manageTemplate.Execute(&manageTemplateByte, manageTemplateData)
                data.Main = template.HTML(manageTemplateByte.String())
            }else{
                http.Redirect(w, r, "/?notlogin", 302)
                return
            }
        }
    }

    if !ok {
        switch path {
        case "/":
            data.Title = "國立中興大學電機資訊學院學士班"
            data.Isindex = true

            template_index, _ := template.ParseFiles("./include/index.html")
            art := article.New();
            art.Connect("./db/main.db")

            // Default from = 0, to = 19
            // return (list []art.Article_Format, hasNext bool)
            artFormatList, _ := art.GetLatest("public", "normal", "", int32(0), int32(7))
            data_index := new(struct{
                Article_list_brief template.HTML
            })
            data_index.Article_list_brief = renderer.RenderPublicArticleBriefList(artFormatList)

            var buf bytes.Buffer
            template_index.Execute(&buf, data_index)
            data.Main = template.HTML(buf.String())
        case "/news":
            data.Title = "最新消息"
            artType := function.GET("type", r);
            var dict = map[string]string{
                "normal" : "一般消息",
                "activity" : "演講 & 活動",
                "course" : "課程 & 招生",
                "scholarships" : "獎學金訊息",
                "recruit" : "徵才資訊",
            }
            subtitle, ok := dict[artType]
            if ok{
                data.Title = subtitle +" | "+ data.Title;
            }

            if id := function.GET("id", r); id != ""{
                //id is uint32
                serial_u64, err := strconv.ParseUint(id, 10, 32)

                if err != nil{
                    http.Redirect(w, r, "/error/404", 302)
                    return
                }else{
                    art := article.New();
                    art.Connect("./db/main.db")

                    user := ""
                    if data.IsLogin{
                        user = loginInfo.UserID
                    }

                    artInfo := art.GetArticleBySerial(uint32(serial_u64), user)

                    // avoid /news?id=xxx
                    if artInfo == nil{
                        http.Redirect(w, r, "/error/404", 302)
                        return
                    }

                    data.Title = artInfo.Title + " | 國立中興大學電機資訊學院學士班"
                    data.Main  = renderer.RenderPublicArticle(artInfo)
                }
            }else{
                data.Title += " | 國立中興大學電機資訊學院學士班"
                data.Main = getContent(path)
            }
        case "/login":
            if login.CheckLogin(w, r) != nil{
                http.Redirect(w, r, "/manage", 302)
                return
            }else{
                data.Title = "登入"
            }
        case "/logout":
            l := login.New()
            l.Connect("./db/main.db")
            if err := l.Logout(w, r); err!=nil {
                fmt.Fprint(w, `{"err" : true, "msg" : "登出失敗"}`)
            }else{
                http.Redirect(w, r, "/", 302)
                return
            }

        default:
            fmt.Println("未預期的路徑", path)
            fmt.Printf("%#v\n", r)

            http.Redirect(w, r, "/error/404", 302)
            return
        }
    }

    if(path == "/manage"){
        data.Title += " | 國立中興大學電機資訊學院學士班"
        // retain data.Main
    }else if(path == "/news"){
    }else if(path != "/"){
        data.Title += " | 國立中興大學電機資訊學院學士班"
        data.Main = getContent(path)
    }

    data.Time = time.Now().Unix() >> 10

    t.Execute(w, data)
}
