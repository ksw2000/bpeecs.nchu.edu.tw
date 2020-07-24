package main

import(
    "log"
    "net/http"
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

func main(){
    go http.ListenAndServe(":80", http.HandlerFunc(redirect))

    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./assets/"))
    mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

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
