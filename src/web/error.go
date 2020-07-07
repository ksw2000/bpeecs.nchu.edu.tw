package web

import(
    "fmt"
    "net/http"
)

func ErrorWeb(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path
    fmt.Println("path", path)
    fmt.Println("scheme", r.URL.Scheme)
}
