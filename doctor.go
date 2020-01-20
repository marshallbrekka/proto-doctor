package pbdoctor

import (
	"github.com/golang/protobuf/proto"
)

func Doctor(data []byte, mutator Mutator) []byte {
	return iterateFieldsDr(data, mutator)
}

func iterateFieldsDr(data []byte, dr Mutator) []byte {
	result := make([]byte, 0, len(data))
	var field *Field

	buffer := NewBuffer(data)
	for {
		f, read := buffer.ReadField()
		if read == 0 {
			break
		}

		if f.Type == proto.WireBytes {
			subDr := dr.MessageMutator(f.Number)
			if subDr != nil {
				field = &Field{
					Number: f.Number,
					Type:   f.Type,
					Data:   iterateFieldsDr(f.Data, subDr),
				}
			} else {
				field = dr.Mutate(f)
			}
		}
		field = dr.Mutate(f)
		if field == nil {
			field = f
		}
		result = append(result, field.Serialize()...)
	}
	return result
}

// Given the proto tag, returns the field number and wire type.
//
// https://developers.google.com/protocol-buffers/docs/encoding#structure
//
// Each key in the streamed message is a varint with the value
// (field_number << 3) | wire_type
// in other words, the last three bits of the number store the wire type.
// (field_number << 3) | wire_type
func ParseTag(tag byte) (byte, byte) {
	return tag >> 3, tag & 0x7
}

func EncodeTag(fieldNumber, fieldType byte) byte {
	return fieldNumber<<3 | fieldType
}

func varIntLength(data []byte) int {
	for i, b := range data {
		if b&0x80 == 0x0 {
			return i + 1
		}
	}
	return 0
}
