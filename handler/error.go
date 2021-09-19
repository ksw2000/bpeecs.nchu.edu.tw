package handler

import (
	"html/template"
	"log"
	"net/http"
)

// ErrorWebHandler is a handler for handling url whose prefix is /error
func ErrorWebHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path
	data := initPageData()

	switch path {
	case "/error/404":
		w.WriteHeader(404)
		data.Title = "HTTP 404 Not Found"
		data.Main, _ = getHTML("/http404")
	case "/error/403":
		w.WriteHeader(403)
		data.Title = "HTTP 403 Forbidden"
		data.Main, _ = getHTML("/http403")
	default:
		NotFound(w, r)
	}

	t, _ := template.ParseFiles("./include/layout.gohtml")
	t.Execute(w, data)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP404 %s IP: %s\n", r.URL.Path, r.RemoteAddr)
	http.Redirect(w, r, "/error/404", 302)
	return
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP403 %s IP: %s\n", r.URL.Path, r.RemoteAddr)
	http.Redirect(w, r, "/error/403", 302)
	return
}
