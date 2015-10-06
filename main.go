package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SvenDowideit/markdownlint/checkers"
	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"
)

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
		data.VerboseLog(" %s\n", file)

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
	fmt.Println("\tfor example: markdownlint .")
	flag.PrintDefaults()
}
