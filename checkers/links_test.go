package checkers

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/docker/markdownlint/data"
	"github.com/miekg/mmark"
)

func TestMarkdownLinks(t *testing.T) {
	file := "test/index.md"
	data.AllFiles = make(map[string]*data.FileDetails)
	data.AddFile(file, file)

	htmlFlags := 0
	renderParameters := mmark.HtmlRendererParameters{}
	renderer := &TestRenderer{
		LinkFrom: file,
		//Html:     mmark.HtmlRenderer(htmlFlags, "", "").(*mmark.Html),
		Renderer: mmark.HtmlRendererWithParameters(htmlFlags, "", "", renderParameters),
	}
	out := bytes.NewBuffer(make([]byte, 1024))

	tests := map[string]string{
		"../first.md":            "first.md",
		"second.md":              "test/second.md",
		"./second.md":            "test/second.md",
		"banana/second.md":       "test/banana/second.md",
		"/test/banana/second.md": "test/banana/second.md",
		"twice.md":               "test/twice.md",
		"banana/twice.md":        "test/banana/twice.md",
	}

	for _, path := range tests {
		data.AddFile(path, path)
	}
	for link, _ := range tests {
		renderer.Link(out, []byte(link), []byte("title"), []byte("content"))
	}

	for link, details := range data.AllLinks {
		data.AllLinks[link].Response = testUrl(link, "")
		//fmt.Printf("\t\t(%d) %d links to %s\n", data.AllLinks[link].Response, details.Count, link)
		fmt.Printf("%s links to %s\n", details.ActualLink[0], link)
		if _, ok := data.AllFiles[link]; !ok {
			t.Errorf("ERROR(%d): not found %s links to %s\n", details.Response, details.ActualLink[0], link)
		}
		if tests[details.ActualLink[0]] != link {
			t.Errorf("ERROR(%d): %s links to %s, should link to %s\n", details.Response, details.ActualLink[0], link, tests[details.ActualLink[0]])
		}
	}
}
