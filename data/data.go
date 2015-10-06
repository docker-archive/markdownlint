package data

import (
	"flag"
	"fmt"
)

var verbose = flag.Bool("v", true, "verbose log output")

func VerboseLog(format string, a ...interface{}) (n int, err error) {
	if !*verbose {
		return 0, nil
	}
	return fmt.Printf(format, a...)
}

// allFiles a lookup table of all the files in the 'docs' dir
// also takes advantage of the random order to avoid testing markdown files in the same order.
type FileDetails struct {
	FullPath string
	Meta     map[string]string
}

var AllFiles map[string]*FileDetails

type LinkDetails struct {
	Count    int
	Response int
}

var AllLinks map[string]*LinkDetails = make(map[string]*LinkDetails)
