package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"portal/docs"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "index.html")
	})
	r.HandleFunc("/newPage", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "docs/newPage.html")
	})
	r.Handle("/docs/savepage", new(docs.SavePage))
	r.Handle("/docs/upload", new(docs.UploadImage))
	r.Handle("/docs/movefile", new(docs.MoveFile))

	r.PathPrefix("/docs/editormd").Handler(http.StripPrefix("/docs/editormd", http.FileServer(http.Dir("docs/editormd"))))

	file, err := os.Open("config/config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	config := make(map[string]string)
	json.NewDecoder(file).Decode(&config)

	host, exist := config["host"]
	if !exist || strings.Compare(host, "0.0.0.0") == 0 {
		host = ""
	}

	port, exist := config["port"]
	if !exist {
		port = "10010"
	}

	for k, v := range config {
		if strings.Compare("host", k) == 0 || strings.Compare("port", k) == 0 {
			continue
		}
		r.PathPrefix(k).Handler(http.StripPrefix(k, http.FileServer(http.Dir(v))))
	}

	srv := http.Server{
		Handler:      r,
		Addr:         host + ":" + port,
		WriteTimeout: time.Second * 10,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("start http server......")
	if err := srv.ListenAndServeTLS("config/cert.pem", "config/key.pem"); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}
