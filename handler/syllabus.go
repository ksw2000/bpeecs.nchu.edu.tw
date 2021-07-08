package handler

import (
	"fmt"
	"html/template"
	"net/http"
)

// SyllabusWebHandler is a handler for handling url whose prefix is /syllabus
func SyllabusWebHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path
	data := initPageData()

	// route for: /syllabus/year/109
	var year uint
	n, err := fmt.Sscanf(path, "/syllabus/year/%d", &year)
	if err == nil && n == 1 {
		data.Main, err = RenderCourseByYear(year)
		if err != nil {
			NotFound(w, r)
			return
		}
		data.Title = fmt.Sprintf("%d學年度課程內容 | 國立中興大學電機資訊學院學士班", year)
		t, _ := template.ParseFiles("./include/layout.gohtml")
		t.Execute(w, data)
		return
	}

	// route for: /syllabus/1091/1136
	var semester, courseNumber int
	n, err = fmt.Sscanf(path, "/syllabus/%d/%d", &semester, &courseNumber)

	if err == nil && n == 2 {
		var courseName string
		data.Main, courseName, err = RenderSyllabus(semester, courseNumber)
		if err != nil {
			NotFound(w, r)
			return
		}
		data.Title = fmt.Sprintf("%s | 教學大綱 | 國立中興大學電機資訊學院學士班", courseName)
		t, _ := template.ParseFiles("./include/layout.gohtml")
		t.Execute(w, data)
		return
	}

	// else
	NotFound(w, r)
	return
}
