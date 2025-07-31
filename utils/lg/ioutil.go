package lg

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

var _ io.Reader = (*StripColorReader)(nil)

type StripColorReader struct {
	r io.Reader
}

func NewStripColorReader(r io.Reader) *StripColorReader {
	return &StripColorReader{
		r: r,
	}
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func (scr *StripColorReader) Read(p []byte) (int, error) {
	n, err := scr.r.Read(p)
	if n > 0 {
		line := re.ReplaceAll(p[:n], nil)
		if len(line) == len(p) {
			return n, err
		}
		copy(p[:len(line)], line)
		return len(line), err
	}
	return n, err
}

var _ io.Writer = (*StripColorWriter)(nil)

type StripColorWriter struct {
	w io.Writer
}

func NewStripColorWriter(w io.Writer) *StripColorWriter {
	return &StripColorWriter{
		w: w,
	}
}

func (scw *StripColorWriter) Write(p []byte) (int, error) {
	data := re.ReplaceAll(p, nil)
	n, err := scw.w.Write(data)
	if n == len(data) {
		n = len(p)
	}
	return n, err
}

var _ io.Reader = (*StripPrefixReader)(nil)

type StripPrefixReader struct {
	prefix  []byte
	scanner *bufio.Scanner
	r       io.Reader
}

func NewStripPrefixReader(r io.Reader, prefix []byte) *StripPrefixReader {
	return &StripPrefixReader{
		r:      r,
		prefix: prefix,
	}
}

func (scr *StripPrefixReader) Read(p []byte) (int, error) {
	if scr.scanner == nil {
		scr.scanner = bufio.NewScanner(scr.r)
	}
	if !scr.scanner.Scan() {
		defer func() { scr.scanner = nil }()
		if scr.scanner.Err() == nil {
			// io.EOF
			return 0, io.EOF
		}
		return 0, scr.scanner.Err()
	}

	line := scr.scanner.Bytes()
	if bytes.HasPrefix(line, scr.prefix) {
		// Skip, and get next line.
		return scr.Read(p)
	}
	n := copy(p, line)
	// TODO(yuheng): If one line exceed input buf, what should we do? store the buffer?
	if n < len(p) {
		p[n] = '\n'
		n++
	}
	return n, nil
}
