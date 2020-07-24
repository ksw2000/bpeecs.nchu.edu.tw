package main

import(
    "log"
    "os"
    "net/http"
    "bpeecs.nchu.edu.tw/web"
)

func StaticSiteHandler(h http.Handler) http.Handler{
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        r.ParseForm()
        path := r.URL.Path

        file, err := os.Open("." + path)
        // cannot open
        if err!= nil{
            http.Redirect(w, r, "/error/404", 302)
            return
        }

        fileInfo, err := file.Stat()

        // cannot open
        if err!= nil{
            http.Redirect(w, r, "/error/404", 302)
            return
        }

        // the path is dir
        if fileInfo.IsDir() {
            http.Redirect(w, r, "/error/403", 302)
            return
        }

        // default method (call http.FileServer())
        h.ServeHTTP(w, r)
    })
}

func main(){
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./assets/"))
    mux.Handle("/assets/", StaticSiteHandler(http.StripPrefix("/assets/", fs)))

    mux.HandleFunc("/function/", web.FunctionWeb)
    mux.HandleFunc("/error/", web.ErrorWeb)
    mux.HandleFunc("/",  web.BasicWeb)

    server := &http.Server {
        Addr: ":9000",
        Handler: mux,
    }

    //if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil{
    if err := server.ListenAndServe(); err != nil{
        log.Fatal("ListenAndServe: ", err)
    }
}
