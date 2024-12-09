package routes

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"math/rand/v2"
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

type MuxWrapper struct {
	*http.ServeMux
}

var Sessions = map[string]user.User{}

func writeBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(http.StatusText(http.StatusBadRequest)))
}

func (mux *MuxWrapper) HandleFuncErr(path string, callback func(http.ResponseWriter, *http.Request) error) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		err := callback(w, r)

		if err != nil {
			log.Println(err)
			writeBadRequest(w)
			return
		}
	})
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

func (app *App) setUser(r *http.Request, user *user.User) error {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return errors.New("Cookie not found")
	}
	token := cookie.Value

	json_playlist, err := json.Marshal(user.Tracks)

	_, err = app.DB.Exec("UPDATE users SET playlist = ? WHERE ID = ?", string(json_playlist), user.UserID)
	if err != nil {
		return err
	}

	Sessions[token] = *user

	return nil
}

func (app *App) AuthenticatedRouter() *MuxWrapper {
	router := &MuxWrapper{http.NewServeMux()}

	router.HandleFuncErr("GET /audio/{id}/", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")

		if id == "" {
			return errors.New("No ID given")
		}

		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		cUser.CurAudioID = id
		err = app.setUser(r, cUser)
		if err != nil {
			return err
		}

		executable_path, _ := os.Executable()
		executable_dir := filepath.Dir(executable_path)

		audio_dir := filepath.Join(
			executable_dir,
			fmt.Sprintf("/uploads/%s/%s", cUser.UserID, id),
		)

		if _, err := os.Stat(audio_dir); !os.IsNotExist(err) {
			fs := http.FileServer(http.Dir(audio_dir))

			http.StripPrefix("/audio/"+id+"/", fs).ServeHTTP(w, r)
		}

		return nil
	})

	router.HandleFuncErr("POST /next_track/shuffle", func(w http.ResponseWriter, r *http.Request) error {
		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}
		rnd := rand.IntN(len(cUser.Tracks))

		fmt.Println(rnd)

		fmt.Fprintf(w, "%s", cUser.Tracks[rnd].Id)
		return nil
	})

	router.HandleFuncErr("POST /next_track/{id}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		idx := 0
		for i, val := range cUser.Tracks {
			if val.Id == id {
				if i == len(cUser.Tracks)-1 {
					idx = 0
					break
				}
				idx = i + 1
				break
			}
		}

		fmt.Println(idx)

		fmt.Fprint(w, cUser.Tracks[idx].Id)
		return nil

	})

	router.HandleFuncErr("POST /track_display", func(w http.ResponseWriter, r *http.Request) error {
		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		track, err := cUser.Tracks.GetTrack(cUser.CurAudioID)
		if err != nil {
			return err
		}

		components.TrackDisplay(track).Render(r.Context(), w)
		return nil
	})

	router.HandleFuncErr("POST /track_display/{id}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")

		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		track, err := cUser.Tracks.GetTrack(id)
		if err != nil {
			return err
		}

		components.TrackDisplay(track).Render(r.Context(), w)
		return nil
	})

	router.HandleFuncErr("POST /refresh_token", func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		user.ExpiresAt = user.ExpiresAt.Add(time.Hour)

		_ = app.setUser(r, user)

		return nil
	})

	router.HandleFuncErr("POST /loggedin", func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		components.LoggedIn(user).Render(r.Context(), w)
		return nil
	})

	router.HandleFuncErr("POST /playlist", func(w http.ResponseWriter, r *http.Request) error {

		user, err := app.getUser(r)
		if err != nil {
			return err
		}

		components.Playlist(user).Render(r.Context(), w)

		return nil
	})

	router.HandleFuncErr("POST /moved", func(w http.ResponseWriter, r *http.Request) error {
		cUser, err := app.getUser(r)
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
		for i, val := range cUser.Tracks {
			if val.Id == data.Grabbed {
				grabbed_id = i
			}
			if val.Id == data.Droped {
				droped_id = i
			}
		}

		cUser.Tracks.Move(grabbed_id, droped_id)

		app.setUser(r, cUser)

		return nil
	})

	router.Handle("POST /fileupload", templ.Handler(components.FileUpload()))

	router.HandleFuncErr("POST /delete/{id}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")

		executable_path, err := os.Executable()
		executable_dir := filepath.Dir(executable_path)

		cUser, err := app.getUser(r)
		if err != nil {
			return err
		}

		index := -1
		for i, val := range cUser.Tracks {
			if val.Id == id {
				index = i
				os.RemoveAll(filepath.Join(
					executable_dir,
					fmt.Sprintf("/uploads/%s/%s", cUser.UserID, cUser.Tracks[i].Id)),
				)
				break
			}
		}

		cUser.Tracks = slices.Delete(cUser.Tracks, index, index+1)

		app.setUser(r, cUser)
		return nil
	})

	router.HandleFuncErr("POST /upload/{name}", func(w http.ResponseWriter, r *http.Request) error {
		bytes := []byte{}
		buf := make([]byte, 1024)
		defer r.Body.Close()
		for {
			n, err := r.Body.Read(buf)
			if err == io.EOF {
				log.Println("EOF: ", n)
				break
			}

			if err != nil {
				return err
			}

			bytes = append(bytes, (buf[:n])...)
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

		hash := sha256.New()
		hash.Write(bytes)
		sha := base64.URLEncoding.EncodeToString(hash.Sum(nil))

		trackData := user.AudioTrack{Title: onlyFilename, Id: sha}

		if slices.Contains(cUser.Tracks, trackData) {
			return errors.New("Track allready exist")
		}

		out_dir := filepath.Join(
			executable_dir,
			fmt.Sprintf("/uploads/%s/%s", cUser.UserID, trackData.Id),
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

		metadata_file, err := os.Open(fmt.Sprintf("%s/metadata.txt", out_dir))
		if err != nil {
			return err
		}
		defer metadata_file.Close()

		scanner := bufio.NewScanner(metadata_file)

		metadata_map := map[string](*string){
			"TITLE":  &trackData.Title,
			"ARTIST": &trackData.Artist,
		}

		for scanner.Scan() {
			text := scanner.Text()
			for key := range maps.Keys(metadata_map) {
				if strings.HasPrefix(text, key) {
					after, found := strings.CutPrefix(text, key+"=")
					if !found {
						continue
					}
					*metadata_map[key] = after
				}
			}
		}

		cUser.Tracks = append(cUser.Tracks, trackData)

		fmt.Println(cUser.Tracks)
		app.setUser(r, cUser)

		return nil
	})

	return router

}

func convert_audio_2_hls(src_file string, out_dir string) error {
	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-i", src_file,
		"-profile:v", "baseline", // baseline profile is compatible with most devices
		"-level", "3.0",
		"-start_number", "0", // start numbering segments from 0
		"-hls_time", strconv.Itoa(10), // duration of each segment in seconds
		"-hls_list_size", "0", // keep all segments in the playlist
		"-f", "hls",
		fmt.Sprintf("%s/output.m3u8", out_dir),
		"-f",
		"ffmetadata",
		fmt.Sprintf("%s/metadata.txt", out_dir),
	)

	stderr, err := ffmpegCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	ffmpegCmd.Stdout = os.Stdout

	if err := ffmpegCmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stderr)
	colorRed := "\033[31m"
	colorReset := "\033[0m"
	for scanner.Scan() {
		text := scanner.Text()

		if strings.Contains(text, "Error") {
			fmt.Println(colorRed, text, colorReset)
			return errors.New("FFMPEG: " + text)
		} else {
			fmt.Println(text)
		}

	}

	if err := ffmpegCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (app *App) Router() *http.ServeMux {
	router := http.NewServeMux()

	static_fs := http.FileServer(http.Dir("./include_dir/"))

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
			User:          gUser,
			CurAudioID:    "",
			CurPlaylistID: 0,
			Tracks:        user.Tracks{},
		}

		new_user.User.Name = strings.Split(new_user.Email, "@")[0]

		playlist_str, err := json.Marshal(new_user.Tracks)
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

			json.Unmarshal([]byte(temp), &new_user.Tracks)
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
