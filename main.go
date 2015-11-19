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
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(-1)
	}
	dir := args[0]
	filter := ""
	if len(args) >= 2 {
		filter = args[1]
	}

	data.AllFiles = make(map[string]*data.FileDetails)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		data.VerboseLog("FOUND: %s\n", path)
		if info.IsDir() {
			return nil
		}
		file, err := filepath.Rel(dir, path)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		// verboseLog("\t walked to %s\n", file)
		data.AddFile(file, path)
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(-1)
	}

	frontmatterErrors := ""
	errorCount := 0
	for file, details := range data.AllFiles {
		if !strings.HasSuffix(file, ".md") {
			data.VerboseLog("SKIPPING: %s\n", file)
			continue
		}
		data.VerboseLog("PROCESSING: %s\n", file)

		reader, err := linereader.OpenReader(details.FullPath)
		if err != nil {
			fmt.Printf("ERROR opening: %s\n", err)
			errorCount++
		}

		err = checkers.CheckHugoFrontmatter(reader, file)
		if err != nil {
			fmt.Printf("ERROR (%s) frontmatter: %s\n", file, err)
			if strings.HasPrefix(file, filter) {
				frontmatterErrors = fmt.Sprintf("%sfrontmatter: (%s) %s\n", frontmatterErrors, file, err)
				errorCount++
			}
		}

		err = checkers.CheckMarkdownLinks(reader, file)
		if err != nil {
			// this only errors if there is a fatal issue
			fmt.Printf("ERROR (%s) links: %s\n", file, err)
			errorCount++
		}
		reader.Close()
	}
	checkers.LinksSummary()

	// TODO (JIRA: DOCS-181): Title, unique across products if not, file should include an {identifier}

	fmt.Printf("\n======================\n")
	if filter != "" {
		fmt.Printf("Filtered (%s) Summary:\n\n", filter)
	} else {
		fmt.Printf("Summary:\n\n")
	}
	fmt.Printf(frontmatterErrors)
	count, linkErr := checkers.LinkErrors(filter)
	errorCount = errorCount + count
	fmt.Printf(linkErr)
	fmt.Printf("\n\tFound %d files\n", len(data.AllFiles))
	fmt.Printf("\tFound %d errors\n", errorCount)
	// return the number of 404's to show that there are things to be fixed
	os.Exit(errorCount)
}

func printUsage() {
	fmt.Println("Please specify a directory to check")
	fmt.Println("\tfor example: markdownlint . [filter]")
	fmt.Println("\t [filter] can be any string prefix inside the dir specified")
	flag.PrintDefaults()
}
