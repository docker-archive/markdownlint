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

		f, err := os.Open(file)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}
		reader := bufio.NewReader(f)

		err = checkHugoFrontmatter(reader)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}
		f.Close()
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
func checkHugoFrontmatter(reader *bufio.Reader) (err error) {
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
	}
	//	if foundComment {
	//		return 99, fmt.Errorf("missing a close html comment around frontmatter")
	//	}
	return nil
}
