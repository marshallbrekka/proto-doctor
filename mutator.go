package pbdoctor

import "github.com/golang/protobuf/proto"

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

func (f Field) Serialize() []byte {
	result := make([]byte, 0, len(f.Data)+1)
	result = append(result, EncodeTag(f.Number, f.Type))

	// if its length delim, then write the length and then the data.
	// Otherwise we just write the data.
	if f.Type == 2 {
		result = append(result, proto.EncodeVarint(uint64(len(f.Data)))...)
	}
	return append(result, f.Data...)
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
