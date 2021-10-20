package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"portal/post"
)

//全局变量=>init()=>main()
var cfg_path string
var cfg_map = make(map[string]string)
var router = mux.NewRouter()

func main() {

	cfgs := os.Args
	if len(cfgs) == 0 {
		cfg_path = "../config"
	} else {
		cfg_path = cfgs[1]
	}
	if !strings.HasSuffix(cfg_path, "/") {
		cfg_path = cfg_path + "/"
	}
	file, err := os.Open(cfg_path + "config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&cfg_map)

	host, exist := cfg_map["host"]
	if !exist || strings.Compare(host, "0.0.0.0") == 0 {
		host = ""
	}
	port, exist := cfg_map["port"]
	if !exist {
		port = "20001"
	}

	handleRoute()

	srv := http.Server{
		Handler:      router,
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
			log.Printf("HTTPs server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("start https server......")
	if err := srv.ListenAndServeTLS(cfg_path+"cert.pem", cfg_path+"key.pem"); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTPs server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}

func handleRoute() {
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, cfg_path+"../static/index.html")
	})
	router.HandleFunc("/newPage", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, cfg_path+"../static/newPage.html")
	})
	router.PathPrefix("/editormd").Handler(http.StripPrefix("/editormd",
		http.FileServer(http.Dir(cfg_path+"../static/editormd"))))
	router.Handle("/upload", new(post.UploadImage))

	for k, v := range cfg_map {
		if strings.Compare("host", k) == 0 || strings.Compare("port", k) == 0 {
			continue
		}
		router.PathPrefix(k).Handler(http.StripPrefix(k, http.FileServer(http.Dir(v))))
	}

	router.HandleFunc("/moveimage", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		var err error
		if err = r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		oldpath := r.FormValue("oldpath")
		newpath := r.FormValue("newpath")
		var relpath string
		dir, fname := path.Split(newpath)
		if !strings.HasPrefix(newpath, "/") {
			if strings.HasSuffix(dir, "/") {
				dir = dir[:len(dir)-1]
			}
			relpath += dir
			absp, exist := cfg_map["/"+dir]
			if !exist {
				rw.Write([]byte("relative name[" + dir + "] can not find in configuration"))
				return
			}
			dir = absp + "/_images"
			newpath = dir + "/" + fname
			relpath += "/_images/" + fname
		} else {
			relpath = newpath
		}

		if _, err = os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0744); err != nil {
					emsg = fmt.Sprintf("%v", err)
					rw.Write([]byte(emsg))
				}
			} else {
				emsg = fmt.Sprintf("%v", err)
				rw.Write([]byte(emsg))
			}
			return
		}
		if err = os.Rename(oldpath, newpath); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}

		rw.Write([]byte(relpath))
	})

	router.HandleFunc("/savepage", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		var err error
		if err = r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		pagePath := r.FormValue("pagepath")
		md_text := r.FormValue("mdtext")
		dir, fname := path.Split(pagePath)
		if !strings.HasPrefix(pagePath, "/") {
			if strings.HasSuffix(dir, "/") {
				dir = dir[:len(dir)-1]
			}
			absp, exist := cfg_map["/"+dir]
			if !exist {
				rw.Write([]byte("relative name[" + dir + "] can not find in configuration"))
				return
			}
			dir = absp
			pagePath = dir + "/" + fname
		}

		if _, err = os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0744); err != nil {
					emsg = fmt.Sprintf("%v", err)
					rw.Write([]byte(emsg))
				}
			} else {
				emsg = fmt.Sprintf("%v", err)
				rw.Write([]byte(emsg))
			}
			return
		}
		if err = os.WriteFile(pagePath, []byte(md_text), 0744); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}

		if err = post.GenerateDocsifyIndex(dir); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		rw.Write([]byte("ok"))
	})
}
