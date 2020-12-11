package main
import(
    "flag"
    "fmt"
    "log"
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
    mux := http.NewServeMux()
    static_folder := []string{"/assets/", "/.well-known/pki-validation/"}

    for _, v := range static_folder {
        mux.Handle(v, http.StripPrefix(v, http.FileServer(http.Dir("." + v))))
    }

    mux.HandleFunc("/function/", web.FunctionWeb)
    mux.HandleFunc("/error/", web.ErrorWeb)
    mux.HandleFunc("/",  web.BasicWeb)

    server := &http.Server {
        Addr: fmt.Sprintf(":%d", *port),
        Handler: mux,
    }

    if *port == 443 {
        go http.ListenAndServe(":80", http.HandlerFunc(redirect))
        if err := server.ListenAndServeTLS("certificate.crt", "private.key"); err != nil{
            log.Fatal("ListenAndServe: ", err)
        }
    }else{
        if err := server.ListenAndServe(); err != nil{
            log.Fatal("ListenAndServe: ", err)
        }
    }
}
