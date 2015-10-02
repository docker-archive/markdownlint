package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// allFiles a lookup table of all the files in the 'docs' dir
// also takes advantage of the random order to avoid testing markdown files in the same order.
var allFiles map[string]bool

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(-1)
	}
	dir := os.Args[1]

	allFiles = make(map[string]bool)

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
		// fmt.Printf("\t walked to %s\n", file)
		allFiles[file] = true
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(-1)
	}

	for file, _ := range allFiles {
		fmt.Printf(" %s\n", file)

		reader, err := os.Open(file)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}

		sectionReader := io.NewSectionReader(reader, 0, 2048)
		offset, err := checkHugoFrontmatter(sectionReader)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}
		if offset <= 0 {
			fmt.Printf("ERROR: no frontmatter found\n")
			os.Exit(-1)
		}
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("\tFound %d files\n", len(allFiles))
	// return the number of 404's to show that there are things to be fixed
	os.Exit(0)
}

func printUsage() {
	fmt.Println("Please specify a directory to check")
	fmt.Println("\tfor example: docscheck .")
}

// https://gohugo.io/content/front-matter/
func checkHugoFrontmatter(reader *io.SectionReader) (offset int, err error) {
	byteBuff := make([]byte, 2048)
	length, err := reader.Read(byteBuff)
	if err != nil && err != io.EOF {
		return length, err
	}
	buff := string(byteBuff)

	// remove any leading empty lines
	i := 0
	for i < len(buff) {
		runeValue, width := utf8.DecodeRuneInString(buff[i:])
		if unicode.IsSpace(runeValue) {
			i += width
		} else {
			break
		}
	}
	// remove the next line if it starts with `<!--'
	if strings.HasPrefix(buff[i:], "<!--") {
		lineEnd := strings.IndexAny(buff[i:], "\n")
		if lineEnd != -1 {
			startComment := strings.TrimSuffix(buff[i:lineEnd], "\r")
			if !strings.HasSuffix(startComment, "-->") {
				fmt.Println("found comment start")
				i += lineEnd + 1
			}
		}
	}
	// frontmatter marker
	if !strings.HasPrefix(buff[i:], "+++") {
		return 0, fmt.Errorf("No TOML fronmatter marker (+++) found (next string: %s)", buff[i:i+10])
	}
	i += len("+++\n")
	if buff[i] == '\r' {
		i++ // \r\n
	}
	fmt.Printf("Next up: %s\n", buff[i:i+10])
	// read lines until `+++` ending
	// remove trailing close comment
	return i, nil
}
