package pbdoctor

import (
	"io"

	"github.com/golang/protobuf/proto"
)

// Calls the provided mutator for every field in the serialized protobuf message,
// modifying the byte array with any of the modifications the mutator makes.
//
// Returns an error if the provided byte array is not valid, or if the mutator
// ever returns an error.
//
// If you provide a noop mutator, then this is essentially the same as iterating
// through the structure with a Buffer, and re-assembling by calling Serialize()
// on each field.
//
//   buf := NewBuffer(serialized)
//   result := make([]byte, 0)
//   for {
//     field, _, err := buf.ReadField()
//     if err == io.EOF {
//       break
//     }
//     result = append(result, field.Serialize())
//   }
func Doctor(data []byte, mutator Mutator) ([]byte, error) {
	return iterateFieldsDr(data, mutator)
}

func iterateFieldsDr(data []byte, dr Mutator) ([]byte, error) {
	result := make([]byte, 0, len(data))
	var field *Field

	buffer := NewBuffer(data)
	for {
		f, _, err := buffer.ReadField()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if f.Type == proto.WireBytes {
			subDr := dr.MessageMutator(f.Number)
			if subDr != nil {
				subData, err := iterateFieldsDr(f.Data, subDr)
				if err != nil {
					return nil, err
				}
				field = &Field{
					Number: f.Number,
					Type:   f.Type,
					Data:   subData,
				}
			} else {
				field, err = dr.Mutate(f)
			}
		}
		field, err = dr.Mutate(f)
		if err != nil {
			return nil, err
		}
		if field == nil {
			field = f
		}
		result = append(result, field.Serialize()...)
	}
	return result, nil
}

// Given a byte array reads a varint and returns the field number, wire type,
// and the number of bytes read.
//
// https://developers.google.com/protocol-buffers/docs/encoding#structure
//
// Each key in the streamed message is a varint with the value
// (field_number << 3) | wire_type
// in other words, the last three bits of the number store the wire type.
// (field_number << 3) | wire_type
func ParseTag(b []byte) (uint64, byte, int) {
	x, n := proto.DecodeVarint(b)
	if n == 0 {
		return 0, 0, 0
	}
	return x >> 3, byte(x) & 0x7, n
}

func EncodeTag(fieldNumber uint64, fieldType byte) []byte {
	return proto.EncodeVarint(fieldNumber<<3 | uint64(fieldType))
}

func varIntLength(data []byte) int {
	for i, b := range data {
		if b&0x80 == 0x0 {
			return i + 1
		}
	}
	return 0
}
