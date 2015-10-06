package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SvenDowideit/docscheck/checkers"
	"github.com/SvenDowideit/docscheck/linereader"
)

var verbose = flag.Bool("v", false, "verbose log output")

func verboseLog(format string, a ...interface{}) (n int, err error) {
	if !*verbose {
		return 0, nil
	}
	return fmt.Printf(format, a...)
}

// allFiles a lookup table of all the files in the 'docs' dir
// also takes advantage of the random order to avoid testing markdown files in the same order.
type fileDetails struct {
	fullPath string
	meta     map[string]string
}

var allFiles map[string]*fileDetails

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(-1)
	}
	dir := os.Args[1]

	allFiles = make(map[string]*fileDetails)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := filepath.Rel(dir, path)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		// verboseLog("\t walked to %s\n", file)
		allFiles[file] = new(fileDetails)
		allFiles[file].fullPath = path
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(-1)
	}

	errorCount := 0
	for file, details := range allFiles {
		if !strings.HasSuffix(file, ".md") {
			continue
		}
		verboseLog(" %s\n", file)

		reader, err := linereader.OpenReader(details.fullPath)
		if err != nil {
			fmt.Printf("ERROR opening: %s\n", err)
			errorCount++
		}

		err = checkers.checkHugoFrontmatter(reader, file)
		if err != nil {
			fmt.Printf(" %s\n", file)
			fmt.Printf("ERROR frontmatter: %s\n", err)
			errorCount++
		}

		err = checkers.checkMarkdownLinks(reader, file)
		if err != nil {
			fmt.Printf(" %s\n", file)
			fmt.Printf("ERROR links: %s\n", err)
			errorCount++
		}
		reader.Close()
	}

	// TODO (JIRA: DOCS-181): Title, unique across products if not, file should include an {identifier}

	fmt.Printf("Summary:\n")
	fmt.Printf("\tFound %d files\n", len(allFiles))
	fmt.Printf("\tFound %d errors\n", errorCount)
	// return the number of 404's to show that there are things to be fixed
	os.Exit(errorCount)
}

func printUsage() {
	fmt.Println("Please specify a directory to check")
	fmt.Println("\tfor example: docscheck .")
	flag.PrintDefaults()
}
