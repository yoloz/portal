package post

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestGenerateDocsifyIndex(t *testing.T) {
	home, _ := os.UserHomeDir()
	fs, err := ioutil.ReadDir(home + "/projects/docs")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		if f.IsDir() && strings.Compare(f.Name(), "template") != 0 {
			parentpath := home + "/projects/docs/" + f.Name()
			if err := os.Remove(parentpath + "/_sidebar.md"); err != nil {
				log.Fatal(err)
			}
			if err := os.Remove(parentpath + "/README.md"); err != nil {
				log.Fatal(err)
			}
			GenerateDocsifyIndex(parentpath)
		}

	}
}
