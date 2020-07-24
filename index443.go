package main

import(
    "log"
    "net/http"
    "os"
    "bpeecs.nchu.edu.tw/web"
)

func redirect(w http.ResponseWriter, req *http.Request) {
    target := "https://" + req.Host + req.URL.Path
    if len(req.URL.RawQuery) > 0 {
        target += "?" + req.URL.RawQuery
    }
    log.Printf("redirect to: %s", target)
    http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

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
    go http.ListenAndServe(":80", http.HandlerFunc(redirect))

    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./assets/"))
    mux.Handle("/assets/", StaticSiteHandler(http.StripPrefix("/assets/", fs)))

    // https://www.sslforfree.com/
    /*
    well_known := http.FileServer(http.Dir("./.well-known/pki-validation/"))
    mux.Handle("/.well-known/pki-validation/", http.StripPrefix("/.well-known/pki-validation/", well_known))
    */

    mux.HandleFunc("/function/", web.FunctionWeb)
    mux.HandleFunc("/error/", web.ErrorWeb)
    mux.HandleFunc("/",  web.BasicWeb)

    server := &http.Server {
        Addr: ":443",
        Handler: mux,
    }

    if err := server.ListenAndServeTLS("certificate.crt", "private.key"); err != nil{
        log.Fatal("ListenAndServe: ", err)
    }
}
