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

	"bpeecs.nchu.edu.tw/article"
	"bpeecs.nchu.edu.tw/login"
	"bpeecs.nchu.edu.tw/renderer"
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

	// Handle static file
	staticFiles := []string{"/robot.txt", "/sitemap.xml", "/favicon.ico"}
	for _, f := range staticFiles {
		if r.URL.Path == f {
			//http.Redirect(w, r, "/assets/img/favicon.ico", 301)
			http.StripPrefix("/", http.FileServer(http.Dir("./"))).ServeHTTP(w, r)
			return
		}
	}

	// Handle simple web and check login info
	data := initPageData()

	loginInfo := login.CheckLogin(w, r)
	data.IsLogin = loginInfo != nil

	var simpleWeb = map[string]string{
		"/about/education-goal-and-core-ability": "教育目標及核心能力",
		"/about/enrollment":                      "招生方式",
		"/about/feature":                         "特色",
		"/about/future-development-direction":    "學生未來發展方向",
		"/about/official-document":               "系務相關辦法",
		"/about/why-establish":                   "創系緣由",
		"/course":                                "課程內容",
		"/course/graduation-conditions":          "畢業條件",
		"/course/109":                            "109學年度課程內容",
		"/member/admin-staff":                    "行政人員",
		"/member/faculty":                        "師資陣容",
		"/member/class-teacher":                  "班主任",
		"/syllabus":                              "課程大綱",
	}

	// Handle non simple web
	if title, ok := simpleWeb[r.URL.Path]; !ok {
		data.Title = title
		switch r.URL.Path {
		case "/":
			data.Title = "國立中興大學電機資訊學院學士班"
			data.Isindex = true

			t, _ := template.ParseFiles("./include/index.html")
			art := article.New()
			// Default from = 0, to = 19
			// return (list []art.Format, hasNext bool)
			artFormatList, _ := art.GetLatest("public", "normal", "", int32(0), int32(7))
			data2 := new(struct {
				ArticleListBrief template.HTML
			})
			data2.ArticleListBrief = renderer.RenderPublicArticleBriefList(artFormatList)

			var buf bytes.Buffer
			t.Execute(&buf, data2)
			data.Main = template.HTML(buf.String())
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
				//id is uint32
				aidUint64, err := strconv.ParseUint(id, 10, 32)

				if err != nil {
					http.Redirect(w, r, "/error/404", 302)
					return
				}
				art := article.New()
				user := ""
				if data.IsLogin {
					user = loginInfo.UserID
				}

				artInfo := art.GetArticleByAid(uint32(aidUint64), user)

				// avoid /news?id=xxx
				if artInfo == nil {
					http.Redirect(w, r, "/error/404", 302)
					return
				}

				data.Title = artInfo.Title + " | 國立中興大學電機資訊學院學士班"
				data.Main = renderer.RenderPublicArticle(artInfo)
			} else {
				data.Title += " | 國立中興大學電機資訊學院學士班"
				data.Main, _ = getHTML(r.URL.Path)
			}
		case "/login":
			if login.CheckLogin(w, r) != nil {
				http.Redirect(w, r, "/manage", 302)
				return
			}
			data.Title = "登入"
		case "/logout":
			ret := struct {
				Err string `json:"err"`
			}{}
			if err := login.New().Logout(w, r); err != nil {
				ret.Err = "登出失敗，重試，或清除 cookie"
				json.NewEncoder(w).Encode(w)
				return
			}

			http.Redirect(w, r, "/", 302)
			return
		default:
			log.Printf("Error path %s IP: %s\n", r.URL.Path, r.RemoteAddr)
			http.Redirect(w, r, "/error/404", 302)
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
