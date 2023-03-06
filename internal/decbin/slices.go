package decbin

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"encoding"
	"fmt"
)

func EncSliceString(sl []string) []byte {
	if sl == nil {
		return EncInt(0)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncString(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceString(data []byte) ([]string, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("not enough bytes")
	}

	sl := []string{}

	for i := 0; i < toConsume; i++ {
		s, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode item: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = append(sl, s)
	}

	return sl, totalConsumed, nil

}

func EncSliceBinary[E encoding.BinaryMarshaler](sl []E) []byte {
	if sl == nil {
		return EncInt(0)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncBinary(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceBinary[E encoding.BinaryUnmarshaler](data []byte) ([]E, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("not enough bytes")
	}

	sl := []E{}

	for i := 0; i < toConsume; i++ {
		v := initType[E]()

		n, err := DecBinary(data, v)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode item: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = append(sl, v)
	}

	return sl, totalConsumed, nil
}
