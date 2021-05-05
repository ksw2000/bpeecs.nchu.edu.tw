package handler

import (
	"html/template"
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
	}

	t, _ := template.ParseFiles("./include/layout.gohtml")
	t.Execute(w, data)
}
