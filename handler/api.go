package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func attachmentJSONtoClientName(attachmentJSON string) []string {
	attachment := []struct {
		ClientName string `json:"client_name"`
		Path       string `json:"path"`
		ServerName string `json:"server_name"`
	}{}
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
	encoder := json.NewEncoder(w)

	if path == "/api/login" {
		LoginHandler(w, r)
	} else if path == "/api/reg" {
		RegHandler(w, r)
	} else if path == "/api/news" {
		save := exist("save", r)
		publish := exist("publish", r)
		del := exist("del", r)

		ret := struct {
			Err         string `json:"err"`
			ErrNotLogin bool   `json:"errNotLogin"`
			Aid         int64  `json:"aid"`
		}{}

		// step0: check login
		user := CheckLoginBySession(w, r)
		if user == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			encoder.Encode(ret)
			return
		}

		// write to database
		// step1: fetch http POST
		getNum, err := strconv.ParseInt(get("aid", r), 10, 64)
		if err != nil {
			ret.Err = "文章代碼錯誤 (POST參數錯誤)"
			encoder.Encode(ret)
			return
		}

		// step2: if num == -1
		// create new article and get aid
		art := new(Article)
		if save || publish {
			if getNum == -1 {
				err = art.Create(user.ID)
				ret.Aid = art.ID
				if err != nil {
					ret.Err = err.Error()
					encoder.Encode(ret)
					return
				}
			} else {
				ret.Aid = getNum
			}

			art = &Article{
				ID:      ret.Aid,
				User:    user.ID,
				Type:    get("type", r),
				Title:   get("title", r),
				Content: get("content", r),
			}

			serverNameList := attachmentJSONtoClientName(get("attachment", r))
			if save {
				err = art.Save(serverNameList)
			} else if publish {
				err = art.Publish(serverNameList)
			}
		} else if del {
			art = &Article{
				ID: getNum,
			}
			// delete
			art.Del(user.ID)
		}

		if err != nil {
			ret.Err = err.Error()
		}
		encoder.Encode(ret)
		return
	} else if path == "/api/get_news" {
		// step0: prepare return format
		ret := new(struct {
			Err     string    `json:"err"`
			List    []Article `json:"list"`
			HasNext bool      `json:"hasNext"`
		})

		// step1: parse GET paramters
		scope := get("scope", r)
		artType := get("type", r)
		aidStr := get("id", r)
		from, to := 0, 19 // Default from = 0, to = 19
		scopes := [...]string{"all", "draft", "published", "public", "public-with-type"}
		checkValidScope := false
		var aid int64
		var err error
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

			aid, err = strconv.ParseInt(aidStr, 10, 64)
			if err != nil {
				ret.Err = fmt.Sprintf("Invalid request. %v", err)
				encoder.Encode(ret)
				return
			}
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
		uid := ""
		if user := CheckLoginBySession(w, r); user != nil {
			uid = user.ID
		}

		// step3: call GetLatesetArticles(scope, from, to)
		if scope != "" {
			ret.List, ret.HasNext = GetLatesetArticles(scope, artType, uid, from, to)
			encoder.Encode(ret)
			return
		}

		if aidStr != "" {
			if ret := GetArticleByAid(aid, uid); ret != nil {
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

		// check login
		if CheckLoginBySession(w, r) == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			json.NewEncoder(w).Encode(ret)
			return
		}

		r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
		fhs := r.MultipartForm.File["files"]
		ret.FileList = []fileInfo{}
		for _, fh := range fhs {
			f, err := NewFile(fh)
			if err != nil {
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
		return
	} else if path == "/api/del_attachment" {
		ret := struct {
			Err         string `json:"err"`
			ErrNotLogin bool   `json:"errNotLogin"`
		}{}

		// check login
		user := CheckLoginBySession(w, r)
		if user == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			encoder.Encode(ret)
			return
		}

		serverName := get("server_name", r)
		aidStr := get("aid_num", r)
		aid, err := strconv.ParseInt(aidStr, 10, 64)
		if err != nil {
			ret.Err = "@param aid_num is invalid"
			encoder.Encode(ret)
			return
		}

		serverNameList := attachmentJSONtoClientName(get("new_attachment", r))

		// Delete file record in database and delete file in system
		f := Files{ServerName: serverName}
		if err := f.Del(); err != nil {
			ret.Err = fmt.Sprintf("檔案資料庫連結失敗或檔案刪除失敗 %v", err)
			encoder.Encode(ret)
			return
		}

		// Update databse article (prevent user from not storing the article)
		art := Article{ID: aid}
		if err := art.UpdateAttachment(serverNameList); err != nil {
			ret.Err = fmt.Sprintf("資料庫更新失敗 art.UpdateAttachment() %v", err)
			encoder.Encode(ret)
			return
		}
		encoder.Encode(ret)
		return
	} else if path == "/api/calendar" {
		ret := struct {
			Err         string `json:"err"`
			ErrNotLogin bool   `json:"errNotLogin"`
		}{}

		// check login
		user := CheckLoginBySession(w, r)
		if user == nil {
			ret.Err = "權限不足，尚未登入"
			ret.ErrNotLogin = true
			encoder.Encode(ret)
			return
		}

		// step1: parse action type (add, edit, del)
		add := exist("add", r)
		edit := exist("edit", r)
		del := exist("del", r)

		// step2: parse GET or POST param
		var date, event, link string
		var year, month, day uint64
		var id int64
		var err, err2, err3 error

		if add || edit {
			date = get("date", r)
			event = get("event", r)
			link = get("link", r)

			if date == "" || event == "" {
				ret.Err = "日期及事件標題不可為空"
				encoder.Encode(ret)
				return
			}

			dateParts := strings.Split(date, "-")
			if len(dateParts) != 3 {
				ret.Err = "日期格式錯誤"
				encoder.Encode(ret)
				return
			}

			year, err = strconv.ParseUint(dateParts[0], 10, 12)
			month, err2 = strconv.ParseUint(dateParts[1], 10, 4)
			day, err3 = strconv.ParseUint(dateParts[2], 10, 5)

			if err != nil || err2 != nil || err3 != nil {
				ret.Err = "日期格式錯誤"
				encoder.Encode(ret)
				return
			}
		}

		if edit || del {
			id, err = strconv.ParseInt(get("id", r), 10, 64)
			if err != nil {
				ret.Err = "ID 錯誤"
				encoder.Encode(ret)
				return
			}
		}

		cal := Calendar{
			ID:    id,
			Year:  uint(year),
			Month: uint(month),
			Day:   uint(day),
			Event: event,
			Link:  link,
		}

		if add {
			if err := cal.Add(); err != nil {
				ret.Err = err.Error()
			}
		} else if edit {
			if err := cal.Update(); err != nil {
				ret.Err = err.Error()
			}
		} else if del {
			if err := cal.Del(); err != nil {
				ret.Err = err.Error()
			}
		}
		encoder.Encode(ret)
		return
	} else if path == "/api/get_calendar" {
		// 本 API 僅供行事曆檢視，不檢查是否登入
		ret := struct {
			InfoList []Calendar `json:"infoList"`
			Err      string     `json:"err"`
		}{}

		year, err1 := strconv.ParseUint(get("year", r), 10, 12)
		month, err2 := strconv.ParseUint(get("month", r), 10, 8)

		if err1 != nil || err2 != nil {
			log.Println(err1)
			log.Println(err2)
			ret.Err = "年份、月份格式錯誤"
			encoder.Encode(ret)
			return
		}

		ret.InfoList = GetCalendarByYearMonth(uint(year), uint(month))
		encoder.Encode(ret)
		return
	} else {
		NotFound(w, r)
	}
}

func get(key string, r *http.Request) string {
	return strings.Join(r.Form[key], "")
}

func exist(key string, r *http.Request) bool {
	_, ok := r.Form[key]
	return ok
}
