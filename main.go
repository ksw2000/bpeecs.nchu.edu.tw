package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"bpeecs.nchu.edu.tw/handler"
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

func main() {
	// parse flag
	rend := flag.Bool("r", false, "Render static page or not")
	port := flag.Int("p", 9000, "Port number (default: 9000)")
	flag.Parse()

	// render static page
	if *rend {
		go func() {
			renderer.RenderCourseByYear(109)
		}()
	}

	// web server
	mux := http.NewServeMux()
	staticFolder := []string{"/assets", "/.well-known/pki-validation"}

	for _, v := range staticFolder {
		fileServer := http.FileServer(http.Dir("." + v))
		mux.Handle(v+"/", http.StripPrefix(v, neuter(fileServer)))
	}

	mux.HandleFunc("/function/", handler.FunctionWebHandler)
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

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
