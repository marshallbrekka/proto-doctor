package pbdoctor

import (
	"io"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestBufferErrors(t *testing.T) {
	varInt8 := proto.EncodeVarint(8)
	testCases := []struct {
		Input []byte
		Error error
	}{
		// Happy path EOF
		{
			Input: []byte{},
			Error: io.EOF,
		},

		// Bad tag (7 is not a valid wire type)
		{
			Input: []byte{EncodeTag(1, 7)},
			Error: errInternalBadWireType,
		},

		// Tag ony, missing any data
		{
			Input: []byte{EncodeTag(1, proto.WireVarint)},
			Error: io.ErrUnexpectedEOF,
		},

		// fixed 32, but only 2 bytes
		{
			Input: []byte{EncodeTag(1, proto.WireFixed32), 1, 2},
			Error: io.ErrUnexpectedEOF,
		},

		// fixed 64, but only 4 bytes
		{
			Input: []byte{EncodeTag(1, proto.WireFixed64), 1, 2, 3, 4},
			Error: io.ErrUnexpectedEOF,
		},

		// WireBytes, but less than specified length
		{
			Input: []byte{EncodeTag(1, proto.WireBytes), varInt8[0], 1, 2, 3, 4, 5},
			Error: io.ErrUnexpectedEOF,
		},
	}

	for i, testCase := range testCases {
		b := NewBuffer(testCase.Input)
		_, _, err := b.ReadField()
		if err != testCase.Error {
			t.Errorf("[%01d] Expected err:\n\t%s\nActual err:\n\t%s", i, testCase.Error, err)
		}
	}
}

func TestBuffer(t *testing.T) {
	enc := proto.NewBuffer(nil)
	varInt8 := proto.EncodeVarint(8)
	sampleBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	enc.EncodeFixed32(8)
	fixed32 := enc.Bytes()

	enc.Reset()
	enc.EncodeFixed64(8)
	fixed64 := enc.Bytes()

	type expectedT struct {
		Field Field
		Read  int
	}

	testCases := []struct {
		Input    []byte
		Expected []expectedT
	}{
		// Happy path each type
		{
			Input: Field{1, proto.WireVarint, varInt8}.Serialize(),
			Expected: []expectedT{
				{
					Field: Field{1, proto.WireVarint, varInt8},
					Read:  1 + len(varInt8),
				},
			},
		},
		{
			Input: Field{1, proto.WireFixed32, fixed32}.Serialize(),
			Expected: []expectedT{
				{
					Field: Field{1, proto.WireFixed32, fixed32},
					Read:  1 + len(fixed32),
				},
			},
		},
		{
			Input: Field{1, proto.WireFixed64, fixed64}.Serialize(),
			Expected: []expectedT{
				{
					Field: Field{1, proto.WireFixed64, fixed64},
					Read:  1 + len(fixed64),
				},
			},
		},
		{
			Input: Field{1, proto.WireBytes, sampleBytes}.Serialize(),
			Expected: []expectedT{
				{
					Field: Field{1, proto.WireBytes, sampleBytes},
					Read:  1 + 1 + len(sampleBytes),
				},
			},
		},

		// Multiple fields
		{
			Input: concatBytes(
				Field{1, proto.WireVarint, varInt8}.Serialize(),
				Field{1, proto.WireFixed32, fixed32}.Serialize(),
				Field{1, proto.WireFixed64, fixed64}.Serialize(),
				Field{1, proto.WireBytes, sampleBytes}.Serialize(),
				Field{1, proto.WireVarint, varInt8}.Serialize(),
			),
			Expected: []expectedT{
				{
					Field: Field{1, proto.WireVarint, varInt8},
					Read:  1 + len(varInt8),
				},
				{
					Field: Field{1, proto.WireFixed32, fixed32},
					Read:  1 + len(fixed32),
				},
				{
					Field: Field{1, proto.WireFixed64, fixed64},
					Read:  1 + len(fixed64),
				},
				{
					Field: Field{1, proto.WireBytes, sampleBytes},
					Read:  1 + 1 + len(sampleBytes),
				},
				{
					Field: Field{1, proto.WireVarint, varInt8},
					Read:  1 + len(varInt8),
				},
			},
		},
	}

outer:
	for i, testCase := range testCases {
		b := NewBuffer(testCase.Input)

		read := make([]expectedT, 0)
		for {
			f, n, err := b.ReadField()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("[%01d] %s", i, err)
				break outer
			}
			read = append(read, expectedT{*f, n})
		}

		if !reflect.DeepEqual(read, testCase.Expected) {
			t.Errorf("[%01d] Expected\n\t%#+v\nActual\n\t%#+v", i, testCase.Expected, read)
		}
	}
}

func concatBytes(slices ...[]byte) []byte {
	result := make([]byte, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
