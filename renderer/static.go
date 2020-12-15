package renderer
import(
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "os"
    "strings"
)

type Course_render struct{
    Title string
    Course []map[string]interface{}
}

const COURSE_DATA_DIR = "./assets/json/course/"
const COURSE_TEMPLATE = "./include/course/template.gohtml"
const COURSE_OUTPUT = "./include/course/"

func translate_strm(i int) string{
    switch i {
    case 1:
        return "上"
    case 2:
        return "下"
    }
    return ""
}

func translate_level(i int) string{
    switch i {
    case 1:
        return "一"
    case 2:
        return "二"
    case 3:
        return "三"
    case 4:
        return "四"
    }
    return ""
}

func RenderCourseByYear(year uint){
    path := fmt.Sprintf("%s%d.json", COURSE_DATA_DIR, year)

    json_data, err := ioutil.ReadFile(path)
	if err != nil {
		panic("render/static.go RenderCourseByYear() can not found " + path)
	}

    temp := Course_render{}
    year_unit := []map[string]interface{}{}
    json.Unmarshal(json_data, &year_unit)

    path = fmt.Sprintf("%s%d.html", COURSE_OUTPUT, year)
    os.Remove(path)
    f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
    defer f.Close()
    if err != nil{
        panic("render/static.go RenderCourseByYear() can not found " + path)
    }

    // render title e.g. 109學年度課程內容
    fmt.Fprintf(f, fmt.Sprintf("<h1>%d學年度課程內容</h1>", year))

    for _, s := range year_unit{
        temp.Title  = "大" + translate_level(int(s["level"].(float64)))
        temp.Title += translate_strm(int(s["strm"].(float64))) + "學期"
        temp.Course = []map[string]interface{}{}
        for _, v := range s["list"].([]interface{}){
            info := v.(map[string]interface{})
            // ["course"] ["required"] ["prereq"] ["teacher"]
            teacher := []string{}
            for _, k := range info["teacher"].([]interface{}){
                teacher = append(teacher, k.(string))
            }
            info["teacher"] = strings.Join(teacher, ",")
            if info["required"].(bool){
                info["required"] = "必修"
            }else{
                info["required"] = "選修"
            }

            temp.Course = append(temp.Course, info)
        }

        t, err := template.ParseFiles(COURSE_TEMPLATE)
        if err != nil{
            panic("render/static.go RenderCourseByYear() can not found template " +
            COURSE_TEMPLATE)
        }
        t.Execute(f, temp)
    }
}
