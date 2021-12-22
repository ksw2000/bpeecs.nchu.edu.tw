package handler

import (
	"html/template"
	"net/http"
)

// ManageWebHandler is a handler for handling url whose prefix is /manage
func ManageWebHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path

	data := initPageData()

	// Is login?
	user := CheckLoginBySession(w, r)
	data.IsLogin = user != nil

	var manageWeb = map[string]string{
		"/manage/":         "歡迎進入後台管理系統",
		"/manage/article":  "文章管理",
		"/manage/calendar": "行事曆管理",
		"/manage/reg":      "註冊新用戶",
		"/manage/reg-done": "新用戶註冊成功",
	}

	var ok bool
	data.Title, ok = manageWeb[path]

	if !ok {
		NotFound(w, r)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/?notlogin", 302)
		return
	} else if path == "/manage/" {
		data.Main = RenderMangePage(user)
	} else {
		data.Main, _ = getHTML(path)
	}

	data.Title += " | 國立中興大學電機資訊學院學士班"

	// TEMPLATE
	t, _ := template.ParseFiles("./html/layout.gohtml")
	t.Execute(w, data)
}
