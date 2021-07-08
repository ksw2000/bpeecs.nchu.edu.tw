package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"bpeecs.nchu.edu.tw/config"
	"github.com/go-session/session"

	_ "github.com/mattn/go-sqlite3"
)

// Login handles manipulations about login
type Login struct {
	IsLogin  bool
	UserID   string
	UserName string
}

// ErrorReapeatID is returned when the user want to sign up an account which has already existed
var ErrorReapeatID error

// New returns new instance of Login
func NewLogin() (l *Login) {
	l = new(Login)
	ErrorReapeatID = errors.New("ID-Repeat")
	return
}

// Login is a function handle login
func (l *Login) Login(w http.ResponseWriter, r *http.Request) error {
	id, pwd := r.FormValue("id"), r.FormValue("pwd")
	log.Printf("%s try to login\n", id)

	d, err := sql.Open("sqlite3", config.MainDB)
	if err != nil {
		log.Println(err)
		return err
	}
	defer d.Close()
	row := d.QueryRow("SELECT `password`, `name`, `salt` FROM user WHERE `id` = ?", id)

	var enryptedPwd, name, salt string
	err = row.Scan(&enryptedPwd, &name, &salt)

	if err == sql.ErrNoRows || pwdHash(pwd, salt) != enryptedPwd {
		return fmt.Errorf("帳號或密碼錯誤")
	}

	l.IsLogin = true
	l.UserID = id
	l.UserName = name

	// Session srart
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		return fmt.Errorf("session.Start() error %v", err)
	}

	store.Set("isLogin", "yes")
	store.Set("userID", l.UserID)
	store.Set("userName", l.UserName)

	if err = store.Save(); err != nil {
		return fmt.Errorf("Session store error")
	}

	log.Printf("%s login success\n", id)
	return nil
}

// NewAcount creates a new account
func (l *Login) NewAcount(id string, pwd string, name string) error {
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

	return ErrorReapeatID
}

// Logout is a function deal with sign out
func (l *Login) Logout(w http.ResponseWriter, r *http.Request) (err error) {
	store, err := session.Start(context.Background(), w, r)
	store.Set("isLogin", "no")
	store.Set("userID", "")
	store.Set("userName", "")
	return err
}

// CheckLogin checks if ID and Password is match
func CheckLogin(w http.ResponseWriter, r *http.Request) *Login {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		return nil
	}

	isLogin, ok1 := store.Get("isLogin")
	userID, ok2 := store.Get("userID")
	userName, ok3 := store.Get("userName")

	if !(ok1 && ok2 && ok3) {
		return nil
	}

	if isLogin != "yes" {
		return nil
	}

	return &Login{
		IsLogin:  true,
		UserID:   userID.(string),
		UserName: userName.(string),
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
