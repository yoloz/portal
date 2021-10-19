package docs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type SavePage struct{}

func (ac *SavePage) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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
	dir, _ := path.Split(pagePath)
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
	rw.Write([]byte("ok"))
}

type UploadImage struct{}

type imageCallBack struct {
	success int
	message string
	url     string
}

func (icb *imageCallBack) toString() string {
	format := `{"success":%d,"message":"%s","url":"%s"}`
	return fmt.Sprintf(format, icb.success, icb.message, icb.url)
}

func (ui *UploadImage) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	var icb imageCallBack
	var err error
	if err = r.ParseForm(); err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}
	uploadFile, handle, err := r.FormFile("editormd-image-file")
	if err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}

	if err = os.MkdirAll("docs/upload/", 0744); err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}
	saveFile, err := os.OpenFile("docs/upload/"+handle.Filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}
	if _, err = io.Copy(saveFile, uploadFile); err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}

	defer uploadFile.Close()
	defer saveFile.Close()

	icb.success = 1
	icb.message = "ok"
	icb.url = saveFile.Name()
	rw.Write([]byte(icb.toString()))
}

type MoveFile struct{}

func (ac *MoveFile) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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
	dir, _ := path.Split(newpath)
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

	rw.Write([]byte(newpath))
}
