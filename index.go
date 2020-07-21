package main

import(
    "log"
    "net/http"
    "bpeecs.nchu.edu.tw/web"
)

func main(){
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./assets/"))
    mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

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
