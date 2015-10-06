package checkers

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/SvenDowideit/markdownlint/linereader"
)

// https://gohugo.io/content/front-matter/
func checkHugoFrontmatter(reader *linereader.LineReader, file string) (err error) {
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

	allFiles[file].meta = make(map[string]string)

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

		meta := strings.SplitN(buff, "=", 2)
		verboseLog("\t%d\t%v\n", len(meta), meta)
		if len(meta) == 2 {
			verboseLog("\t\t%s: %s\n", meta[0], meta[1])
			allFiles[file].meta[strings.Trim(meta[0], " ")] = strings.Trim(meta[1], " ")
		}
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

	// ensure that the minimum metadata keys are set
	// ignore draft files
	if draft, ok := allFiles[file].meta["draft"]; !ok || draft != "true" {
		if _, ok := allFiles[file].meta["title"]; !ok {
			return fmt.Errorf("Did not find `title` metadata element")
		}
		if _, ok := allFiles[file].meta["description"]; !ok {
			return fmt.Errorf("Did not find `description` metadata element")
		}
		if _, ok := allFiles[file].meta["keywords"]; !ok {
			return fmt.Errorf("Did not find `keywords` metadata element")
		}
	}
	return nil
}
