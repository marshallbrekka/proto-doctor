package pbdoctor

import "github.com/golang/protobuf/proto"

// Represents a serialized proto field.
//
// Example:
// Given the proto structure
//
//   message MyMessage {
//     string name = 4;
//   }
//
// If serialized with name = "John"
//
//   Field{
//     Number: 4,
//     Type: 2,
//     Data: []byte("John"),
//   }
type Field struct {
	// The field number is what is specified in the .proto file.
	Number uint64

	// Type can be one of 0-5.
	// For more details see the encoding docs:
	// https://developers.google.com/protocol-buffers/docs/encoding#structure
	Type byte

	// The serialized form of the field.
	Data []byte
}

// Serializes the field back into the protobuf format.
// For fields of type WireBytes the format is <tag><varint><data>
// For all other fields it is just <tag><data>
func (f Field) Serialize() []byte {
	var result []byte
	tag := EncodeTag(f.Number, f.Type)

	// If its length delim, then write the length and then the data.
	// Otherwise we just write the data.
	if f.Type == proto.WireBytes {
		length := proto.EncodeVarint(uint64(len(f.Data)))
		result = make([]byte, 0, len(tag)+len(length)+len(f.Data))
		result = append(result, tag...)
		result = append(result, length...)
	} else {
		result = make([]byte, 0, len(tag)+len(f.Data))
		result = append(result, tag...)
	}
	return append(result, f.Data...)
}
