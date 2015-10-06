package linereader

import (
	"bufio"
	"os"
	"strings"
)

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
