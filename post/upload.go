package post

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

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

	if err = os.MkdirAll("static/upload/", 0744); err != nil {
		icb.success = 0
		icb.message = fmt.Sprintf("%v", err)
		icb.url = ""
		rw.Write([]byte(icb.toString()))
		return
	}
	saveFile, err := os.OpenFile("static/upload/"+handle.Filename, os.O_WRONLY|os.O_CREATE, 0644)
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
