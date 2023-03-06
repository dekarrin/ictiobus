// Package marshal contains functions for marshaling data as binary bytes using
// a simple encoding scheme. Encoding output length is predictible. For ints and
// bools, this is accomplished by having constant length of encoded output (8
// bytes for ints, 1 byte for bools); for types with variable length encoded
// values such as strings or other types, this is accomplished by placing an
// encoded integer at the head of the bytes which indicates how many bytes that
// follow are part of the type being decoded.
package decbin

// TODO: rename this decbin

import (
	"encoding"
	"fmt"
	"reflect"
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

// EncBool encodes the bool value as a slice of bytes. The value can later
// be decoded with DecBinaryBool. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 1 byte.
func EncBool(b bool) []byte {
	enc := make([]byte, 1)

	if b {
		enc[0] = 1
	} else {
		enc[0] = 0
	}

	return enc
}

// EncInt encodes the int value as a slice of bytes. The value can later
// be decoded with DecBinaryInt. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 8 bytes.
func EncInt(i int) []byte {
	b1 := byte((i >> 52) & 0xff)
	b2 := byte((i >> 48) & 0xff)
	b3 := byte((i >> 40) & 0xff)
	b4 := byte((i >> 32) & 0xff)
	b5 := byte((i >> 24) & 0xff)
	b6 := byte((i >> 16) & 0xff)
	b7 := byte((i >> 8) & 0xff)
	b8 := byte(i & 0xff)
	enc := []byte{b1, b2, b3, b4, b5, b6, b7, b8}
	return enc
}

// EncString encodes a string value as a slice of bytes. The value can
// later be decoded with DecBinaryString. Encoded string output starts with an
// integer (as encoded by EncBinaryInt) indicating the number of bytes following
// that make up the string, followed by that many bytes containing the string
// encoded as UTF-8.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
func EncString(s string) []byte {
	enc := make([]byte, 0)

	chCount := 0
	for _, ch := range s {
		chBuf := make([]byte, utf8.UTFMax)
		byteLen := utf8.EncodeRune(chBuf, ch)
		enc = append(enc, chBuf[:byteLen]...)
		chCount++
	}

	countBytes := EncInt(chCount)
	enc = append(countBytes, enc...)

	return enc
}

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

// Order of keys in output is gauranteed to be consistent.
func EncMapStringToBinary[E encoding.BinaryMarshaler](m map[string]E) []byte {
	if m == nil {
		return EncInt(0)
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

// Order of keys in output is gauranteed to be consistent.
func EncMapStringToInt(m map[string]int) []byte {
	if m == nil {
		return EncInt(0)
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

func DecMapStringToInt(data []byte) (map[string]int, int, error) {
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

	m := map[string]int{}

	for i := 0; i < toConsume; i++ {
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

func DecMapStringToBinary[E encoding.BinaryUnmarshaler](data []byte) (map[string]E, int, error) {
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

	m := map[string]E{}

	for i := 0; i < toConsume; i++ {
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

// EncBinary encodes a BinaryMarshaler as a slice of bytes. The value can later
// be decoded with DecBinary. Encoded output starts with an integer (as encoded
// by EncBinaryInt) indicating the number of bytes following that make up the
// object, followed by that many bytes containing the encoded value.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
func EncBinary(b encoding.BinaryMarshaler) []byte {
	enc, _ := b.MarshalBinary()

	enc = append(EncInt(len(enc)), enc...)

	return enc
}

// DecBool decodes a bool value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecBool(data []byte) (bool, int, error) {
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

// DecString decodes a string value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecString(data []byte) (string, int, error) {
	if len(data) < 8 {
		return "", 0, fmt.Errorf("unexpected end of data")
	}
	runeCount, _, err := DecInt(data)
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

// DecInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecInt(data []byte) (int, int, error) {
	if len(data) < 8 {
		return 0, 0, fmt.Errorf("data does not contain 8 bytes")
	}

	intData := data[:8]
	var iVal int
	iVal |= (int(intData[0]) << 52)
	iVal |= (int(intData[1]) << 48)
	iVal |= (int(intData[2]) << 40)
	iVal |= (int(intData[3]) << 32)
	iVal |= (int(intData[4]) << 24)
	iVal |= (int(intData[5]) << 16)
	iVal |= (int(intData[6]) << 8)
	iVal |= (int(intData[7]))

	return iVal, 8, nil
}

// DecBinary decodes a value at the start of the given bytes and calls
// UnmarshalBinary on the provided object with those bytes.
//
// It returns the total number of bytes read from the data bytes.
func DecBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	var readBytes int
	var byteLen int
	var err error

	byteLen, readBytes, err = DecInt(data)
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

// get zero value for a type if not a pointer, or a pointer to a valid 0 value
// if a pointer type.
func initType[E any]() E {
	var v E

	vType := reflect.TypeOf(v)

	if vType.Kind() == reflect.Pointer {
		pointedTo := vType.Elem()
		pointedVal := reflect.New(pointedTo)
		pointedIFace := pointedVal.Interface()
		var ok bool
		v, ok = pointedIFace.(E)
		if !ok {
			// should never happen
			panic("could not convert returned type")
		}
	}

	return v
}
