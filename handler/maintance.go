package handler

import (
	"html/template"
	"net/http"
)

// BasicWebHandler is a handler for handling url whose prefix is /
func MaintainWebHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	// Handle static file
	staticFiles := []string{"/robots.txt", "/sitemap.xml", "/favicon.ico"}
	for _, f := range staticFiles {
		if r.URL.Path == f {
			http.StripPrefix("/", http.FileServer(http.Dir("./"))).ServeHTTP(w, r)
			return
		}
	}

	// Handle simple web and check login info
	data := initPageData()
	user := CheckLoginBySession(w, r)
	data.IsLogin = user != nil
	data.Title = "/maintance | 國立中興大學電機資訊學院學士班"
	data.Main, _ = getHTML("/maintance")

	t, _ := template.ParseFiles("./html/layout.gohtml")
	t.Execute(w, data)
}
