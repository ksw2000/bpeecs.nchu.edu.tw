package main
import(
    "flag"
    "fmt"
    "log"
    "time"
    "net/http"
    "bpeecs.nchu.edu.tw/web"
    "bpeecs.nchu.edu.tw/renderer"
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
    // parse flag
    rend := flag.Bool("r", false, "Render static page or not")
    port := flag.Int("p", 9000, "Port number (default: 9000)")
    flag.Parse()

    // render static page
    if *rend{
        go func(){
            renderer.RenderCourseByYear(109)
        }()
    }

    // server
    halting := make(chan bool)
    counter := 0
    go server(port, halting)

    for _ = range halting{
        counter++
        log.Printf("%2d halting, try to restart... after %ds\n", counter, counter)
        time.Sleep(time.Duration(counter)*time.Second)
        log.Println("restarting")
        go server(port, halting)
    }
}

func server(port *int, halting chan bool){
    // server recover()
    defer func(){
       if err := recover(); err != nil {
           log.Println(err)
           halting <- true
       }
    }()

    mux := http.NewServeMux()
    static_folder := []string{"/assets/", "/.well-known/pki-validation/"}

    for _, v := range static_folder {
        mux.Handle(v, http.StripPrefix(v, http.FileServer(http.Dir("." + v))))
    }

    mux.HandleFunc("/function/", web.FunctionWebHandler)
    mux.HandleFunc("/error/", web.ErrorWebHandler)
    mux.HandleFunc("/manage/", web.ManageWebHandler)
    mux.HandleFunc("/syllabus/", web.SyllabusWebHandler)
    mux.HandleFunc("/",  web.BasicWebHandler)

    server := &http.Server {
        Addr: fmt.Sprintf(":%d", *port),
        Handler: mux,
    }

    if *port == 443 {
        go http.ListenAndServe(":80", http.HandlerFunc(redirect))
        if err := server.ListenAndServeTLS("certificate.crt", "private.key"); err != nil{
            log.Fatalln("ListenAndServe: ", err)
        }
    }else{
        if err := server.ListenAndServe(); err != nil{
            log.Fatalln("ListenAndServe: ", err)
        }
    }
    halting <- true
}
