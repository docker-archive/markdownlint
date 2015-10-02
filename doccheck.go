package main

import (
	"bufio"
	"fmt"
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

		reader, err := OpenReader(file)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}

		err = checkHugoFrontmatter(reader)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}
		reader.Close()
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
func checkHugoFrontmatter(reader *LineReader) (err error) {
	foundComment := false
	for err == nil {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		if buff == "+++" {
			fmt.Println("Found TOML start")
			break
		}
		if strings.HasPrefix(buff, "<!--") {
			if !strings.HasSuffix(buff, "-->") {
				fmt.Println("found comment start")
				foundComment = true
				continue
			}
		}
		//fmt.Printf("ReadLine: %s, %v, %s\n", string(byteBuff), isPrefix, err)
		for i := 0; i < len(buff); {
			runeValue, width := utf8.DecodeRuneInString(buff[i:])
			if unicode.IsSpace(runeValue) {
				i += width
			} else {
				fmt.Printf("Unexpected non-whitespace char: %s", buff)
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
			fmt.Println("Found TOML end")
			break
		}
		fmt.Printf("\t%s\n", buff)
	}
	// remove trailing close comment
	if foundComment {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		fmt.Printf("is this a comment? (%s)\n", buff)
		if strings.HasSuffix(buff, "-->") {
			if !strings.HasPrefix(buff, "<!--") {
				fmt.Println("found comment end")
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
