package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"server/components"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuth struct {
	AuthCode string
	Config   oauth2.Config
}

type App struct {
	DB         *sql.DB
	GoogleAuth GoogleAuth
}

type User struct {
	Id      int
	Email   string
	Picture string
}

type GoogleUser struct {
	Id      string
	Email   string
	Picture string
}

func (app *App) Routes() *http.ServeMux {
	router := http.NewServeMux()

	static_fs := http.FileServer(http.Dir("./include_dir/"))
	assets_fs := http.FileServer(http.Dir("./assets/"))

	router.Handle("/", static_fs)
	router.Handle("/audio_hls", assets_fs)
	router.Handle("POST /clicked", templ.Handler(components.Hello()))

	router.HandleFunc("POST /audio/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("./assets/coverless-book-lofi-186307.mp3")
		stats, err := file.Stat()

		n, err := io.CopyN(w, file, stats.Size())
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("written %d bytes over network \n", n)
	})

	router.HandleFunc("/login", app.GoogleAuth.LoginGoogle)
	router.HandleFunc("/callback", app.GoogleCallback)

	return router
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger(next, w, r)
		Auth(next, w, r)
	})
}

func Logger(next http.Handler, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	next.ServeHTTP(w, r)
	log.Println(r.Method, r.URL.Path, time.Since(start))
}

func Auth(next http.Handler, w http.ResponseWriter, r *http.Request) {
}

func (auth *GoogleAuth) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	url := auth.Config.AuthCodeURL(auth.AuthCode)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *App) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := app.GoogleAuth.GetUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Println(string(content))
	var login_data GoogleUser
	err = json.Unmarshal(content, &login_data)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	log.Println(login_data)
	stmt := `INSERT INTO users (id, email, picture) VALUES (?, ?, ?)`
	_, err = app.DB.Exec(stmt, login_data.Id, login_data.Email, login_data.Picture)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (auth *GoogleAuth) GetUserInfo(state string, code string) ([]byte, error) {
	if state != auth.AuthCode {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := auth.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}

func GenStrin(length int) string {
	letter_runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	str := make([]rune, length)
	for i := range str {
		str[i] = letter_runes[rand.Intn(len(letter_runes))]
	}

	return string(length)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	google_auth := GoogleAuth{
		AuthCode: GenStrin(8),
		Config: oauth2.Config{
			RedirectURL:  "http://localhost:8080/callback",
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		}}

	log.Println(google_auth.Config)

	log.Println("Listening on :8080")
	app := App{
		DB:         db,
		GoogleAuth: google_auth,
	}

	_ = app.DB

	server := http.Server{
		Addr:    ":8080",
		Handler: Middleware(app.Routes()),
	}

	server.ListenAndServe()
}
