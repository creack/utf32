package utf32

import "errors"

// UTF32 represents an UTF-32 character
type UTF32 uint32

// Index into the table below with the first byte of a UTF-8 sequence to
// get the number of trailing bytes that are supposed to follow it.
// Note that *legal* UTF-8 values can't have 4 or 5-bytes. The table is
// left as-is for anyone who may want to do such conversion, which was
// allowed in earlier algorithms.
var trailingBytesForUTF8 = [256]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5,
}

// Magic values subtracted from a buffer value during UTF8 conversion.
// This table contains as many values as there might be trailing bytes
// in a UTF-8 sequence.
var offsetsFromUTF8 = [6]UTF32{
	0x00000000,
	0x00003080,
	0x000E2080,
	0x03C82080,
	0xFA082080,
	0x82082080,
}

// Once the bits are split out into bytes of UTF-8, this is a mask OR-ed
// into the first byte, depending on how many bytes follow.  There are
// as many entries in this table as there are UTF-8 sequence types.
// (I.e., one byte sequence, two byte... etc.). Remember that sequencs
// for *legal* UTF-8 will be 4 or fewer bytes total.
var firstByteMark = [7]byte{0x00, 0x00, 0xC0, 0xE0, 0xF0, 0xF8, 0xFC}

// Bound constants.
const (
	UniSurHighStart  UTF32 = 0xd800
	UniSurLowEnd     UTF32 = 0xdfff
	UniMaxLegalUTF32 UTF32 = 0x0010ffff
)

// Mask constants.
const (
	byteMask = 0xbf
	byteMark = 0x80
)

// Common errors.
var (
	ErrInvalidSource = errors.New("illegal source")
)

func lookupBytesToWrite(ch UTF32) (int, error) {
	bytesToWrite := 0
	switch {
	case ch < 0x80:
		bytesToWrite = 1
	case ch < 0x800:
		bytesToWrite = 2
	case ch < 0x10000:
		bytesToWrite = 3
	case ch <= UniMaxLegalUTF32:
		bytesToWrite = 4
	default:
		return -1, ErrInvalidSource
	}
	return bytesToWrite, nil
}

// ConvertUTF32toUTF8 converts the given utf32 as a utf8 string.
func ConvertUTF32toUTF8(src []UTF32) (string, error) {
	// TODO: improve allocations.
	ret := make([]byte, 0, len(src)*4)
	idx := 0
	for _, ch := range src {
		// UTF-16 surrogate values are illegal in UTF-32.
		if ch >= UniSurHighStart && ch <= UniSurLowEnd {
			return "", ErrInvalidSource
		}

		// Figure out how many bytes the result will require.
		bytesToWrite, err := lookupBytesToWrite(ch)
		if err != nil {
			return "", err
		}

		// Extend `ret` length
		for i := 0; i < bytesToWrite; i++ {
			ret = append(ret, 0)
		}
		switch bytesToWrite {
		case 4:
			ret[idx+3] = byte((int(ch) | byteMark) & byteMask)
			ch >>= 6
			fallthrough
		case 3:
			ret[idx+2] = byte((int(ch) | byteMark) & byteMask)
			ch >>= 6
			fallthrough
		case 2:
			ret[idx+1] = byte((int(ch) | byteMark) & byteMask)
			ch >>= 6
			fallthrough
		case 1:
			ret[idx] = byte((int(ch) | int(firstByteMark[bytesToWrite])))
		}
		idx += bytesToWrite
	}
	return string(ret), nil
}

// ConvertUTF8toUTF32 converts the given utf-8 string to an utf-32 buffer.
func ConvertUTF8toUTF32(src string) ([]UTF32, error) {
	ret := []UTF32{}
	for i := 0; i < len(src); i++ {
		var extraBytesToRead = trailingBytesForUTF8[src[i]]

		if i+int(extraBytesToRead) >= len(src) {
			return nil, ErrInvalidSource
		}

		var ch UTF32
		for j := 0; j < int(extraBytesToRead); j++ {
			ch += UTF32(src[i])
			i++
			ch <<= 6
		}
		ch += UTF32(src[i]) - offsetsFromUTF8[extraBytesToRead]

		if ch > UniMaxLegalUTF32 || (ch >= UniSurHighStart && ch <= UniSurLowEnd) {
			return nil, ErrInvalidSource
		}
		ret = append(ret, ch)
	}
	return ret, nil
}
