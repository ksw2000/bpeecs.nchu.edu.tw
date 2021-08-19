package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// PageData is a type for filling HTML template
type PageData struct {
	Title   string
	Isindex bool
	IsLogin bool
	Main    template.HTML
	Time    int64
	Year    int
}

func initPageData() *PageData {
	data := new(PageData)
	data.Isindex = false // default value
	data.Time = time.Now().Unix() >> 10
	data.Year, _, _ = time.Now().Date()
	return data
}

func getHTML(fileName string) (template.HTML, error) {
	t, err := template.ParseFiles("./include" + fileName + ".html")
	if err != nil {
		log.Println(err)
		return template.HTML("error try again"), err
	}

	var buf bytes.Buffer
	t.Execute(&buf, nil)
	return template.HTML(buf.String()), nil
}

// BasicWebHandler is a handler for handling url whose prefix is /
func BasicWebHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// Handle static file
	staticFiles := []string{"/robot.txt", "/sitemap.xml", "/favicon.ico"}
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

	var simpleWeb = map[string]string{
		"/about/education-goal-and-core-ability": "教育目標及核心能力",
		"/about/enrollment":                      "招生方式",
		"/about/feature":                         "特色",
		"/about/future-development-direction":    "學生未來發展方向",
		"/about/official-document":               "系務相關辦法",
		"/about/why-establish":                   "創系緣由",
		"/course":                                "課程內容",
		"/course/graduation-conditions":          "畢業條件及輔系雙主修",
		"/member/admin-staff":                    "行政人員",
		"/member/faculty":                        "師資陣容",
		"/member/class-teacher":                  "班主任",
		"/syllabus":                              "課程大綱",
	}

	// Handle simple web
	if title, ok := simpleWeb[r.URL.Path]; ok {
		data.Title = title
	} else {
		// Handle non simple web
		switch r.URL.Path {
		case "/":
			data.Title = "國立中興大學電機資訊學院學士班"
			data.Isindex = true
			data.Main = RenderIndexPage()
		case "/news":
			data.Title = "最新消息"
			artType := strings.Join(r.Form["type"], "")
			var dict = map[string]string{
				"normal":       "一般消息",
				"activity":     "演講 & 活動",
				"course":       "課程 & 招生",
				"scholarships": "獎學金訊息",
				"recruit":      "徵才資訊",
			}

			if subtitle, ok := dict[artType]; ok {
				data.Title = subtitle + " | " + data.Title
			}

			if id := strings.Join(r.Form["id"], ""); id != "" {
				aid, err := strconv.ParseInt(id, 10, 64)

				if err != nil {
					NotFound(w, r)
					return
				}

				uid := ""
				if data.IsLogin {
					uid = user.ID
				}

				artInfo := GetArticleByAid(aid, uid)

				// avoid /news?id=xxx
				if artInfo == nil {
					NotFound(w, r)
					return
				}

				data.Title = artInfo.Title + " | 國立中興大學電機資訊學院學士班"
				data.Main = RenderPublicArticle(artInfo)
			} else {
				data.Title += " | 國立中興大學電機資訊學院學士班"
				data.Main, _ = getHTML(r.URL.Path)
			}
		case "/login":
			if CheckLoginBySession(w, r) != nil {
				http.Redirect(w, r, "/manage", 302)
				return
			}
			data.Title = "登入"
		case "/logout":
			ret := struct {
				Err string `json:"err"`
			}{}
			if err := Logout(w, r); err != nil {
				ret.Err = "登出失敗，重試，或清除 cookie"
				json.NewEncoder(w).Encode(w)
				return
			}

			http.Redirect(w, r, "/", 302)
			return
		default:
			NotFound(w, r)
			return
		}
	}

	if r.URL.Path != "/" && r.URL.Path != "/news" {
		data.Title += " | 國立中興大學電機資訊學院學士班"
		data.Main, _ = getHTML(r.URL.Path)
	}

	t, _ := template.ParseFiles("./include/layout.gohtml")
	t.Execute(w, data)
}
