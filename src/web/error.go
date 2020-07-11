package web

import(
    "html/template"
    "net/http"
    "time"
)

func ErrorWeb(w http.ResponseWriter, r *http.Request){
    w.WriteHeader(404)
    data := new(PageData);
    data.Title = "HTTP404 Not Found"
    data.Isindex = false
    data.Main = getContent("/http404")
    data.Time = time.Now().Unix();
    t, _ := template.ParseFiles("../include/layout.html")
    t.Execute(w, data)
}
