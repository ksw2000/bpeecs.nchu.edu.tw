package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

const articleTemplate = "./include/article_layout.gohtml"

const syllabusDataDir = "./assets/json/syllabus/"
const syllabusTemplate = "./include/syllabus/template.gohtml"

const courseDataDir = "./assets/json/course/"
const courseTemplate = "./include/course/template.gohtml"

const calendarTemplate = "./include/calendar_layout.gohtml"
const indexTemplate = "./include/index.gohtml"

type courseInfo struct {
	Subtitle string
	Course   []map[string]interface{}
}

type fileInfo struct {
	ClientName string
	ServerName string
	Mime       string
	Path       string
}

type articleRenderInfo struct {
	ID              int64
	User            string
	Type            string
	CreateTime      string
	PublishTime     string
	LastModified    string
	Title           string
	Content         template.HTML
	Attachment      []fileInfo
	PhotoAttachment []fileInfo // render photos by HTML <img>
}

type calendarRenderInfo struct {
	Calendar
	HaveLink bool
	ReadOnly bool
}

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

func renderDate(timestamp uint64) string {
	t := time.Unix(int64(timestamp), 0)
	return t.Format("2006-01-02")
	//(t.Format("2006-01-02 15:04:05"))
}

func renderArticleType(key string) string {
	dict := map[string]string{
		"normal":       "一般消息",
		"activity":     "演講 & 活動",
		"course":       "課程 & 招生",
		"scholarships": "獎學金",
		"recruit":      "徵才資訊",
	}

	val, ok := dict[key]
	if ok {
		return val
	}
	return ""
}

// RenderPublicArticle renders an article at url: /news/[articleID]
func RenderPublicArticle(artInfo *Article) template.HTML {
	data := new(articleRenderInfo)
	data.ID = artInfo.ID
	data.User = artInfo.User
	data.Type = renderArticleType(artInfo.Type)
	data.CreateTime = renderDate(artInfo.CreateTime)
	data.PublishTime = renderDate(artInfo.PublishTime)
	data.LastModified = renderDate(artInfo.LastModified)
	data.Title = artInfo.Title
	data.Content = template.HTML(artInfo.Content)
	data.PhotoAttachment = []fileInfo{}
	data.Attachment = []fileInfo{}

	for _, v := range artInfo.Attachment {
		var extName string
		matchNum, _ := fmt.Sscanf(v.Mime, "image/%s", &extName)
		if matchNum > 0 {
			data.PhotoAttachment = append(data.PhotoAttachment, fileInfo{
				Path: v.Path,
			})
		} else {
			data.Attachment = append(data.Attachment, fileInfo{
				ClientName: v.ClientName,
				ServerName: v.ServerName,
				Mime:       v.Mime,
				Path:       v.Path,
			})
		}
	}

	var buf bytes.Buffer
	t, err := template.ParseFiles(articleTemplate)

	if err != nil {
		log.Println("handler/renderer.go RenderPublicArticle() template error " + err.Error())
		return template.HTML("")
	}

	t.Execute(&buf, data)
	return template.HTML(buf.String())
}

// RenderPublicArticleBriefList dynamically renders article list an home page
func RenderPublicArticleBriefList(artInfoList []Article) template.HTML {
	data := new(articleRenderInfo)
	ret := ""
	for _, artInfo := range artInfoList {
		data.ID = artInfo.ID
		data.PublishTime = renderDate(artInfo.PublishTime)
		data.Title = artInfo.Title
		var buf bytes.Buffer
		t, _ := template.New("article_list_brief").Parse(`
        <div class="article brief">
            <div class="candy-header"><span class="single">{{.PublishTime}}</span></div>
            <a href="/news?id={{.ID}}">{{.Title}}</a>
        </div>`)
		t.Execute(&buf, data)
		ret += buf.String()
	}

	return template.HTML(ret)
}

