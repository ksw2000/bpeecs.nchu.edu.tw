package main

import(
    _"fmt"
    "log"
    "net/http"
    "web"
)

func main(){
    fs := http.FileServer(http.Dir("../assets/"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

    http.HandleFunc("/function/", web.FunctionWeb)
    http.HandleFunc("/error/", web.ErrorWeb)
    http.HandleFunc("/",  web.BasicWeb)

    err := http.ListenAndServe(":9000", nil)
    if err != nil{
        log.Fatal("ListenAndServe: ", err)
    }
}
