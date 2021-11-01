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

// absPath 返回绝对路径地址
func absPath(inpath string) (string, string, error) {
	dir, fname := path.Split(inpath)
	if !strings.HasPrefix(inpath, "/") {
		if strings.HasSuffix(dir, "/") {
			dir = dir[:len(dir)-1]
		}
		absp, exist := cfg_map["/"+dir]
		if !exist {
			return "", "", fmt.Errorf("relative name[%s] can not find in configuration", dir)
		}
		return absp, absp + "/" + fname, nil
	}
	return dir, inpath, nil
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

	router.HandleFunc("/rename", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		if err := r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		oldpath := r.FormValue("oldpath")
		newpath := r.FormValue("newpath")

		oldir, oldabs, err := absPath(oldpath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		newdir, newabs, err := absPath(newpath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}

		if _, err := os.Stat(newdir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(newdir, 0744); err != nil {
					emsg = fmt.Sprintf("%v", err)
					rw.Write([]byte(emsg))
				}
			} else {
				emsg = fmt.Sprintf("%v", err)
				rw.Write([]byte(emsg))
			}
			return
		}
		if err := os.Rename(oldabs, newabs); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		if strings.Compare(oldir, newdir) != 0 {
			if err = post.GenerateDocsifyIndex(oldir); err != nil {
				emsg = fmt.Sprintf("%v", err)
				rw.Write([]byte(emsg))
				return
			}
		}
		if err = post.GenerateDocsifyIndex(newdir); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		rw.Write([]byte("ok"))
	})

	router.HandleFunc("/moveimage", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		if err := r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		oldpath := r.FormValue("oldpath")
		newpath := r.FormValue("newpath")
		var newrelativepath string

		dir, fname := path.Split(newpath)
		if !strings.HasPrefix(newpath, "/") {
			if strings.HasSuffix(dir, "/") {
				dir = dir[:len(dir)-1]
			}
			newrelativepath += dir
			newrelativepath += "/_images/" + fname
			absp, exist := cfg_map["/"+dir]
			if !exist {
				emsg = fmt.Sprintf("relative name[%s] can not find in configuration", dir)
				rw.Write([]byte(emsg))
				return
			}
			dir = absp
			newpath = absp + "/_images/" + fname
		} else {
			newrelativepath = newpath
		}

		if _, err := os.Stat(dir); err != nil {
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
		if err := os.Rename(oldpath, newpath); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}

		rw.Write([]byte(newrelativepath))
	})

	router.HandleFunc("/delete", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		if err := r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		pagePath := r.FormValue("pagepath")
		dir, abspath, err := absPath(pagePath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		if err := os.RemoveAll(abspath); err != nil {
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

	router.HandleFunc("/savepage", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		if err := r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		pagePath := r.FormValue("pagepath")
		md_text := r.FormValue("mdtext")
		dir, abspath, err := absPath(pagePath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
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
		if err = os.WriteFile(abspath, []byte(md_text), 0744); err != nil {
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

	router.HandleFunc("/editpage", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		var emsg string
		if err := r.ParseForm(); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		pagePath := r.FormValue("pagepath")
		_, abspath, err := absPath(pagePath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		file, err := os.Open(abspath)
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		defer file.Close()

		fileinfo, err := file.Stat()
		if err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}

		filesize := fileinfo.Size()
		buffer := make([]byte, filesize)

		if _, err := file.Read(buffer); err != nil {
			emsg = fmt.Sprintf("%v", err)
			rw.Write([]byte(emsg))
			return
		}
		rw.Write(buffer)
	})
}
