package checkers

import (
	"bytes"
	"fmt"

	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"

	"github.com/russross/blackfriday"
)

func CheckMarkdownLinks(reader *linereader.LineReader, file string) (err error) {
	// blackfriday.HtmlRendererWithParameters(htmlFlags, "", "", renderParameters)
	htmlFlags := 0
	renderer := &TestRenderer{Html: blackfriday.HtmlRenderer(htmlFlags, "", "").(*blackfriday.Html)}

	extensions := 0
	//var output []byte
	buf := make([]byte, 32*1024)
	length, err := reader.Read(buf)
	if length == 0 || err != nil {
		return err
	}
	_ = blackfriday.Markdown(buf, renderer, extensions)

	return nil
}

func LinksSummary() {
	linkCount := 0
	for link, details := range data.AllLinks {
		data.VerboseLog("\t\t%d links to %s\n", details.Count, link)
		linkCount++
		// TODO: check the links
	}
	fmt.Printf("\tTotal Links: %d\n", linkCount)
}

type TestRenderer struct {
	*blackfriday.Html
}

func (renderer *TestRenderer) Link(out *bytes.Buffer, linkB []byte, title []byte, content []byte) {
	link := string(linkB)
	_, ok := data.AllLinks[link]
	if !ok {
		data.AllLinks[link] = new(data.LinkDetails)
	}
	data.AllLinks[link].Count++
}
