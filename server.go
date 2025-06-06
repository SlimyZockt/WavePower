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
	_ "github.com/tursodatabase/go-libsql"
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
	_, isDev := os.LookupEnv("DEV")
	log.Println("Is dev: ", isDev)

	// AUTH
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackLink := os.Getenv("CALLBACK_LINK")

	if err != nil {
		log.Fatal(err)
	}

	dbUrl := "file:./app.db"

	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		log.Fatalf("failed to open db %s: %s", dbUrl, err)
	}
	defer db.Close()

	app := routes.App{
		IsDev:    isDev,
		AuthCode: GenStrin(16),
		DB:       db,
	}

	os.Setenv("SESSION_KEY", app.AuthCode)

	store := sessions.NewCookieStore([]byte(app.AuthCode))
	store.MaxAge(86400 * 30)

	store.Options.Path = "/"
	store.Options.Secure = !isDev
	store.Options.HttpOnly = true

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, callbackLink),
	)

	h2s := &http2.Server{}

	router := app.Router()
	stack := middleware.CreateStack(
		middleware.Logging,
		func(next http.Handler) http.Handler {
			return h2c.NewHandler(next, h2s)
		},
	)

	authRouter := app.AuthenticatedRouter()
	authHandler := http.StripPrefix(
		"/api",
		middleware.IsAuthenticated(authRouter),
	)
	router.Handle("/api/", authHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	log.Println("Listening on :8080")
	if isDev {
		err = server.ListenAndServeTLS("server.pem", "server.key")

	} else {
		err = server.ListenAndServe()
	}
	log.Println(err)
}
