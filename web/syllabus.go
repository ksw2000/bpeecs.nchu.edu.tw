package web
import(
    "fmt"
    "html/template"
    "net/http"
    "bpeecs.nchu.edu.tw/renderer"
)

func SyllabusWebHandler(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    path := r.URL.Path

    data := initPageData()

    var semester, course_number int
    n, err := fmt.Sscanf(path, "/syllabus/%d/%d", &semester, &course_number)

    if err != nil || n != 2 {
        fmt.Println("未預期的路徑 /syllabus/*", path)
        fmt.Printf("%#v\n", r)
        http.Redirect(w, r, "/error/404", 302)
        return
    }

    var course_name string

    data.Main, course_name, err = renderer.RenderSyllabus(semester, course_number)
    if err != nil {
        fmt.Println("未預期的路徑 /syllabus/*", path)
        fmt.Printf("%#v\n", r)
        http.Redirect(w, r, "/error/404", 302)
        return
    }

    data.Title = fmt.Sprintf("%s | 教學大綱 | 國立中興大學電機資訊學院學士班", course_name)

    // TEMPLATE
    t, _ := template.ParseFiles("./include/layout.gohtml")
    t.Execute(w, data)
}
