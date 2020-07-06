package functionWeb

import(
    "strconv"
    "strings"
    "article"
    "net/http"
    "fmt"
    "github.com/go-session/session"
    "context"
    "path/filepath"
    "io"
    "os"
    "encoding/json"
    "function"
)

type Login struct{
    IsLogin bool
    UserID string
    UserName string
}

func GET(key string, r *http.Request) string{
    return strings.Join(r.Form[key], "")
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

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func FunctionWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path

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
        art := new(article.Article)
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
        art := new(article.Article)
        art.Connect("./sql/article.db")
        if err := art.GetErr(); err != nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
            return
        }

        // step4: call GetLatestArticle(whatType, from, to)
        if t!=""{
            ret := new(struct{
                NewsList []article.Article_Format
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
    }else if path == "/function/upload"{
        // is login？
        if checkLogin(w, r) == nil{
            fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
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
            f, _ := fh.Open()
            defer f.Close()

            filePath := "../assets/upload/"
            fileExt  := filepath.Ext(fh.Filename)
            fileName := strings.TrimRight(fh.Filename, fileExt)
            for fileExists(filePath + fileName + fileExt){
                fileName = function.RandomString(10)
            }
            newFile, err := os.OpenFile(filePath + string(fileName) + fileExt, os.O_WRONLY | os.O_CREATE, 0666)
            if err != nil{
                fmt.Fprint(w, `{"err" : true , "msg" : "檔案處理錯誤(新建失敗)", "code" : 4}`)
                return;
            }
            _, err = io.Copy(newFile, f)
            if err != nil{
                fmt.Fprint(w, `{"err" : true , "msg" : "檔案處理錯誤(移動失敗)", "code" : 4}`)
                return;
            }else{
                ret.Filename = append(ret.Filename, fileName)
                ret.Filepath = append(ret.Filepath, filePath)
            }
        }
        ret.Err = false
        json.NewEncoder(w).Encode(ret)
    }
}
