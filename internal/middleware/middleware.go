package middleware

import (
	"log"
	"net/http"
	"slices"
	"time"

	"server/internal/routes"
)

type Middleware func(next http.Handler) http.Handler

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

type AppWrapper struct {
	*routes.App
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func CreateStack(stack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for _, item := range slices.Backward(stack) {
			next = item(next)
		}

		return next
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)
		log.Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}

func IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// _, ok := routes.Sessions["dev"]; ok &&
		cookie, err := r.Cookie("session_token")

		if err == http.ErrNoCookie {
			writeUnauthed(w)
			log.Println(err)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			log.Println(err)
			return
		}

		token := cookie.Value

		userSession, ok := routes.Sessions[token]
		if !ok {
			log.Println("No Session")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		if !userSession.ExpiresAt.After(time.Now()) {
			log.Println("NO Time")
			delete(routes.Sessions, token)
			writeUnauthed(w)
			return
		}

		log.Println("OK")
		next.ServeHTTP(w, r)
	})
}
