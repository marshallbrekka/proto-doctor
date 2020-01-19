package doctor

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

// Represents a serialized proto field.
//
// Example:
// Given the proto structure
//
// message MyMessage {
//   string name = 4;
// }
//
// If serialized with name = "John"
//
// Field{
//   Number: 4,
//   Type: 2,
//   Data: []byte("John"),
// }
type Field struct {
	// The field number is what is specified in the .proto file.
	Number byte

	// Type can be one of 0-5.
	// For more details see the encoding docs:
	// https://developers.google.com/protocol-buffers/docs/encoding#structure
	Type byte

	// The serialized form of the field.
	Data []byte
}

type Mutator interface {
	// When a field is length delimited (bytes, string, message) this method is called
	// with the field number. If the field is a message type you can return
	// a Mutator that will be called for each of the detected submessage fields.
	MessageMutator(fieldNumber byte) Mutator

	// Called for every field in the message.
	//
	// If a non-nil Field is returned it will replace the original
	// data in the serialized proto.
	Mutate(*Field) *Field
}

func Doctor(mutator Mutator, data []byte) []byte {
	// TODO: actually allow mutations
	iterateFieldsDr(data, mutator)
	return data
}

func iterateFieldsDr(data []byte, dr Mutator) {
	for i := 0; i < len(data); {
		//		fmt.Printf("starting with i %d, len: %d\n", i, len(data))
		number, typ := ParseTag(data[i])
		var length int
		switch typ {

		case 0:
			// varint
			length = varIntLength(data[i+1:])
			dr.Mutate(&Field{number, typ, data[i+1 : i+1+length]})

		case 1:
			// 64bit
			length = 8
			dr.Mutate(&Field{number, typ, data[i+1 : i+1+length]})

		case 2:
			// length delim
			var vi uint64
			vi, length = proto.DecodeVarint(data[i+1:])
			if length == 0 {
				panic(fmt.Errorf("zero length varint"))
			}
			subData := data[i+1+length : i+1+int(vi)+length]

			subDr := dr.MessageMutator(number)
			if subDr == nil {
				dr.Mutate(&Field{number, typ, subData})
			} else {
				iterateFieldsDr(subData, subDr)
			}
			length = length + int(vi)

		case 5:
			// 32bit

			length = 4
			dr.Mutate(&Field{number, typ, data[i+1 : i+1+length]})
		}
		i = i + length + 1
	}
}

// Given the proto tag, returns the field number and wire type.
//
// https://developers.google.com/protocol-buffers/docs/encoding#structure
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
