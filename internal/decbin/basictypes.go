package decbin

// basictypes.go contains functions for encoding and decoding ints, strings,
// bools, and objects that directly implement BinaryUnmarshaler and
// BinaryMarshaler.

import (
	"encoding"
	"fmt"
	"strings"
	"unicode/utf8"
)

// EncBool encodes the bool value as a slice of bytes. The value can later
// be decoded with DecBool. No type indicator is included in the output;
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

// DecBool decodes a bool value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecBool(data []byte) (bool, int, error) {
	if len(data) < 1 {
		return false, 0, fmt.Errorf("unexpected end of data")
	}

	if data[0] == 0 {
		return false, 1, nil
	} else if data[0] == 1 {
		return true, 1, nil
	} else {
		return false, 0, fmt.Errorf("unknown non-bool value")
	}
}

// EncInt encodes the int value as a slice of bytes. The value can later
// be decoded with DecInt. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will contain a byte specifying how many bytes the integer is in
// the least-significant nibble, with the sign encoded as the first bit,
// followed by that many bytes, with the exception of 0 which is encoded as a
// single 0-byte, and -1 which is encoded as 0x80.
//
// TODO: betta description of encoding for negative values and how the
// non-significant bytes are assumed to be either 0xff or 0x00.
func EncInt(i int) []byte {
	if i == 0 {
		return []byte{0x00}
	}

	negative := i < 0

	b1 := byte((i >> 56) & 0xff)
	b2 := byte((i >> 48) & 0xff)
	b3 := byte((i >> 40) & 0xff)
	b4 := byte((i >> 32) & 0xff)
	b5 := byte((i >> 24) & 0xff)
	b6 := byte((i >> 16) & 0xff)
	b7 := byte((i >> 8) & 0xff)
	b8 := byte(i & 0xff)

	parts := []byte{b1, b2, b3, b4, b5, b6, b7, b8}

	enc := []byte{}
	var hitMSB bool
	for i := range parts {
		if hitMSB {
			enc = append(enc, parts[i])
		} else if (!negative && parts[i] != 0x00) || (negative && parts[i] != 0xff) {
			enc = append(enc, parts[i])
			hitMSB = true
		}
	}

	byteCount := uint8(len(enc))

	// byteCount will never be more than 8 so we can encode sign info in most
	// significant bit
	if negative {
		byteCount |= 0x80
	}

	enc = append([]byte{byteCount}, enc...)

	return enc
}

// DecInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecInt(data []byte) (int, int, error) {
	if len(data) < 1 {
		return 0, 0, fmt.Errorf("data does not contain at least 1 byte")
	}

	byteCount := data[0]

	if byteCount == 0 {
		return 0, 1, nil
	}
	data = data[1:]

	// pull count and sign out of byteCount
	negative := byteCount&0x80 != 0
	byteCount &= 0x0f

	if len(data) < int(byteCount) {
		return 0, 0, fmt.Errorf("unexpected EOF")
	}

	intData := data[:byteCount]

	// put missing other bytes back in

	padByte := byte(0x00)
	if negative {
		padByte = 0xff
	}
	for len(intData) < 8 {
		// if we're negative, we need to pad with 0xff bytes, otherwise 0x00
		intData = append([]byte{padByte}, intData...)
	}

	// keep value as uint until we return so we avoid logical shift semantics
	var iVal uint
	iVal |= (uint(intData[0]) << 56)
	iVal |= (uint(intData[1]) << 48)
	iVal |= (uint(intData[2]) << 40)
	iVal |= (uint(intData[3]) << 32)
	iVal |= (uint(intData[4]) << 24)
	iVal |= (uint(intData[5]) << 16)
	iVal |= (uint(intData[6]) << 8)
	iVal |= (uint(intData[7]))

	return int(iVal), int(byteCount + 1), nil
}

// EncString encodes a string value as a slice of bytes. The value can
// later be decoded with DecString. Encoded string output starts with an
// integer (as encoded by EncInt) indicating the number of bytes following
// that make up the string, followed by that many bytes containing the string
// encoded as UTF-8.
//
// The output will be variable length; it will contain 8 bytes followed by the
// bytes that make up X characters, where X is the int value contained in the
// first 8 bytes. Due to the specifics of how UTF-8 strings are encoded, this
// may or may not be the actual number of bytes used.
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
