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
	if err != nil {
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

	// os.Setenv("SESSION_SECRET", app.AuthCode)
	os.Setenv("SESSION_KEY", app.AuthCode)
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackLink := os.Getenv("CALLBACK_LINK")

	log.Println(callbackLink)

	store := sessions.NewCookieStore([]byte(app.AuthCode))
	store.MaxAge(86400 * 30)

	store.Options.Path = "/"
	store.Options.Secure = false
	store.Options.HttpOnly = true

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, callbackLink),
	)

	log.Println("Listening on :8080")

	router := app.Router()
	stack := middleware.CreateStack(
		middleware.Logging,
	)

	authRouter := app.AuthenticatedRouter()
	router.Handle("/api/", http.StripPrefix("/api", middleware.IsAuthenticated(authRouter)))

	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	err = server.ListenAndServeTLS("./server.pem", "./server.key")

	log.Println(err)
}
