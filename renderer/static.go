package renderer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type courseInfo struct {
	Title  string
	Course []map[string]interface{}
}

const courseDataDir = "./assets/json/course/"
const courseTemplate = "./include/course/template.gohtml"
const courseOutput = "./include/course/"

func convertStrm(i int) string {
	switch i {
	case 1:
		return "上"
	case 2:
		return "下"
	}
	return ""
}

func convertLevel(i int) string {
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

// RenderCourseByYear statically renders course page by inputing year
func RenderCourseByYear(year uint) {
	path := fmt.Sprintf("%s%d.json", courseDataDir, year)

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("render/static.go RenderCourseByYear() can not found " + path)
	}

	temp := courseInfo{}
	yearUnit := []map[string]interface{}{}
	json.Unmarshal(jsonData, &yearUnit)

	path = fmt.Sprintf("%s%d.html", courseOutput, year)
	os.Remove(path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		log.Fatalln("render/static.go RenderCourseByYear() can not found " + path)
	}

	// render title e.g. 109學年度課程內容
	fmt.Fprintf(f, fmt.Sprintf("<h1>%d學年度課程內容</h1>", year))

	for _, s := range yearUnit {
		temp.Title = "大" + convertLevel(int(s["level"].(float64))) +
			convertStrm(int(s["strm"].(float64))) + "學期"
		temp.Course = []map[string]interface{}{}
		for _, v := range s["list"].([]interface{}) {
			info := v.(map[string]interface{})
			// ["course"] ["required"] ["prereq"] ["teacher"]
			teacher := []string{}
			for _, k := range info["teacher"].([]interface{}) {
				teacher = append(teacher, k.(string))
			}
			info["teacher"] = strings.Join(teacher, ",")
			if info["required"].(bool) {
				info["required"] = "必修"
			} else {
				info["required"] = "選修"
			}
			info["link"] = (info["number"].(float64) > 0)
			info["semester"] = fmt.Sprintf("%d%d", year, int(s["strm"].(float64)))
			temp.Course = append(temp.Course, info)
		}

		t, err := template.ParseFiles(courseTemplate)
		if err != nil {
			log.Fatalln("render/static.go RenderCourseByYear() can not found template " +
				courseTemplate)
		}
		t.Execute(f, temp)
	}
}
