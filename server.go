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
	store.Options.HttpOnly = true

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, callbackLink),
	)

	router := app.Router()
	stack := middleware.CreateStack(
		middleware.Logging,
	)

	authRouter := app.AuthenticatedRouter()
	router.Handle("/api/", http.StripPrefix("/api", middleware.IsAuthenticated(authRouter)))

	if err != nil {
		log.Fatal(err)
	}

	protocols := http.Protocols{}
	protocols.SetUnencryptedHTTP2(true)

	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	if !is_dev {
		server.Protocols = &protocols
	}

	log.Println("Listening on :8080")

	certFile := ""
	keyFile := ""

	if is_dev {
		certFile = "server.pem"
		keyFile = "server.key"
	}

	err = server.ListenAndServeTLS(certFile, keyFile)
	log.Println(err)
}