// RenderSyllabus dynamically renders syllabus page
func RenderSyllabus(semester int, courseNumber int) (template.HTML, string, error) {
	path := fmt.Sprintf("%s%d/%d.json", syllabusDataDir, semester, courseNumber)

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("handler/renderer.go RenderSyllabus() template error %s %v\n", path, err)
		return template.HTML(""), "", err
	}

	maps := map[string]interface{}{}
	json.Unmarshal(jsonData, &maps)

	courseName, _ := maps["Course_name_zh"].(string)

	for k, v := range maps {
		switch u := v.(type) {
		case string:
			maps[k] = template.HTML(u)
		}
	}

	t, err := template.ParseFiles(syllabusTemplate)
	if err != nil {
		log.Println("handler/renderer.go RenderSyllabus() template error " + err.Error())
		return template.HTML(""), "", err
	}

	var buf bytes.Buffer
	t.Execute(&buf, maps)
	return template.HTML(buf.String()), courseName, nil
}

// RenderCourseByYear statically renders course page by inputing year
func RenderCourseByYear(year uint) (template.HTML, error) {
	path := fmt.Sprintf("%s%d.json", courseDataDir, year)

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("handler/renderer.go RenderCourseByYear() template error %s %v\n", path, err)
		return template.HTML(""), err
	}

	// create a template
	unit := courseInfo{}
	units := struct {
		Title string
		Unit  []courseInfo
	}{}

	// e.g. 大一上學期、大一下學期...
	yearUnit := []map[string]interface{}{}

	// Get json data
	json.Unmarshal(jsonData, &yearUnit)

	// render title e.g. 109學年度課程內容
	units.Title = fmt.Sprintf("%d學年度課程內容", year)

	for _, s := range yearUnit {
		unit.Subtitle = "大" + convertLevel(int(s["level"].(float64))) +
			convertStrm(int(s["strm"].(float64))) + "學期"
		unit.Course = []map[string]interface{}{}
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
			unit.Course = append(unit.Course, info)
		}
		units.Unit = append(units.Unit, unit)
	}

	t, err := template.ParseFiles(courseTemplate)
	if err != nil {
		log.Println("handler/renderer.go RenderCourseByYear() template error " + err.Error())
		return template.HTML(""), err
	}

	var buf bytes.Buffer
	t.Execute(&buf, units)
	return template.HTML(buf.String()), nil
}

func RenderCalendarList(calendarList []Calendar, readOnly bool) template.HTML {
	dataList := []calendarRenderInfo{}
	for _, calendar := range calendarList {
		data := calendarRenderInfo{
			Calendar: calendar,
			HaveLink: calendar.Link != "",
			ReadOnly: readOnly,
		}
		dataList = append(dataList, data)
	}
	t, err := template.ParseFiles(calendarTemplate)
	if err != nil {
		log.Println("handler/renderer.go RenderCalendarList() template error " + err.Error())
		return template.HTML("")
	}

	var buf bytes.Buffer
	t.Execute(&buf, dataList)
	return template.HTML(buf.String())
}

func RenderIndexPage() template.HTML {
	t, err := template.ParseFiles(indexTemplate)
	if err != nil {
		log.Println("handler/renderer.go RenderIndexPage() template error " + err.Error())
		return template.HTML("")
	}

	// GetLatesetArticles() returns (list []Article, hasNext bool)
	artList, _ := GetLatesetArticles("public", "normal", "", 0, 7)
	// GetLatestCalendar() returns (list []Calendar, hasNext bool)
	calendarList, _ := GetLatestCalendar(0, 9)

	var buf bytes.Buffer
	t.Execute(&buf, struct {
		ArticleListBrief template.HTML
		CalendarList     template.HTML
	}{
		ArticleListBrief: RenderPublicArticleBriefList(artList),
		CalendarList:     RenderCalendarList(calendarList, true),
	})
	return template.HTML(buf.String())
}
