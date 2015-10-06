package checkers

import (
	"bytes"
	"fmt"

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

type TestRenderer struct {
	*blackfriday.Html
}

func (renderer *TestRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	fmt.Printf("link: %s\n", link)
}
