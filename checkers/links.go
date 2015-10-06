package checkers

import (
	"bytes"
	"fmt"

	"github.com/SvenDowideit/makedownlint/linereader"

	"github.com/russross/blackfriday"
)

func checkMarkdownLinks(reader *linereader.LineReader, file string) (err error) {
	// blackfriday.HtmlRendererWithParameters(htmlFlags, "", "", renderParameters)
	htmlFlags := 0
	title := ""
	var css string
	renderer := &TestRenderer{Html: blackfriday.HtmlRenderer(htmlFlags, "", "").(*blackfriday.Html)}

	extensions := 0
	var output []byte
	output = blackfriday.Markdown(reader.ReadAll(), renderer, extensions)

	return nil
}

type TestRenderer struct {
	*blackfriday.Html
}

func (renderer *TestRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	fmt.Printf("link: %s\n", link)
}
