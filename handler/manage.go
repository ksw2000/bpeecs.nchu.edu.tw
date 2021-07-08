package handler

import (
	"bytes"
	"html/template"
	"net/http"
)

// ManageWebHandler is a handler for handling url whose prefix is /manage
func ManageWebHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path

	data := initPageData()

	// Is login?
	loginInfo := CheckLogin(w, r)
	data.IsLogin = (loginInfo != nil)

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

	if !data.IsLogin {
		http.Redirect(w, r, "/?notlogin", 302)
		return
	}

	if path == "/manage/" {
		manageTemplate, _ := template.ParseFiles("./include/manage.gohtml")
		var manageTemplateByte bytes.Buffer
		manageTemplateData := struct {
			UserID   string
			UserName string
		}{
			UserID:   loginInfo.UserID,
			UserName: loginInfo.UserName,
		}
		manageTemplate.Execute(&manageTemplateByte, manageTemplateData)
		data.Main = template.HTML(manageTemplateByte.String())
	} else {
		data.Main, _ = getHTML(path)
	}

	data.Title += " | 國立中興大學電機資訊學院學士班"

	// TEMPLATE
	t, _ := template.ParseFiles("./include/layout.gohtml")
	t.Execute(w, data)
}
