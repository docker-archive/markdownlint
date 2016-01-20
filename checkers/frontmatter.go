package checkers

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"
)

// https://gohugo.io/content/front-matter/
func CheckHugoFrontmatter(reader *linereader.LineReader, file string) (err error) {
	err = doCheckHugoFrontmatter(reader, file)
	if err != nil {
		data.AllFiles[file].FormatErrors = fmt.Sprintf("%s* frontmatter: (%s) %s\n", data.AllFiles[file].FormatErrors, file, err)
		data.AllFiles[file].FormatErrorCount++

	}
	return err
}

func doCheckHugoFrontmatter(reader *linereader.LineReader, file string) (err error) {
	foundComment := false
	for err == nil {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		if buff == "+++" {
			data.VerboseLog("Found TOML start")
			break
		}
		if strings.HasPrefix(buff, "<!--") {
			if !strings.HasSuffix(buff, "-->") {
				data.VerboseLog("found comment start")
				foundComment = true
				continue
			}
		}
		//data.VerboseLog("ReadLine: %s, %v, %s\n", string(byteBuff), isPrefix, err)
		for i := 0; i < len(buff); {
			runeValue, width := utf8.DecodeRuneInString(buff[i:])
			if unicode.IsSpace(runeValue) {
				i += width
			} else {
				data.VerboseLog("Unexpected non-whitespace char: %s", buff)
				return fmt.Errorf("Unexpected non-whitespace char: %s", buff)
			}
		}
	}

	data.AllFiles[file].Meta = make(map[string]string)

	// read lines until `+++` ending
	for err == nil {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		if buff == "+++" {
			data.VerboseLog("Found TOML end")
			break
		}
		data.VerboseLog("\t%s\n", buff)

		meta := strings.SplitN(buff, "=", 2)
		data.VerboseLog("\t%d\t%v\n", len(meta), meta)
		if len(meta) == 2 {
			data.VerboseLog("\t\t%s: %s\n", meta[0], meta[1])
			data.AllFiles[file].Meta[strings.Trim(meta[0], " ")] = strings.Trim(meta[1], " ")
		}
	}
	// remove trailing close comment
	if foundComment {
		byteBuff, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		buff := string(byteBuff)
		data.VerboseLog("is this a comment? (%s)\n", buff)
		if strings.HasSuffix(buff, "-->") {
			if !strings.HasPrefix(buff, "<!--") {
				data.VerboseLog("found comment end\n")
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
	if draft, ok := data.AllFiles[file].Meta["draft"]; !ok || draft != "true" {
		if _, ok := data.AllFiles[file].Meta["title"]; !ok {
			return fmt.Errorf("Did not find `title` metadata element")
		}
		if _, ok := data.AllFiles[file].Meta["description"]; !ok {
			return fmt.Errorf("Did not find `description` metadata element")
		}
		if _, ok := data.AllFiles[file].Meta["keywords"]; !ok {
			return fmt.Errorf("Did not find `keywords` metadata element")
		}
	}
	return nil
}

func FrontSummary(filter string) (int, string) {
	errorCount := 0
	errorString := ""
	for file, details := range data.AllFiles {
		if strings.HasPrefix(file, filter) {
			errorCount += details.FormatErrorCount
			errorString = fmt.Sprintf("%s%s", errorString, details.FormatErrors)
		}
	}
	return errorCount, errorString
}
