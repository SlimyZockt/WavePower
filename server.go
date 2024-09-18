package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"server/components"
	"time"

	"github.com/a-h/templ"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(r.Method, r.URL.Path, time.Since(start))
	})
}

var auth_config oauth2.Config

var auth_code string

func LoginGoogle(w http.ResponseWriter, r *http.Request) {
	url := auth_config.AuthCodeURL(auth_code)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Content: %s\n", content)
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != auth_code {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := auth_config.Exchange(context.Background(), code)
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

	auth_code = GenStrin(8)

	auth_config = oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	router := http.NewServeMux()
	static_dir := http.Dir("./include_dir/")
	static_fs := http.FileServer(static_dir)

	assets_dir := http.Dir("./assets/")
	assets_fs := http.FileServer(assets_dir)

	router.Handle("/", static_fs)
	router.Handle("/audio_hls", assets_fs)
	router.Handle("POST /clicked", templ.Handler(components.Hello()))

	router.HandleFunc("POST /audio/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("./assets/coverless-book-lofi-186307.mp3")
		stats, err := file.Stat()

		// binary.Write(w, binary.LittleEndian, stats.Size())
		n, err := io.CopyN(w, file, stats.Size())
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("written %d bytes over network \n", n)
	})

	router.HandleFunc("/login", LoginGoogle)
	router.HandleFunc("/callback", GoogleCallback)

	log.Println("Listening on :8080")

	server := http.Server{
		Addr:    ":8080",
		Handler: Logger(router),
	}

	server.ListenAndServeTLS("./certificate.crt", "./privatekey.key")
}
