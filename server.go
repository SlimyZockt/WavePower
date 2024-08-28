package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"server/components"
	"time"

	"github.com/a-h/templ"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	router := http.NewServeMux()
	dir := http.Dir("./include_dir/")
	fs := http.FileServer(dir)

	router.Handle("/", fs)
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

		fmt.Printf("written %d bytes over network \n", n)
	})

	log.Println("Listening on :8080")

	server := http.Server{
		Addr:    ":8080",
		Handler: Logger(router),
	}

	server.ListenAndServe()

}
