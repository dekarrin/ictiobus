package decbin

// stringmaps.go contains functions for encoding and decoding maps of string to
// the basic types.

import (
	"encoding"
	"fmt"

	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// Order of keys in output is gauranteed to be consistent.
func EncMapStringToInt(m map[string]int) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	keys := textfmt.OrderedKeys(m)
	for i := range keys {
		enc = append(enc, EncString(keys[i])...)
		enc = append(enc, EncInt(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecMapStringToInt(data []byte) (map[string]int, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]int{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("not enough bytes")
	}

	m := map[string]int{}

	var i int
	for i < toConsume {
		k, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v, n, err := DecInt(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}

// Order of keys in output is gauranteed to be consistent.
func EncMapStringToBinary[E encoding.BinaryMarshaler](m map[string]E) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	keys := textfmt.OrderedKeys(m)
	for i := range keys {
		enc = append(enc, EncString(keys[i])...)
		enc = append(enc, EncBinary(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecMapStringToBinary[E encoding.BinaryUnmarshaler](data []byte) (map[string]E, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]E{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("not enough bytes")
	}

	m := map[string]E{}

	var i int
	for i < toConsume {
		k, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v := initType[E]()
		n, err = DecBinary(data, v)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}
