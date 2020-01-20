package pbdoctor

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

// Iterates through a serialized protocol buffer field by field.
// Decscending into sub-messages requires initializing a new
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

// Reads the next field from the buffer. If the buffer is empty the bytes read will
// be zero.
func (b *Buffer) ReadField() (field *Field, read int) {
	if b.read == len(b.data) {
		return nil, 0
	}

	number, typ := ParseTag(b.data[b.read])
	var length int
	var value []byte

	switch typ {
	case proto.WireVarint:
		// varint
		length = varIntLength(b.data[b.read+1:])
		value = b.data[b.read+1 : b.read+1+length]
		field = &Field{number, typ, value}

	case proto.WireFixed64:
		// 64bit
		length = 8
		value = b.data[b.read+1 : b.read+1+length]
		field = &Field{number, typ, value}

	case proto.WireFixed32:
		// 32bit
		length = 4
		value = b.data[b.read+1 : b.read+1+length]
		field = &Field{number, typ, value}

	case proto.WireBytes:
		// length delim
		var vi uint64
		vi, length = proto.DecodeVarint(b.data[b.read+1:])
		if length == 0 {
			panic(fmt.Errorf("zero length varint"))
		}
		value = b.data[b.read+1+length : b.read+1+int(vi)+length]
		length = length + int(vi)
		field = &Field{number, typ, value}
	}
	b.read = b.read + length + 1
	return field, length + 1
}
