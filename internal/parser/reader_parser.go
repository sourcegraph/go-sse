package parser

import (
	"bufio"
	"io"

	"github.com/tmaxmax/go-sse/internal/util"
)

var splitFunc bufio.SplitFunc = func(data []byte, _ bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return
	}

	var start, index, endlineLen int
	for {
		index, endlineLen = NewlineIndex(data[advance:])
		advance += index + endlineLen
		if index == 0 {
			start += endlineLen
		}
		if advance == len(data) || (util.IsNewlineChar(data[advance]) && index > 0) {
			break
		}
	}

	if l := len(data); advance < l {
		advance++
		if advance < l && data[advance-1] == '\r' && data[advance] == '\n' {
			advance++
		}
	}

	token = data[start:advance]

	return
}

type ReaderParser struct {
	sc *bufio.Scanner
	p  *ByteParser
}

// Scan parses a single field from the reader. It returns false when there are no more fields to parse.
func (r *ReaderParser) Scan() bool {
	if !r.p.Scan() {
		if !r.sc.Scan() {
			return false
		}

		r.p.Reset(r.sc.Bytes())

		return r.p.Scan()
	}

	return true
}

// Field returns the last parsed field. The Field's Value byte slice isn't owned by the field, the underlying buffer
// is owned by the bufio.Scanner that is used by the ReaderParser.
func (r *ReaderParser) Field() Field {
	return r.p.Field()
}

// Err returns the last read error.
func (r *ReaderParser) Err() error {
	return r.sc.Err()
}

func NewReaderParser(r io.Reader) *ReaderParser {
	sc := bufio.NewScanner(util.RemovePrefix(r, "\xEF\xBB\xBF"))
	sc.Split(splitFunc)

	return &ReaderParser{sc: sc, p: NewByteParser(nil)}
}
