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
	"os"
	"os/exec"
	"path/filepath"
	"server/components"
	"server/internal/user"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
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

var Sessions = map[string]user.User{}

func writeBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(http.StatusText(http.StatusBadRequest)))
}

func errWrapper(callback func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := callback(w, r)

		if err != nil {
			log.Println(err)
			writeBadRequest(w)
			return
		}
	}
}

func (app *App) getUser(r *http.Request) (*user.User, error) {
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

func (app *App) setUser(r *http.Request, user user.User) error {
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

	router.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})

	router.HandleFunc("GET /audio/", errWrapper(func(w http.ResponseWriter, r *http.Request) error {

		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		executable_path, _ := os.Executable()
		executable_dir := filepath.Dir(executable_path)

		audio_dir := filepath.Join(
			executable_dir,
			fmt.Sprintf("/uploads/%s", cUser.UserID),
		)

		if _, err := os.Stat(audio_dir); !os.IsNotExist(err) {
			fs := http.FileServer(http.Dir(audio_dir))

			http.StripPrefix("/audio/", fs).ServeHTTP(w, r)
		}

		return nil
	}))

	router.HandleFunc("POST /refresh_token", errWrapper(func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		user.ExpiresAt = user.ExpiresAt.Add(time.Hour)

		_ = app.setUser(r, *user)

		return nil
	}))

	router.HandleFunc("POST /loggedin", errWrapper(func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		components.LoggedIn(user).Render(r.Context(), w)
		return nil
	}))

	router.HandleFunc("POST /playlist", errWrapper(func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		components.Playlist(user).Render(r.Context(), w)

		return nil
	}))

	router.HandleFunc("POST /moved", errWrapper(func(w http.ResponseWriter, r *http.Request) error {
		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			return err
		}

		data := GrabData{}

		err = json.Unmarshal(bytes, &data)

		if err != nil {
			return err
		}

		if data.Grabbed == data.Droped || data.Grabbed == "" || data.Droped == "" {
			return errors.New("Grabbed Data is wrong")
		}

		droped_id := 0
		grabbed_id := 0
		for i, val := range user.Playlists[0] {
			if strconv.Itoa(val.Id) == data.Grabbed {
				grabbed_id = i
			}
			if strconv.Itoa(val.Id) == data.Droped {
				droped_id = i
			}
		}

		user.Playlists[0].Move(grabbed_id, droped_id)

		json_playlist, err := json.Marshal(user.Playlists)

		_, err = app.DB.Exec("UPDATE users SET playlist = ? WHERE ID = ?", string(json_playlist), user.UserID)
		if err != nil {
			return err
		}

		return nil
	}))

	router.Handle("POST /fileupload", templ.Handler(components.FileUpload()))

	router.HandleFunc("POST /upload/{name}", errWrapper(func(w http.ResponseWriter, r *http.Request) error {
		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			return err
		}

		filename := r.PathValue("name")

		cUser, err := app.getUser(r)

		if err != nil {
			return err
		}

		executable_path, err := os.Executable()
		executable_dir := filepath.Dir(executable_path)

		temp_path := filepath.Join(
			executable_dir,
			"uploads/temp/"+cUser.UserID+"@"+filename,
		)

		err = os.MkdirAll(filepath.Dir(temp_path), os.ModeDir)

		if err != nil && !os.IsExist(err) {
			return err
		}

		err = os.WriteFile(temp_path, bytes, os.ModeTemporary)
		if err != nil {
			return err
		}

		temp := strings.Split(filename, ".")
		onlyFilename := strings.Join(temp[:len(temp)-1], ".")

		audioData := user.AudioTrack{Name: onlyFilename, Id: len(cUser.Playlists[0])}

		if slices.Contains(cUser.Playlists[0], audioData) {
			return errors.New("Song allready exist")
		}

		out_dir := filepath.Join(
			executable_dir,
			fmt.Sprintf("/uploads/%s/%d", cUser.UserID, audioData.Id),
		)

		err = os.MkdirAll(out_dir, os.ModeDir)
		if err != nil && !os.IsExist(err) {
			return err
		}

		errChan := make(chan error)

		go func() {
			err := convert_audio_2_hls(temp_path, out_dir)
			errChan <- err

			err = os.Remove(temp_path)
			if err != nil {
				errChan <- err
				return
			}

		}()

		if <-errChan != nil {
			return err
		}

		song := user.AudioTrack{
			Name: onlyFilename,
		}

		cUser.Playlists[0] = append(cUser.Playlists[0], song)

		json_playlist, err := json.Marshal(cUser.Playlists)

		_, err = app.DB.Exec("UPDATE users SET playlist = ? WHERE ID = ?", string(json_playlist), cUser.UserID)
		if err != nil {
			return err
		}

		return nil
	}))

	return router

}

func convert_audio_2_hls(src_file string, out_dir string) error {
	fmpegCmd := exec.Command(
		"ffmpeg",
		"-i", src_file,
		"-profile:v", "baseline", // baseline profile is compatible with most devices
		"-level", "3.0",
		"-start_number", "0", // start numbering segments from 0
		"-hls_time", strconv.Itoa(10), // duration of each segment in seconds
		"-hls_list_size", "0", // keep all segments in the playlist
		"-f", "hls",
		fmt.Sprintf("%s/output.m3u8", out_dir),
	)

	fmpegCmd.Stderr = os.Stderr
	fmpegCmd.Stdout = os.Stdout

	err := fmpegCmd.Start()
	if err != nil {
		return err
	}

	err = fmpegCmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) Router() *http.ServeMux {
	router := http.NewServeMux()

	static_fs := http.FileServer(http.Dir("./include_dir/"))

	// router.HandleFunc("GET /audio/{hash}", errWrapper(func(w http.ResponseWriter, r *http.Request) error {
	//
	// 	name := r.PathValue("hash")
	//
	// 	req, err := http.NewRequest("POST", fmt.Sprintf("/audio/%s", name), nil)
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	io.Copy(w, req.Body)
	//
	// 	return nil
	// }))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		http.StripPrefix("", static_fs).ServeHTTP(w, r)
	})

	router.HandleFunc("GET /auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

		gUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			// fmt.Fprint(w, err)
			log.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		new_user := user.User{
			User:      gUser,
			Playlists: []user.Playlist{[]user.AudioTrack{}},
		}

		new_user.User.Name = strings.Split(new_user.Email, "@")[0]

		playlist_str, err := json.Marshal(new_user.Playlists)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		row, err := app.DB.Query("SELECT playlist FROM users WHERE id = ?", gUser.UserID)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer row.Close()

		if row.Next() {
			temp := ""
			if err := row.Scan(&temp); err != nil {
				log.Println(err)
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}

			log.Println(temp)

			json.Unmarshal([]byte(temp), &new_user.Playlists)
		} else {
			_, _ = app.DB.Exec("INSERT INTO users (id, email, playlist, name) VALUES (?, ?, ?, ?)", new_user.UserID, new_user.Email, string(playlist_str), new_user.Name)
		}

		Sessions[new_user.AccessToken] = new_user
		// log.Println("Expires :", user.ExpiresAt.Format("2006-01-02 15:04:05"))

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   new_user.AccessToken,
			Expires: new_user.ExpiresAt,
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
