package data

import (
	"flag"
	"fmt"
)

var verbose = flag.Bool("v", false, "verbose log output")

func VerboseLog(format string, a ...interface{}) (n int, err error) {
	if !*verbose {
		return 0, nil
	}
	return fmt.Printf(format, a...)
}

// allFiles a lookup table of all the files in the 'docs' dir
// also takes advantage of the random order to avoid testing markdown files in the same order.
type FileDetails struct {
	FullPath         string
	Meta             map[string]string
	FormatErrors     string
	FormatErrorCount int
}

func NewFileDetails(file, path string) *FileDetails {
	detail := new(FileDetails)
	detail.FullPath = path
	detail.Meta = make(map[string]string)
	detail.FormatErrors = ""
	detail.FormatErrorCount = 0
	return detail
}

type BigMap map[string]*FileDetails

var AllFiles BigMap

func AddFile(file, path string) {
	if _, ok := AllFiles[file]; !ok {
		AllFiles[file] = NewFileDetails(file, path)
	}
}

type LinkDetails struct {
	Count      int
	LinksFrom  map[int]string
	ActualLink map[int]string
	Response   int
}

var ResponseCode = map[int]string{
	999:  "failed to parse",
	888:  "failed to crawl",
	2900: "local file path - ok",
	900:  "mail/irc link, not checked",
	200:  "ok",
	777:  "source type path, but no match found",
	290:  "local file path, but missing `.md`",
	404:  "external url, but failed",
	666:  "Don't link to docs.docker.com",
}

var AllLinks map[string]*LinkDetails = make(map[string]*LinkDetails)
