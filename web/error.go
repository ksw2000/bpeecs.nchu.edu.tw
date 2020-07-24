package web

import(
    "html/template"
    "net/http"
    "time"
)

func ErrorWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path
    data := new(PageData)

    switch path {
    case "/error/404":
        w.WriteHeader(404)
        data.Title = "HTTP 404 Not Found"
        data.Isindex = false
        data.Main = getContent("/http404")
        data.Time = time.Now().Unix()
    case "/error/403":
        w.WriteHeader(403)
        data.Title = "HTTP 403 Forbidden"
        data.Isindex = false
        data.Main = getContent("/http403")
        data.Time = time.Now().Unix()
    }

    t, _ := template.ParseFiles("./include/layout.html")
    t.Execute(w, data)
}
