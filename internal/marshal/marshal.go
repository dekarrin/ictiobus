// Package marshal contains functions for marshaling data as binary bytes using
// a simple encoding scheme. Encoding output length is predictible. For ints and
// bools, this is accomplished by having constant length of encoded output (8
// bytes for ints, 1 byte for bools); for types with variable length encoded
// values such as strings or other types, this is accomplished by placing an
// encoded integer at the head of the bytes which indicates how many bytes that
// follow are part of the type being decoded.
package marshal

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// NewBinaryEncoder creates an Encoder that can encode to bytes and uses an
// object's MarshalBinary method to encode non-trivial types.
func NewBinaryEncoder() Encoder[encoding.BinaryMarshaler] {
	enc := &simpleBinaryEncoder{}
	return enc
}

// NewBinaryDecoder creates a Decoder that can decode bytes and uses an object's
// UnmarshalBinary method to decode non-trivial types.
func NewBinaryDecoder() Decoder[encoding.BinaryUnmarshaler] {
	dec := &simpleBinaryDecoder{}
	return dec
}

// EncBinaryBool encodes the bool value as a slice of bytes. The value can later
// be decoded with DecBinaryBool. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 1 byte.
func EncBinaryBool(b bool) []byte {
	enc := make([]byte, 1)

	if b {
		enc[0] = 1
	} else {
		enc[0] = 0
	}

	return enc
}

// EncBinaryInt encodes the int value as a slice of bytes. The value can later
// be decoded with DecBinaryInt. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 8 bytes.
func EncBinaryInt(i int) []byte {
	enc := make([]byte, 8)
	enc = binary.AppendVarint(enc, int64(i))
	return enc
}

// EncBinaryString encodes a string value as a slice of bytes. The value can
// later be decoded with DecBinaryString. Encoded string output starts with an
// integer (as encoded by EncBinaryInt) indicating the number of bytes following
// that make up the string, followed by that many bytes containing the string
// encoded as UTF-8.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
func EncBinaryString(s string) []byte {
	enc := make([]byte, 0)

	chCount := 0
	for _, ch := range s {
		chBuf := make([]byte, utf8.UTFMax)
		byteLen := utf8.EncodeRune(chBuf, ch)
		enc = append(enc, chBuf[:byteLen]...)
		chCount++
	}

	countBytes := EncBinaryInt(chCount)
	enc = append(countBytes, enc...)

	return enc
}

// Order of keys in output is gauranteed to be consistent.
func EncBinaryStringIntMap(m map[string]int) []byte {
	if m == nil {
		return EncBinaryInt(0)
	}

	enc := make([]byte, 0)

	keys := textfmt.OrderedKeys(m)
	for i := range keys {
		enc = append(enc, EncBinaryString(keys[i])...)
		enc = append(enc, EncBinaryInt(m[keys[i]])...)
	}

	enc = append(EncBinaryInt(len(enc)), enc...)
	return enc
}

func DecBinaryStringIntMap(data []byte) (map[string]int, int, error) {
	var totalConsumed int

	toConsume, n, err := DecBinaryInt(data)
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

	m := map[string]int{}

	for i := 0; i < toConsume; i++ {
		k, n, err := DecBinaryString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v, n, err := DecBinaryInt(data)
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

// EncBinary encodes a BinaryMarshaler as a slice of bytes. The value can later
// be decoded with DecBinary. Encoded output starts with an integer (as encoded
// by EncBinaryInt) indicating the number of bytes following that make up the
// object, followed by that many bytes containing the encoded value.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
func EncBinary(b encoding.BinaryMarshaler) []byte {
	enc, _ := b.MarshalBinary()

	enc = append(EncBinaryInt(len(enc)), enc...)

	return enc
}

// DecBinaryBool decodes a bool value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecBinaryBool(data []byte) (bool, int, error) {
	if len(data) < 1 {
		return false, 0, fmt.Errorf("unexpected end of data")
	}

	if data[0] == 0 {
		return false, 0, nil
	} else if data[0] == 1 {
		return true, 0, nil
	} else {
		return false, 0, fmt.Errorf("unknown non-bool value")
	}
}

// DecBinaryString decodes a string value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecBinaryString(data []byte) (string, int, error) {
	if len(data) < 8 {
		return "", 0, fmt.Errorf("unexpected end of data")
	}
	runeCount, _, err := DecBinaryInt(data)
	if err != nil {
		return "", 0, fmt.Errorf("decoding string rune count: %w", err)
	}
	data = data[8:]

	if runeCount < 0 {
		return "", 0, fmt.Errorf("string rune count < 0")
	}

	readBytes := 8

	var sb strings.Builder

	for i := 0; i < runeCount; i++ {
		ch, bytesRead := utf8.DecodeRune(data)
		if ch == utf8.RuneError {
			if bytesRead == 0 {
				return "", 0, fmt.Errorf("unexpected end of data in string")
			} else if bytesRead == 1 {
				return "", 0, fmt.Errorf("invalid UTF-8 encoding in string")
			} else {
				return "", 0, fmt.Errorf("invalid unicode replacement character in rune")
			}
		}

		sb.WriteRune(ch)
		readBytes += bytesRead
		data = data[bytesRead:]
	}

	return sb.String(), readBytes, nil
}

// DecBinaryInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecBinaryInt(data []byte) (int, int, error) {
	if len(data) < 8 {
		return 0, 0, fmt.Errorf("data does not contain 8 bytes")
	}

	val, read := binary.Varint(data[:8])
	if read == 0 {
		return 0, 0, fmt.Errorf("input buffer too small, should never happen")
	} else if read < 0 {
		return 0, 0, fmt.Errorf("input buffer contains value larger than 64 bits, should never happen")
	}
	return int(val), 8, nil
}

// DecBinary decodes a value at the start of the given bytes and calls
// UnmarshalBinary on the provided object with those bytes.
//
// It returns the total number of bytes read from the data bytes.
func DecBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	var readBytes int
	var byteLen int
	var err error

	byteLen, readBytes, err = DecBinaryInt(data)
	if err != nil {
		return 0, err
	}
	data = data[readBytes:]

	if len(data) < byteLen {
		return 0, fmt.Errorf("unexpected end of data")
	}
	binData := data[:byteLen]

	err = b.UnmarshalBinary(binData)
	if err != nil {
		return 0, err
	}

	return byteLen + readBytes, nil
}
