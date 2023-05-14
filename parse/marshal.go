package parse

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dekarrin/ictiobus/internal/rezi"
)

// EncodeBytes takes a Parser and encodes it as a binary value (using an
// internal binary format called 'REZI').
func EncodeBytes(p Parser) []byte {
	data := rezi.EncString(p.Type().String())
	data = append(data, rezi.EncBinary(p)...)
	return data
}

// DecodeBytes takes a slice of bytes containing a Parser encoded as a binary
// value (using an internal binary format called 'REZI') and decodes it into the
// Parser it represents.
func DecodeBytes(data []byte) (p Parser, err error) {
	// first get the string giving the type
	typeStr, n, err := rezi.DecString(data)
	if err != nil {
		return nil, fmt.Errorf("read parser type: %w", err)
	}
	parserType, err := ParseAlgorithm(typeStr)
	if err != nil {
		return nil, fmt.Errorf("decode parser type: %w", err)
	}

	// set var's concrete type by getting an empty copy
	switch parserType {
	case AlgoLL1:
		p = EmptyLL1Parser()
	case AlgoSLR1:
		p = EmptySLR1Parser()
	case AlgoLALR1:
		p = EmptyLALR1Parser()
	case AlgoCLR1:
		p = EmptyCLR1Parser()
	default:
		panic("should never happen: parsed parserType is not valid")
	}

	_, err = rezi.DecBinary(data[n:], p)
	return p, err
}

// WriteFile stores the parser in a binary file (encoded using an internal
// format called 'REZI'). The Parser can later be retrieved from the file by
// calling [ReadFile] on it.
func WriteFile(p Parser, filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	bufWriter := bufio.NewWriter(fp)

	allBytes := EncodeBytes(p)
	_, err = bufWriter.Write(allBytes)
	if err != nil {
		return err
	}
	err = bufWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

// ReadFile retrieves a Parser by reading a file containing one encoded as
// binary bytes. This will read files created with [WriteFile].
func ReadFile(filename string) (Parser, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p, err := DecodeBytes(data)
	if err != nil {
		return nil, err
	}

	return p, nil
}
