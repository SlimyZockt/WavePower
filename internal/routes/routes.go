package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"server/components"
	_ "slices"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type App struct {
	AuthCode string
	DB       *sql.DB
}

type GrabData struct {
	Grabbed string
	Droped  string
}

var Sessions = map[string]goth.User{}

func writeBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(http.StatusText(http.StatusBadRequest)))
}

func (app *App) getUser(r *http.Request) (*goth.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, errors.New("Cookie not found")
	}

	token := cookie.Value

	user, ok := Sessions[token]
	if !ok {
		return nil, errors.New("Unvalid Session token")
	}

	if !user.ExpiresAt.After(time.Now()) {
		return nil, errors.New("Session expired")
	}

	return &user, nil
}

func (app *App) setUser(r *http.Request, user goth.User) error {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return errors.New("Cookie not found")
	}
	token := cookie.Value

	Sessions[token] = user

	return nil
}

func (app *App) AuthenticatedRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("POST /refresh_token", func(w http.ResponseWriter, r *http.Request) {

		user, err := app.getUser(r)
		if err != nil {
			writeBadRequest(w)
			return
		}

		user.ExpiresAt = user.ExpiresAt.Add(time.Hour)

		_ = app.setUser(r, *user)

	})

	router.HandleFunc("POST /loggedin", func(w http.ResponseWriter, r *http.Request) {

		user, err := app.getUser(r)
		if err != nil {
			writeBadRequest(w)
			return
		}

		components.LoggedIn(*user).Render(r.Context(), w)
	})

	router.Handle("POST /playlist", templ.Handler(components.PlaylistCmp()))

	router.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			writeBadRequest(w)
		}
		data := GrabData{}

		err = json.Unmarshal(bytes, &data)

		if err != nil {
			writeBadRequest(w)
			return
		}

		if data.Grabbed == data.Droped || data.Grabbed == "" || data.Droped == "" {
			writeBadRequest(w)
			return
		}

		droped_id := 0
		grabbed_id := 0
		for i, val := range components.Songs {
			if val == data.Grabbed {
				grabbed_id = i
			}
			if val == data.Droped {
				droped_id = i
			}
		}

		components.Songs.Move(grabbed_id, droped_id)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))

	})

	router.Handle("/fileupload", templ.Handler(components.FileUpload()))

	return router

}

func (app *App) Router() *http.ServeMux {
	router := http.NewServeMux()

	static_fs := http.FileServer(http.Dir("./include_dir/"))
	assets_fs := http.FileServer(http.Dir("./assets/"))

	router.Handle("GET /", static_fs)
	router.Handle("GET /audio_hls", assets_fs)

	router.HandleFunc("POST /audio/{id}", func(w http.ResponseWriter, r *http.Request) {
	})

	router.HandleFunc("GET /auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			// fmt.Fprint(w, err)
			log.Println(err)
			http.Redirect(w, r, "/", http.StatusNotFound)
			return
		}

		user.Name = strings.Split(user.Email, "@")[0]

		_, _ = app.DB.Exec("INSERT OR IGNORE INTO user (id, email, picture_url, name) VALUES (?, ?, ?, ?)", user.UserID, user.Email, user.Name)

		Sessions[user.AccessToken] = user
		// log.Println("Expires :", user.ExpiresAt.Format("2006-01-02 15:04:05"))

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   user.AccessToken,
			Expires: user.ExpiresAt,
			Path:    "/",
		})

		_, err = r.Cookie("session_token")
		http.Redirect(w, r, "/", http.StatusFound)
	})

	router.HandleFunc("GET /logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

		err := gothic.Logout(w, r)

		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now(),
			Path:    "/",
		})

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	router.HandleFunc("GET /auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

		if user, err := gothic.CompleteUserAuth(w, r); err == nil {
			log.Println("Logged in: ", user)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	})

	return router
}
