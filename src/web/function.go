package web

import(
    "context"
    "fmt"
    "encoding/json"
    "github.com/go-session/session"
    "net/http"
    "strconv"
    //-------------
    "article"
    "login"
    "function"
    "files"
)

func FunctionWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path

    if path == "/function/login" {
        if function.GET("id", r) != "root"{
            fmt.Fprint(w, `{"err" : true , "msg" : "無此帳號"}`)
            return;
        }else if function.GET("pwd", r) != "00000000"{
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
    }else if path == "/function/logout" {
        //Session srart
        store, err := session.Start(context.Background(), w, r)
        if err != nil {
            fmt.Fprint(w, err)
            return
        }

        store.Set("isLogin", "false")
        store.Set("loginID", "")
        err = store.Save()
        if err != nil {
            fmt.Fprint(w, err)
            return
        }

        fmt.Fprint(w, `{"err" : false}`)
    }else if path == "/function/add_news" {
        // is login？
        if login.CheckLogin(w, r) == nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
            return
        }

        user := "root"

        // step1: connect to database
        art := new(article.Article)
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
        if login.CheckLogin(w, r) == nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
            return
        }

        // write to database
        // step1: fetch http POST
        num, err := strconv.Atoi(function.GET("serial", r))
        if err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "文章代碼錯誤 (POST參數錯誤)", "code": 3}`)
            return
        }
        serial := uint32(num)
        user := "root"
        title := function.GET("title", r)
        content := function.GET("content", r)
        attachment := function.GET("attachment", r)    //string, already convert to string in front-end

        // step2: connect to database
        art := new(article.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step3: call Save() or Publish()
        if path == "/function/save_news" {
            art.Save(serial, user, title, content, attachment)
        }else if path == "/function/publish_news" {
            art.Publish(serial, user, title, content, attachment)
        }else if path == "/function/del_news" {
            art.Del(serial, user)
        }

        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結成功但操作文章失敗", "code": 2}`)
            return
        }
        fmt.Fprint(w, `{"err" : false}`)
    }else if path == "/function/get_news"{
        // read news from database
        // step1: read GET
        t := function.GET("type", r)
        n := function.GET("id", r)
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
            if f, t := function.GET("from", r), function.GET("to", r); f != "" && t != ""{
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
        if loginInfo := login.CheckLogin(w, r); loginInfo != nil{
            user = loginInfo.UserID
        }

        // step3: connect to database
        art := new(article.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step4: call GetLatest(whatType, from, to)
        if t!=""{
            ret := new(struct{
                NewsList []article.Article_Format
                HasNext bool
                Err error
            })
            ret.NewsList, ret.HasNext = art.GetLatest(t, user, int32(from), int32(to))
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
    }else if path == "/function/upload"{
        // is login？
        if login.CheckLogin(w, r) == nil{
            fmt.Fprint(w, `{"Err" : true , "Msg" : "尚未登入", "Code" : 1}`)
            return
        }

        r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
        fhs := r.MultipartForm.File["files"]

        ret := new(struct{
            Err bool
            Filename []string
            Filepath []string
        })

        for _, fh := range fhs {
            f := new(files.Files)
            f.Connect("./sql/files.db")
            if err := f.GetErr(); err != nil{
                fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
                return
            }
            f = f.NewFile(fh)
            if err := f.GetErr(); err != nil{
                fmt.Fprint(w, `{"err" : true , "msg" : "新增檔案失敗", "code": 4}`)
                return
            }
            ret.Filename = append(ret.Filepath, f.Server_name)
            ret.Filepath = append(ret.Filepath, f.Path)
        }
        ret.Err = false
        json.NewEncoder(w).Encode(ret)
    }
}
