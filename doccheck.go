package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
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

		reader, err := OpenReader(details.fullPath)
		if err != nil {
			fmt.Printf("ERROR opening: %s\n", err)
			errorCount++
		}

		err = checkHugoFrontmatter(reader)
		if err != nil {
			fmt.Printf(" %s\n", file)
			fmt.Printf("ERROR frontmatter: %s\n", err)
			errorCount++
		}
		reader.Close()
	}

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

// https://gohugo.io/content/front-matter/
func checkHugoFrontmatter(reader *LineReader) (err error) {
	foundComment := false
	for err == nil {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		if buff == "+++" {
			verboseLog("Found TOML start")
			break
		}
		if strings.HasPrefix(buff, "<!--") {
			if !strings.HasSuffix(buff, "-->") {
				verboseLog("found comment start")
				foundComment = true
				continue
			}
		}
		//verboseLog("ReadLine: %s, %v, %s\n", string(byteBuff), isPrefix, err)
		for i := 0; i < len(buff); {
			runeValue, width := utf8.DecodeRuneInString(buff[i:])
			if unicode.IsSpace(runeValue) {
				i += width
			} else {
				verboseLog("Unexpected non-whitespace char: %s", buff)
				return fmt.Errorf("Unexpected non-whitespace char: %s", buff)
			}
		}
	}

	// read lines until `+++` ending
	for err == nil {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		if buff == "+++" {
			verboseLog("Found TOML end")
			break
		}
		verboseLog("\t%s\n", buff)
	}
	// remove trailing close comment
	if foundComment {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		verboseLog("is this a comment? (%s)\n", buff)
		if strings.HasSuffix(buff, "-->") {
			if !strings.HasPrefix(buff, "<!--") {
				verboseLog("found comment end")
				foundComment = false
			}
		}
		if foundComment {
			reader.UnreadLine(buff)
			return fmt.Errorf("Did not find expected close metadata comment")
		}
	}
	return nil
}

// Fake Reader that can 'unread' a complete line
type LineReader struct {
	file       *os.File
	reader     *bufio.Reader
	unreadLine string
}

// For testing
func ByteReader(str string) *LineReader {
	reader := strings.NewReader(str)
	r := new(LineReader)
	r.reader = bufio.NewReader(reader)
	return r
}

func OpenReader(filename string) (*LineReader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	r := new(LineReader)
	r.file = f
	r.reader = reader
	return r, nil
}

func (r *LineReader) ReadLine() (line []byte, isPrefix bool, err error) {
	if r.unreadLine == "" {
		return r.reader.ReadLine()
	}
	lines := strings.SplitN(r.unreadLine, "\n", 2)
	r.unreadLine = lines[1]
	return []byte(lines[0]), false, nil
}

func (r *LineReader) UnreadLine(str string) {
	r.unreadLine = strings.Join([]string{str, r.unreadLine}, "\n")
}

func (r *LineReader) Close() {
	if r.file != nil {
		r.file.Close()
	}
}
