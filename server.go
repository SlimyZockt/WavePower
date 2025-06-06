package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"server/internal/middleware"
	"server/internal/routes"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	// DB
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	dbUrl := os.Getenv("TURSO_DATABASE_URL")

	// AUTH
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackLink := os.Getenv("CALLBACK_LINK")

	// S3
	// s3Token := os.Getenv("S3_TOKEN")
	s3ID := os.Getenv("S3_ID")
	s3Secret := os.Getenv("S3_SECRET")
	s3Endpoint := os.Getenv("S3_ENDPOINT")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3ID, s3Secret, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
	})

	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	if isDev || dbUrl == "" {
		dbUrl = "file:./app.db"
	}

	dbUrl += "?authToken=" + authToken

	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		log.Fatalf("failed to open db %s: %s", dbUrl, err)
	}
	defer db.Close()

	app := routes.App{
		isDev,
		GenStrin(16),
		db,
		uploader,
		downloader,
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
		middleware.IsAuthenticated(authRouter, app),
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
