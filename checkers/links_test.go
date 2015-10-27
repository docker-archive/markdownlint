package checkers

import (
	"bytes"
	"testing"

	"github.com/SvenDowideit/markdownlint/data"
	"github.com/russross/blackfriday"
)

func TestMarkdownLinks(t *testing.T) {
	file := "test/index.md"
	data.AllFiles = make(map[string]*data.FileDetails)
	data.AllFiles[file] = new(data.FileDetails)
	data.AllFiles[file].FullPath = file

	htmlFlags := 0
	renderer := &TestRenderer{LinkFrom: file, Html: blackfriday.HtmlRenderer(htmlFlags, "", "").(*blackfriday.Html)}

	link := "../first.md"
	out := bytes.NewBuffer(make([]byte, 1024))
	renderer.Link(out, []byte(link), []byte("title"), []byte("content"))

	//if err != nil {
	//	t.Errorf("ERROR parsing: %v", err)
	//}
}
