package checkers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"

	"github.com/miekg/mmark"
)

var skipUrls = map[string]int{
	"https://build.opensuse.org/project/show/Virtualization:containers": 1,
	"https://build.opensuse.org/":                                       1,
	"https://linux.oracle.com":                                          1,
	"http://supervisord.org/":                                           1,
	"http://goo.gl/HSz8UT":                                              1,
	"https://www.linkedin.com/company/docker":                           1,
	"https://cloud.docker.com/stack/deploy/":                            1,
	"https://cloud.docker.com/account/":                                 1,
	"https://reddit.com/r/docker":                                       1,
	"https://www.reddit.com/r/docker":                                   1,
}


func CheckMarkdownLinks(reader *linereader.LineReader, file string) (err error) {
	// mmark.HtmlRendererWithParameters(htmlFlags, "", "", renderParameters)
	htmlFlags := 0
	htmlFlags |= mmark.HTML_FOOTNOTE_RETURN_LINKS

	renderParameters := mmark.HtmlRendererParameters{
	//		FootnoteAnchorPrefix:       viper.GetString("FootnoteAnchorPrefix"),
	//		FootnoteReturnLinkContents: viper.GetString("FootnoteReturnLinkContents"),
	}

	renderer := &TestRenderer{
		LinkFrom: file,
		Renderer: mmark.HtmlRendererWithParameters(htmlFlags, "", "", renderParameters),
	}

	extensions := 0 |
		//mmark.EXTENSION_NO_INTRA_EMPHASIS |
		mmark.EXTENSION_TABLES | mmark.EXTENSION_FENCED_CODE |
		mmark.EXTENSION_AUTOLINK |
		//mmark.EXTENSION_STRIKETHROUGH |
		mmark.EXTENSION_SPACE_HEADERS | mmark.EXTENSION_FOOTNOTES |
		mmark.EXTENSION_HEADER_IDS | mmark.EXTENSION_AUTO_HEADER_IDS //|
	//	mmark.EXTENSION_DEFINITION_LISTS

	//var output []byte
	buf := make([]byte, 1024*1024)
	length, err := reader.Read(buf)
	if length == 0 || err != nil {
		return err
	}
	data.VerboseLog("RUNNING Markdown on %s length(%d) - not counting frontmater\n", file, length)
	_ = mmark.Parse(buf, renderer, extensions)
	data.VerboseLog("FINISHED Markdown on %s\n", file)

	return nil
}

var statusCount = make(map[int]int)

func LinkSummary(filter string) (int, string) {
	okCount := 0
	errorCount := 0
	errorString := ""
	for link, details := range data.AllLinks {
		if details.Response == 200 ||
			details.Response == 900 ||
			details.Response == 299 ||
			details.Response == 2900 ||
			details.Response == 666 {
			okCount++
		} else {
			for i, file := range data.AllLinks[link].LinksFrom {
				if strings.HasPrefix(file, filter) {
					errorCount++
					errorString = fmt.Sprintf("%s* link error: (in page %s) %s\n", errorString, file, data.AllLinks[link].ActualLink[i])
				}
			}
		}
	}
	return errorCount, errorString
}

func TestLinks(filter string) {
	linkCount := 0
	for link, details := range data.AllLinks {
		linkCount++
		status := testUrl(link, filter)
		data.AllLinks[link].Response = status
		statusCount[status]++
		if status == 200 ||
			details.Response == 900 ||
			details.Response == 299 ||
			status == 2900 ||
			status == 666 {
			data.VerboseLog("\t\t(%d) %d links to %s\n", status, details.Count, link)
		} else {
			fmt.Printf("\t\t (%d) %d links to (%s)\n", status, details.Count, link)
			for i, file := range data.AllLinks[link].LinksFrom {
				fmt.Printf("\t\t\t link %s on page %s\n", data.AllLinks[link].ActualLink[i], file)
			}
		}
	}
	for status, count := range statusCount {
		fmt.Printf("\t%d: %d times (%s)\n", status, count, data.ResponseCode[status])
	}
	fmt.Printf("\tTotal Links: %d\n", linkCount)
}

func testUrl(link, filter string) int {
	if _, ok := skipUrls[link]; ok {
		fmt.Printf("Skipping: %s\n", link)
		return 299
	}
	base, err := url.Parse(link)
	if err != nil {
		fmt.Println("ERROR: failed to Parse \"" + link + "\"")
		return 999
	}
	switch base.Scheme {
	case "":
		// Internal markdown link
		// otherwuse, look in data.AllFiles
		if strings.HasPrefix(link, "#") {
			// internal link to an anchor
			//TODO: need to look for anchor
			return 200
		} else {
			path := strings.Split(link, "#")
			relUrl := path[0]
			if !strings.HasPrefix(relUrl, filter) {
				//fmt.Printf("Filtered(%s): %s\n", filter, link)
				return 299
			}
			// TODO: need to test for path[1] anchor
			if _, ok := data.AllFiles[relUrl]; ok {
				return 2900
			}
			if _, ok := data.AllFiles[relUrl+".md"]; ok {
				return 290
			}
			fmt.Printf("\t\tERROR: failed to find %s or %s.md\n", relUrl, relUrl)
		}
		ok := 777
		return ok
	case "mailto", "irc":
		err = fmt.Errorf("%s", base.Scheme)
		return 900
	}
	// http / https
	if base.Host == "docs.docker.com" {
		err = fmt.Errorf("avoid linking directly to %s", base.Host)
		return 666
	}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := httpClient.Get(link)
	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + link + "\"  " + err.Error())
		return 888
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
	LinkFrom string
	mmark.Renderer
}

func (renderer *TestRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte, subFigure bool) {
	renderer.Link(out, link, title, alt)
}

func (renderer *TestRenderer) Link(out *bytes.Buffer, linkB []byte, title []byte, content []byte) {
	actualLink := string(linkB)
	data.VerboseLog("Link [%s](%s) in file %s\n", string(content), actualLink, renderer.LinkFrom)

	var link string

	base, err := url.Parse(actualLink)
	if err == nil && base.Scheme == "" {
		if strings.HasPrefix(actualLink, "#") {
			link = actualLink
		} else if strings.HasPrefix(actualLink, "/") {
			link = strings.TrimLeft(actualLink, "/")
		} else {
			// TODO: fix for relative paths.
			// TODO: need to check the from links are all the same dir too
			link = filepath.Clean(filepath.FromSlash(actualLink))

			if strings.IndexRune(link, os.PathSeparator) == 0 { // filepath.IsAbs fails to me.
				link = link[1:]
			} else {
				// TODO: need to check all the LinksFrom
				link = filepath.Join(filepath.Dir(renderer.LinkFrom), link)
			}
			data.VerboseLog("---- converted %s (on page %s, in %s) into %s\n", actualLink, renderer.LinkFrom, filepath.Dir(renderer.LinkFrom), link)
		}
	} else {
		link = actualLink
	}

	_, ok := data.AllLinks[link]
	if !ok {
		data.AllLinks[link] = new(data.LinkDetails)
		data.AllLinks[link].LinksFrom = make(map[int]string)
		data.AllLinks[link].ActualLink = make(map[int]string)
	}
	data.AllLinks[link].LinksFrom[data.AllLinks[link].Count] = renderer.LinkFrom
	data.AllLinks[link].ActualLink[data.AllLinks[link].Count] = actualLink
	data.AllLinks[link].Count++
}
