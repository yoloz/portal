package post

import (
	"container/list"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GenerateDocsifyIndex 根据目录path生成readme和_sidebar
func GenerateDocsifyIndex(path string) error {

	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	var str strings.Builder
	flist := list.New()
	flist.PushBack("- 目录\n")
	for _, f := range fs {
		if f.IsDir() ||
			strings.HasPrefix(f.Name(), "_") ||
			strings.HasPrefix(f.Name(), ".") ||
			strings.Compare(f.Name(), "index.html") == 0 ||
			strings.Compare(f.Name(), "README.md") == 0 {
			continue
		}
		str.Reset()
		str.WriteString("  - [")
		str.WriteString(strings.TrimSuffix(f.Name(), ".md"))
		str.WriteString("](/")
		str.WriteString(f.Name())
		str.WriteString(")\n")
		flist.PushBack(str.String())
	}
	//生成文件
	siderbar, err := os.Create(filepath.Join(path, "_sidebar.md"))
	if err != nil {
		return err
	}
	defer siderbar.Close()

	readme, err := os.Create(filepath.Join(path, "README.md"))
	if err != nil {
		return err
	}
	defer readme.Close()

	readme.WriteString("## 首页\n")
	for e := flist.Front(); e != nil; e = e.Next() {
		line := e.Value.(string)
		siderbar.WriteString(line)
		// readme.WriteString(line)
	}
	return nil
}
