// Package format contains functions for producing a [CodeReader] from a stream
// that contains markdown files with fishi codeblocks. A CodeReader can be sent
// directly to the frontend and handles all gathering of codeblocks and running
// any preprocessing needed on it.
package format

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"
)

var (
	// should match solitary # on line, but we'd 8etter make sure there's a
	// newline for the [^#] after the # to match on, otherwise this won't be
	// gr8.
	regexCommentStart = regexp.MustCompile(`(?:^|[^#]+)(#)[^#]`)
)

// CodeReader is an implementation of io.Reader that reads fishi code from input
// containing markdown-formatted text with fishi codeblocks. It will gather all
// fishi codeblocks immediately on open and then read bytes from them as Read is
// called. Preprocessing may also be done at that time. The CodeReader will
// return io.EOF when all bytes from fishi codeblocks in the stream have been
// read.
type CodeReader struct {
	r *bytes.Reader
}

// Read reads bytes from the CodeReader. It will return io.EOF when all bytes
// from fishi codeblocks in the stream have been read. It cannot return an
// error as the actual underlying stream it was opened on is fully consumed at
// the time of opening.
func (cr *CodeReader) Read(p []byte) (n int, err error) {
	return cr.r.Read(p)
}

// NewCodeReader creates a new CodeReader from a stream containing markdown
// formatted text with fishi codeblocks. It will immediately read the provided
// stream until it returns EOF and find all fishi codeblocks and run
// preprocessing on them.
//
// Returns non-nil error if there is a problem reading the markdown or
// preprocessing the code.
func NewCodeReader(r io.Reader) (*CodeReader, error) {
	// read the whole stream into a buffer
	allInput := make([]byte, 0)

	bufReader := make([]byte, 256)
	var err error
	for err != io.EOF {
		var n int
		n, err = r.Read(bufReader)

		if n > 0 {
			allInput = append(allInput, bufReader[:n]...)
		}

		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	gatheredFishi := scanMarkdownForFishiBlocks(allInput)
	fishiSource := normalizeFishi(gatheredFishi)

	cr := &CodeReader{
		r: bytes.NewReader(fishiSource),
	}

	return cr, nil
}

// normalizeFishi does a preprocess step on the source, which as of now includes
// stripping comments and normalizing end of lines to \n.
func normalizeFishi(source []byte) []byte {
	toBuf := make([]byte, len(source))
	copy(toBuf, source)
	scanner := bufio.NewScanner(bytes.NewBuffer(toBuf))
	var preprocessed strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "\r\n") || strings.HasPrefix(line, "\n\r") {
			line = line[0 : len(line)-2]
		} else {
			line = strings.TrimSuffix(line, "\n")
		}

		// do *not* take double #'s.
		// we add a \n because that makes the regex match on # at line end.
		indexes := regexCommentStart.FindStringSubmatchIndex(line + "\n")

		if len(indexes) > 1 {
			commentStartIdx := indexes[2]
			if commentStartIdx >= 0 {
				line = line[:commentStartIdx]
			}
		}

		// now replace any double #'s with normal ones:
		line = strings.ReplaceAll(line, "##", "#")

		preprocessed.WriteString(line)
		preprocessed.WriteRune('\n')
	}

	return []byte(preprocessed.String())
}
