package checkers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

// TODO: consider making the Summary function return the summary string, so it can continue processing.
func LinksSummary() {
	var statusCount = make(map[int]int)
	linkCount := 0
	for link, details := range data.AllLinks {
		linkCount++
		status := testUrl(link)
		data.AllLinks[link].Response = status
		statusCount[status]++
		if status == 200 {
			data.VerboseLog("\t\t(%d) %d links to %s\n", status, details.Count, link)
		} else {
			fmt.Printf("\t\t(%d) %d links to %s\n", status, details.Count, link)
		}
	}
	fmt.Printf("\tTotal Links: %d\n", linkCount)
	for status, count := range statusCount {
		fmt.Printf("\t\t%d: %d times\n", status, count)
	}
}

func testUrl(link string) int {
	base, err := url.Parse(link)
	if err != nil {
		fmt.Println("ERROR: failed to Parse \"" + link + "\"")
		return 999
	}
	switch base.Scheme {
	case "":
		// Internal markdown link
		// TODO: if it starts with a `#`, need to look for an anchor
		// otherwuse, look in data.AllFiles
		path := strings.Split(strings.Trim(link, "/"), "#")
		relUrl := strings.Trim(path[0], "/")
		if _, ok := data.AllFiles[relUrl]; ok {
			return 2900
		}
		if _, ok := data.AllFiles[relUrl+".md"]; ok {
			return 290
		}
		ok := 777
		return ok
	case "mailto", "irc":
		err = fmt.Errorf("%s", base.Scheme)
		return 900
	}
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + link + "\"  " + err.Error())
		return resp.StatusCode
		//return 888
	}

	loc, err := resp.Location()
	if err == nil && link != loc.String() {
		fmt.Printf("\t crawled \"%s\"", link)
		fmt.Printf("\t\t to \"%s\"", loc)
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	return resp.StatusCode
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
