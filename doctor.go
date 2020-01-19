package pbdoctor

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

func Doctor(data []byte, mutator Mutator) []byte {
	return iterateFieldsDr(data, mutator)
}

func iterateFieldsDr(data []byte, dr Mutator) []byte {
	result := make([]byte, 0, len(data))
	var field *Field

	for i := 0; i < len(data); {
		field = nil
		number, typ := ParseTag(data[i])
		var length int
		var value []byte
		switch typ {
		case 0:
			// varint
			length = varIntLength(data[i+1:])
			value = data[i+1 : i+1+length]
			field = dr.Mutate(&Field{number, typ, value})

		case 1:
			// 64bit
			length = 8
			value = data[i+1 : i+1+length]
			field = dr.Mutate(&Field{number, typ, value})

		case 5:
			// 32bit
			length = 4
			value = data[i+1 : i+1+length]
			field = dr.Mutate(&Field{number, typ, value})

		case 2:
			// length delim
			var vi uint64
			vi, length = proto.DecodeVarint(data[i+1:])
			if length == 0 {
				panic(fmt.Errorf("zero length varint"))
			}
			value = data[i+1+length : i+1+int(vi)+length]

			subDr := dr.MessageMutator(number)
			if subDr == nil {
				field = dr.Mutate(&Field{number, typ, value})
			} else {
				field = &Field{
					Number: number,
					Type:   typ,
					Data:   iterateFieldsDr(value, subDr),
				}
			}
			length = length + int(vi)
		}
		i = i + length + 1

		if field == nil {
			field = &Field{
				Number: number,
				Type:   typ,
				Data:   value,
			}
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
