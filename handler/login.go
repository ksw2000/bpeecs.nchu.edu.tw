package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"bpeecs.nchu.edu.tw/config"
	"github.com/go-session/session"

	_ "github.com/mattn/go-sqlite3"
)

// Login handles manipulations about login
type User struct {
	ID   string
	Name string
}

// Login is a function handle login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ret := struct {
		Err string `json:"err"`
	}{}
	encoder := json.NewEncoder(w)

	id, pwd := r.FormValue("id"), r.FormValue("pwd")
	log.Printf("%s try to login\n", id)

	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		ret.Err = err.Error()
		encoder.Encode(ret)
		return
	}
	defer d.Close()
	row := d.QueryRow("SELECT `password`, `name`, `salt` FROM user WHERE `id` = ?", id)

	var enryptedPwd, name, salt string
	err = row.Scan(&enryptedPwd, &name, &salt)

	if err == sql.ErrNoRows || pwdHash(pwd, salt) != enryptedPwd {
		ret.Err = "帳號或密碼錯誤"
		encoder.Encode(ret)
		return
	}

	// Session srart
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		ret.Err = fmt.Sprintf("session.Start() error %v", err)
		encoder.Encode(ret)
		return
	}

	store.Set("userID", id)
	store.Set("userName", name)

	if err = store.Save(); err != nil {
		ret.Err = "Session store error"
		encoder.Encode(ret)
		return
	}

	log.Printf("%s login success\n", id)
	encoder.Encode(ret)
	return
}

func RegHandler(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	ret := struct {
		Err string `json:"err"`
	}{}

	// only current accounts are allowed to register a new account
	user := CheckLoginBySession(w, r)
	if user == nil {
		ret.Err = "必需登入才能建立新帳戶"
		encoder.Encode(ret)
		return
	}

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

	if err := NewAcount(id, pwd, name); err != nil {
		ret.Err = err.Error()
		encoder.Encode(ret)
		return
	}

	encoder.Encode(ret)
	return
}

// NewAcount creates a new account
func NewAcount(id string, pwd string, name string) error {
	// check if there are the same id in db
	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		return fmt.Errorf("sql.Open() error %v", err)
	}
	defer d.Close()

	row := d.QueryRow("SELECT COUNT(*) FROM user WHERE `id` = ?", id)

	count := 0
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("row.Scan() error %v", err)
	}

	// Check account
	if count == 0 {
		salt := randomString(64)
		pwd = pwdHash(pwd, salt)

		stmt, err := d.Prepare("INSERT INTO user(id, password, salt, name) values(?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("d.Prepare() error %v", err)
		}

		stmt.Exec(id, pwd, salt, name)
		return nil
	}

	return fmt.Errorf("所申請之 ID 已重複")
}

// Logout is a function deal with sign out
func Logout(w http.ResponseWriter, r *http.Request) (err error) {
	store, err := session.Start(context.Background(), w, r)
	store.Delete("userID")
	store.Delete("userName")
	return err
}

// CheckLoginBySession checks if ID and Password is match
func CheckLoginBySession(w http.ResponseWriter, r *http.Request) *User {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		return nil
	}

	userID, ok2 := store.Get("userID")
	userName, ok3 := store.Get("userName")

	if !(ok2 && ok3) {
		return nil
	}

	return &User{
		ID:   userID.(string),
		Name: userName.(string),
	}
}

func pwdHash(pwd string, salt string) string {
	pwd += salt
	return fmt.Sprintf("%x", sha256.Sum256([]byte(pwd)))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~@!"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
