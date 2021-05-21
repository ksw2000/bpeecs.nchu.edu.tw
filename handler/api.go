package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"bpeecs.nchu.edu.tw/article"
	"bpeecs.nchu.edu.tw/files"
	"bpeecs.nchu.edu.tw/login"
)

type attachmentJSONStruct struct {
	ClientName string `json:"client_name"`
	Path       string `json:"path"`
	ServerName string `json:"server_name"`
}

func attachmentJSONtoClientName(attachmentJSON string) []string {
	attachment := []attachmentJSONStruct{}
	json.Unmarshal([]byte(attachmentJSON), &attachment)
	serverNameList := []string{}
	for _, v := range attachment {
		serverNameList = append(serverNameList, v.ServerName)
	}
	return serverNameList
}

// ApiHandler is a handler for handling url whose prefix is /function
func ApiHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path

	if path == "/api/login" {
		l := login.New()

		if err := l.Login(w, r); err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		fmt.Fprint(w, `{"err" : false}`)

		return
	} else if path == "/api/reg" {
		// only current accounts are allowed to register a new account
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "必需登入才能建立新帳戶(基於網路安全)"}`)
			return
		}

		l := login.New()

		id := get("id", r)
		pwd := get("pwd", r)
		rePwd := get("re_pwd", r)
		name := get("name", r)

		match, err := regexp.MatchString("^[a-zA-Z0-9_]{5,30}$", id)
		if err != nil || !match {
			fmt.Fprint(w, `{"err" : true , "msg" : "帳號僅接受「英文字母、數字、-、_」且需介於 5 到 30 字元`)
			return
		}

		if len(name) > 30 || len(name) < 1 {
			fmt.Fprint(w, `{"err" : true , "msg" : "暱稱需介於 1 到 30 字元"}`)
			return
		}

		if pwd != rePwd {
			fmt.Fprint(w, `{"err" : true , "msg" : "密碼與確認密碼不一致"}`)
			return
		}

		match, err = regexp.MatchString("^[a-zA-Z0-9_]{8,30}$", pwd)
		if err != nil || !match {
			fmt.Fprint(w, `{"err" : true , "msg" : "密碼僅接受「英文字母、數字、-、_」且需介於 8 到 30 字元"}`)
			return
		}

		match, err = regexp.MatchString("^.*?\\d+.*?$", pwd)
		if err != nil || !match {
			fmt.Fprint(w, `{"err" : true , "msg" : "密碼必需含有數字"}`)
			return
		}

		match, err = regexp.MatchString("^.*?[a-zA-Z]+.*?$", pwd)
		if err != nil || !match {
			fmt.Fprint(w, `{"err" : true , "msg" : "密碼必需含有英文字母"}`)
			return
		}

		err = l.NewAcount(id, pwd, name)
		if err == login.ErrorReapeatID {
			fmt.Fprint(w, `{"err" : true , "msg" : "所申請之 ID 重複"}`)
			return
		} else if err != nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "資料庫連結失敗", "code": 2}`)
			return
		} else {
			fmt.Fprint(w, `{"err" : false}`)
		}

		return
	} else if path == "/api/save_news" || path == "/api/publish_news" || path == "/api/del_news" {
		encoder := json.NewEncoder(w)
		ret := struct {
			Err string `json:"err"`
			Aid uint32 `json:"aid"`
		}{}

		// step0: is login?
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			ret.Err = "尚未登入"
			encoder.Encode(ret)
			return
		}

		// write to database
		// step1: fetch http POST
		getNum, err := strconv.Atoi(get("aid", r))
		if err != nil {
			ret.Err = "文章代碼錯誤 (POST參數錯誤)"
			encoder.Encode(ret)
			return
		}

		// step2: if num == -1
		// create new article and get aid
		art := article.New()
		if getNum == -1 {
			ret.Aid, err = art.NewArticle(loginInfo.UserID)
			if err != nil {
				ret.Err = err.Error()
				encoder.Encode(ret)
				return
			}
		} else {
			ret.Aid = uint32(getNum)
		}

		user := loginInfo.UserID
		artType := get("type", r)
		title := get("title", r)
		content := get("content", r)

		artFormat := article.Format{
			ID:      ret.Aid,
			User:    user,
			Type:    artType,
			Title:   title,
			Content: content,
		}

		serverNameList := attachmentJSONtoClientName(get("attachment", r))

		// step3: call Save() or Publish()
		if path == "/api/save_news" {
			err = art.Save(artFormat, serverNameList)
		} else if path == "/api/publish_news" {
			err = art.Publish(artFormat, serverNameList)
		} else if path == "/api/del_news" {
			err = art.Del(ret.Aid, user)
		}

		if err != nil {
			ret.Err = err.Error()
		}
		encoder.Encode(ret)
		return
	} else if path == "/api/get_news" {
		// read news from database
		// step1: read GET
		scope := get("scope", r)
		artType := get("type", r)
		n := get("id", r)
		var aid uint32
		from, to := 0, 19 // Default from = 0, to = 19

		scopes := [...]string{"all", "draft", "published", "public", "public-with-type"}
		checkValidScope := false
		for _, v := range scopes {
			if v == scope {
				checkValidScope = true
				break
			}
		}

		if !checkValidScope {
			if n == "" {
				fmt.Fprint(w, `{"err" : true , "msg" : "錯誤的請求 (GET 參數錯誤)", "code": 3}`)
				return
			}

			num, err := strconv.Atoi(n)
			if err != nil {
				fmt.Fprint(w, `{"err" : true , "msg" : "文章代碼錯誤 (GET 參數錯誤)", "code": 3}`)
				return
			}
			aid = uint32(num)
		} else {
			if f, scope := get("from", r), get("to", r); f != "" && scope != "" {
				var err error
				from, err = strconv.Atoi(f)
				to, err = strconv.Atoi(scope)
				if err != nil {
					fmt.Fprint(w, `{"err" : true , "msg" : "from to 代碼錯誤 (GET 參數錯誤)", "code": 3}`)
					return
				}
			}
		}

		// step2: some request need user id
		user := ""
		if loginInfo := login.CheckLogin(w, r); loginInfo != nil {
			user = loginInfo.UserID
		}

		// step3: connect to database
		art := article.New()

		// step4: call GetLatest(scope, from, to)
		if scope != "" {
			ret := new(struct {
				List    []article.Format `json:"list"`
				HasNext bool             `json:"hasNext"`
			})
			ret.List, ret.HasNext = art.GetLatest(scope, artType, user, int32(from), int32(to))

			// step5: encode to json
			json.NewEncoder(w).Encode(ret)
		} else if n != "" {
			if ret := art.GetArticleByAid(aid, user); ret != nil {
				json.NewEncoder(w).Encode(ret)
			} else {
				fmt.Fprint(w, `{}`)
			}
		}
	} else if path == "/api/upload" {
		// is login？
		if login.CheckLogin(w, r) == nil {
			fmt.Fprint(w, `{"Err" : true , "Msg" : "尚未登入", "Code" : 1}`)
			return
		}

		r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
		fhs := r.MultipartForm.File["files"]

		ret := []struct {
			FileName string `json:"fileName"`
			FilePath string `json:"filePath"`
		}{}

		for _, fh := range fhs {
			f := files.New()

			if err := f.NewFile(fh); err != nil {
				fmt.Fprint(w, `{"err" : true , "msg" : "新增檔案失敗", "code": 4}`)
				return
			}
			ret = append(ret, struct {
				FileName string `json:"fileName"`
				FilePath string `json:"filePath"`
			}{
				FileName: f.ServerName,
				FilePath: f.Path,
			})
		}
		json.NewEncoder(w).Encode(ret)
	} else if path == "/api/del_attachment" {
		// is login？
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "尚未登入", "code" : 1}`)
			return
		}

		serverName := get("server_name", r)
		aidNum := get("aid_num", r)
		num, err := strconv.Atoi(aidNum)
		if err != nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "文章代碼錯誤 (GET 參數錯誤)", "code": 3}`)
			return
		}

		serverNameList := attachmentJSONtoClientName(get("new_attachment", r))

		// Delete file record in database and delete file in system
		f := files.New()
		if err := f.Del(serverName); err != nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "檔案資料庫連結失敗或檔案刪除失敗", "code": 2}`)
			return
		}

		// Update databse article (prevent user from not storing the article)
		art := article.New()
		if err := art.UpdateAttachment(uint32(num), serverNameList); err != nil {
			fmt.Fprint(w, `{"err" : true , "msg" : "Article 資料庫更新失敗", "code": 2}`)
			return
		}

		fmt.Fprint(w, `{"err" : false}`)
	} else {
		fmt.Println("未預期的路徑" + path)
		http.Redirect(w, r, "/error/404", 302)
	}
}

func get(key string, r *http.Request) string {
	return strings.Join(r.Form[key], "")
}
