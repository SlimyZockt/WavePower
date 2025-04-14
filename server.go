package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"server/internal/middleware"
	"server/internal/routes"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func GenStrin(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	str := make([]rune, length)
	for i := range str {
		str[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(str)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}
	app := routes.App{
		AuthCode: GenStrin(16),
		DB:       db,
	}

	os.Setenv("SESSION_KEY", app.AuthCode)
	googleClientId, ok := os.LookupEnv("GOOGLE_CLIENT_ID")

	if !ok {
		log.Fatal("GOOGLE_CLIENT_ID is missing in the env")
	}

	googleClientSecret, ok := os.LookupEnv("GOOGLE_CLIENT_SECRET")

	if !ok {
		log.Fatal("GOOGLE_CLIENT_SECRET is missing in the env")
	}

	callbackLink, ok := os.LookupEnv("CALLBACK_LINK")

	if !ok {
		log.Fatal("CALLBACK_LINK is missing in the env")
	}

	_, is_dev := os.LookupEnv("DEV")
	log.Println("Is dev: ", is_dev)

	log.Println(callbackLink)

	store := sessions.NewCookieStore([]byte(app.AuthCode))
	store.MaxAge(86400 * 30)

	store.Options.Path = "/"
	store.Options.Secure = !is_dev
	store.Options.HttpOnly = is_dev

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, callbackLink),
	)

	log.Println("Listening on :8080")

	h2s := &http2.Server{}

	router := app.Router()
	stack := middleware.CreateStack(
		middleware.Logging,
		func(next http.Handler) http.Handler {
			return h2c.NewHandler(next, h2s)
		},
	)

	authRouter := app.AuthenticatedRouter()
	router.Handle("/api/", http.StripPrefix("/api", middleware.IsAuthenticated(authRouter)))

	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	if is_dev {
		err = server.ListenAndServeTLS("server.pem", "server.key")
	} else {
		err = server.ListenAndServe()
	}
	log.Println(err)
}
