package web
import(
    "fmt"
    "html/template"
    "net/http"
    "bpeecs.nchu.edu.tw/renderer"
)

// SyllabusWebHandler is a handler for handling url whose prefix is /syllabus
func SyllabusWebHandler(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path
    data := initPageData()

    var semester, courseNumber int
    n, err := fmt.Sscanf(path, "/syllabus/%d/%d", &semester, &courseNumber)

    if err != nil || n != 2 {
        fmt.Println("未預期的路徑 /syllabus/*", path)
        fmt.Printf("%#v\n", r)
        http.Redirect(w, r, "/error/404", 302)
        return
    }

    var courseName string

    data.Main, courseName, err = renderer.RenderSyllabus(semester, courseNumber)
    if err != nil {
        fmt.Println("未預期的路徑 /syllabus/*", path)
        fmt.Printf("%#v\n", r)
        http.Redirect(w, r, "/error/404", 302)
        return
    }

    data.Title = fmt.Sprintf("%s | 教學大綱 | 國立中興大學電機資訊學院學士班", courseName)

    // TEMPLATE
    t, _ := template.ParseFiles("./include/layout.gohtml")
    t.Execute(w, data)
}
