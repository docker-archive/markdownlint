package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/markdownlint/checkers"
	"github.com/docker/markdownlint/data"
	"github.com/docker/markdownlint/linereader"
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

	fmt.Println("Finding files")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			data.ErrorLog("%s\n", err)
			return err
		}
		data.VerboseLog("FOUND: %s\n", path)
		if info.IsDir() {
			return nil
		}
		file, err := filepath.Rel(dir, path)
		if err != nil {
			data.ErrorLog("%s\n", err)
			return err
		}
		// verboseLog("\t walked to %s\n", file)
		data.AddFile(file, path)
		return nil
	})
	if err != nil {
		data.ErrorLog("%s\n", err)
		os.Exit(-1)
	}

	count := 0
	for file, details := range data.AllFiles {
		if !strings.HasPrefix(file, filter) {
			data.VerboseLog("FILTERED: %s\n", file)
			continue
		}
		if !strings.HasSuffix(file, ".md") {
			data.VerboseLog("SKIPPING: %s\n", file)
			continue
		}
		// fmt.Printf("opening: %s\n", file)
		count++
		if count%100 == 0 {
			fmt.Printf("\topened %d files so far\n", count)
		}

		reader, err := linereader.OpenReader(details.FullPath)
		if err != nil {
			data.ErrorLog("%s\n", err)
			data.AllFiles[file].FormatErrorCount++
		}

		err = checkers.CheckHugoFrontmatter(reader, file)
		if err != nil {
			data.ErrorLog("(%s) frontmatter: %s\n", file, err)
		}

		if draft, ok := data.AllFiles[file].Meta["draft"]; ok || draft == "true" {
			data.VerboseLog("Draft=%s: SKIPPING %s link check.\n", draft, file)
		} else {
			//fmt.Printf("Draft=%s: %s link check.\n", draft, file)
			err = checkers.CheckMarkdownLinks(reader, file)
			if err != nil {
				// this only errors if there is a fatal issue
				data.ErrorLog("(%s) links: %s\n", file, err)
				data.AllFiles[file].FormatErrorCount++
			}
		}
		reader.Close()
	}
	fmt.Printf("Starting to test links (Filter = %s)\n", filter)
	checkers.TestLinks(filter, true)

	// TODO (JIRA: DOCS-181): Title, unique across products if not, file should include an {identifier}

	summaryFileName := "markdownlint.summary.txt"
	f, err := os.Create(summaryFileName)
	if err == nil {
		fmt.Printf("Also writing summary to %s :\n\n", summaryFileName)
		defer f.Close()
	}

	if filter != "" {
		Printf(f, "# Filtered (%s) Summary:\n\n", filter)
	} else {
		Printf(f, "# Summary:\n\n")
	}
	errorCount, errorString := checkers.FrontSummary(filter)
	Printf(f, errorString)
	count, errorString = checkers.LinkSummary(filter)
	errorCount += count
	//Printf(f, errorString)
	Printf(f, "\n\tFound: %d files\n", len(data.AllFiles))
	Printf(f, "\tFound: %d errors\n", errorCount)
	// return the number of 404's to show that there are things to be fixed
	os.Exit(errorCount)
}

func Printf(f *os.File, format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	fmt.Print(str)
	if f != nil {
		// Don't reall want to know we can't write..
		f.WriteString(str)
	}
}

func printUsage() {
	fmt.Println("Please specify a directory to check")
	fmt.Println("\tfor example: markdownlint . [filter]")
	fmt.Println("\t [filter] can be any string prefix inside the dir specified")
	flag.PrintDefaults()
}
