package pbdoctor

import (
	"errors"
	"io"

	"github.com/golang/protobuf/proto"
)

var errInternalBadWireType = errors.New("proto: internal error: bad wiretype")

// Iterates through a serialized protocol buffer field by field.
// Descending into sub-messages requires initializing a new
// buffer with just the sub-message bytes.
type Buffer struct {
	read int
	data []byte
}

func NewBuffer(b []byte) *Buffer {
	return &Buffer{
		data: b,
	}
}

// Reads the next field from the buffer, and returns the amount of bytes read.
// If there are no more fields in the buffer an io.EOF error will be returned.
func (b *Buffer) ReadField() (field *Field, read int, err error) {
	if b.read == len(b.data) {
		return nil, 0, io.EOF
	}

	number, typ := ParseTag(b.data[b.read])
	var length int
	var value []byte

	// All fields are offset by minimum 1 (the tag).
	// Length delim types are 1 + <varint>
	startOffset := 1

	switch typ {
	case proto.WireVarint:
		length = varIntLength(b.data[b.read+1:])

	case proto.WireFixed64:
		length = 8

	case proto.WireFixed32:
		length = 4

	case proto.WireBytes:
		size, read := proto.DecodeVarint(b.data[b.read+startOffset:])
		if read == 0 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		length = int(size)
		startOffset += read

	default:
		return nil, 0, errInternalBadWireType
	}

	if length == 0 || b.read+startOffset+length > len(b.data) {
		return nil, 0, io.ErrUnexpectedEOF
	}

	value = b.data[b.read+startOffset : b.read+startOffset+length]
	field = &Field{number, typ, value}

	b.read = b.read + startOffset + length
	return field, startOffset + length, nil
}
