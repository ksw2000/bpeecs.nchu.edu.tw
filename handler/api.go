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
		ret := struct {
			Err string `json:"err"`
		}{}
		l := login.New()

		if err := l.Login(w, r); err != nil {
			ret.Err = err.Error()
			json.NewEncoder(w).Encode(ret)
			return
		}

		json.NewEncoder(w).Encode(ret)
		return
	} else if path == "/api/reg" {
		ret := struct {
			Err string `json:"err"`
		}{}
		encoder := json.NewEncoder(w)
		// only current accounts are allowed to register a new account
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			ret.Err = "必需登入才能建立新帳戶"
			encoder.Encode(ret)
			return
		}

		l := login.New()

		id := get("id", r)
		pwd := get("pwd", r)
		rePwd := get("re_pwd", r)
		name := get("name", r)

		match, err := regexp.MatchString("^[a-zA-Z0-9_]{5,30}$", id)
		if err != nil || !match {
			ret.Err = "帳號僅接受「英文字母、數字、-、_」且需介於 5 到 30 字元"
			encoder.Encode(ret)
			return
		}

		if len(name) > 30 || len(name) < 1 {
			ret.Err = "暱稱需介於 1 到 30 字元"
			encoder.Encode(ret)
			return
		}

		if pwd != rePwd {
			ret.Err = "密碼與確認密碼不一致"
			encoder.Encode(ret)
			return
		}

		match, err = regexp.MatchString("^[a-zA-Z0-9_]{8,30}$", pwd)
		if err != nil || !match {
			ret.Err = "密碼僅接受「英文字母、數字、-、_」且需介於 8 到 30 字元"
			encoder.Encode(ret)
			return
		}

		match, err = regexp.MatchString("^.*?\\d+.*?$", pwd)
		if err != nil || !match {
			ret.Err = "密碼必需含有數字"
			encoder.Encode(ret)
			return
		}

		match, err = regexp.MatchString("^.*?[a-zA-Z]+.*?$", pwd)
		if err != nil || !match {
			ret.Err = "密碼必需含有英文字母"
			encoder.Encode(ret)
			return
		}

		err = l.NewAcount(id, pwd, name)
		if err == login.ErrorReapeatID {
			ret.Err = "所申請之 ID 重複"
			encoder.Encode(ret)
			return
		} else if err != nil {
			ret.Err = fmt.Sprintf("資料庫錯誤 %v", err)
			encoder.Encode(ret)
			return
		}

		encoder.Encode(ret)
		return
	} else if path == "/api/save_news" || path == "/api/publish_news" || path == "/api/del_news" {
		encoder := json.NewEncoder(w)
		ret := struct {
			Err         string `json:"err"`
			ErrNotLogin bool   `json:"errNotLogin"`
			Aid         uint32 `json:"aid"`
		}{}

		// step0: is login?
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
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
		// step0: prepare return format
		ret := new(struct {
			Err     string           `json:"err"`
			List    []article.Format `json:"list"`
			HasNext bool             `json:"hasNext"`
		})
		encoder := json.NewEncoder(w)

		// step1: parse GET paramters
		scope := get("scope", r)
		artType := get("type", r)
		aidStr := get("id", r)
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
			if aidStr == "" {
				ret.Err = "Invalid request @param id can not be empty"
				encoder.Encode(ret)
				return
			}

			aidInt, err := strconv.Atoi(aidStr)
			if err != nil {
				ret.Err = fmt.Sprintf("Invalid request. %v", err)
				encoder.Encode(ret)
				return
			}
			aid = uint32(aidInt)
		} else {
			if f, scope := get("from", r), get("to", r); f != "" && scope != "" {
				var err error
				from, err = strconv.Atoi(f)
				to, err = strconv.Atoi(scope)
				if err != nil {
					ret.Err = "Invalid request @param from and @param to"
					encoder.Encode(ret)
					return
				}
			}
		}

		// step2: some request need user id
		user := ""
		if loginInfo := login.CheckLogin(w, r); loginInfo != nil {
			user = loginInfo.UserID
		}

		// step3: call GetLatest(scope, from, to)
		art := article.New()
		if scope != "" {
			ret.List, ret.HasNext = art.GetLatest(scope, artType, user, int32(from), int32(to))
			encoder.Encode(ret)
			return
		}

		if aidStr != "" {
			if ret := art.GetArticleByAid(aid, user); ret != nil {
				encoder.Encode(ret)
				return
			}
		}

		ret.Err = "Inavalid request"
		encoder.Encode(ret)
		return
	} else if path == "/api/upload" {
		type fileInfo struct {
			FileName string `json:"fileName"`
			FilePath string `json:"filePath"`
		}
		ret := struct {
			Err         string     `json:"err"`
			ErrNotLogin bool       `json:"errNotLogin"`
			FileList    []fileInfo `json:"fileList"`
		}{}

		// is login?
		if login.CheckLogin(w, r) == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			json.NewEncoder(w).Encode(ret)
			return
		}

		r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
		fhs := r.MultipartForm.File["files"]
		ret.FileList = []fileInfo{}
		for _, fh := range fhs {
			f := files.New()

			if err := f.NewFile(fh); err != nil {
				ret.Err = "新增檔案失敗"
				json.NewEncoder(w).Encode(ret)
				return
			}

			ret.FileList = append(ret.FileList, fileInfo{
				FileName: f.ServerName,
				FilePath: f.Path,
			})
		}
		json.NewEncoder(w).Encode(ret)
	} else if path == "/api/del_attachment" {
		ret := struct {
			Err         string `json:"err"`
			ErrNotLogin bool   `json:"errNotLogin"`
		}{}
		encoder := json.NewEncoder(w)
		// is login?
		loginInfo := login.CheckLogin(w, r)
		if loginInfo == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			encoder.Encode(ret)
			return
		}

		serverName := get("server_name", r)
		aidStr := get("aid_num", r)
		aidInt, err := strconv.Atoi(aidStr)
		if err != nil {
			ret.Err = "@param aid_num is invalid"
			encoder.Encode(ret)
			return
		}

		serverNameList := attachmentJSONtoClientName(get("new_attachment", r))

		// Delete file record in database and delete file in system
		f := files.New()
		if err := f.Del(serverName); err != nil {
			ret.Err = fmt.Sprintf("檔案資料庫連結失敗或檔案刪除失敗 %v", err)
			encoder.Encode(ret)
			return
		}

		// Update databse article (prevent user from not storing the article)
		art := article.New()
		if err := art.UpdateAttachment(uint32(aidInt), serverNameList); err != nil {
			ret.Err = fmt.Sprintf("資料庫更新失敗 art.UpdateAttachment() %v", err)
			encoder.Encode(ret)
			return
		}
		encoder.Encode(ret)
		return
	}
	NotFound(w, r)
}

func get(key string, r *http.Request) string {
	return strings.Join(r.Form[key], "")
}
