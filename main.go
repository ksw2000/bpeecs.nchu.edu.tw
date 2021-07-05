package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"bpeecs.nchu.edu.tw/handler"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

func main() {
	// parse flag
	port := flag.Int("p", 9000, "Port number (default: 9000)")
	flag.Parse()

	// web server
	mux := http.NewServeMux()
	staticFolder := []string{"/assets", "/.well-known/pki-validation"}

	// minify static files
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	for _, dir := range staticFolder {
		fileServer := http.FileServer(http.Dir("." + dir))
		mux.Handle(dir+"/", http.StripPrefix(dir, m.Middleware(neuter(fileServer))))
	}

	mux.HandleFunc("/api/", handler.ApiHandler)
	mux.HandleFunc("/error/", handler.ErrorWebHandler)
	mux.HandleFunc("/manage/", handler.ManageWebHandler)
	mux.HandleFunc("/syllabus/", handler.SyllabusWebHandler)
	mux.HandleFunc("/", handler.BasicWebHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	if *port == 443 {
		fmt.Println("https://localhost")
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))
		if err := server.ListenAndServeTLS("certificate.crt", "private.key"); err != nil {
			log.Fatalln("ListenAndServe: ", err)
		}
	} else {
		fmt.Printf("http://localhost:%d\n", *port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln("ListenAndServe: ", err)
		}
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			handler.Forbidden(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
