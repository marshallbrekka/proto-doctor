package pbdoctor

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestField(t *testing.T) {
	varInt8 := proto.EncodeVarint(8)
	sampleWireBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	testCases := []struct {
		Input    Field
		Expected []byte
	}{
		// Non-length delim
		{
			Input: Field{
				Number: 8,
				Type:   0,
				Data:   varInt8,
			},
			Expected: append([]byte{8 << 3}, varInt8...),
		},
		// Length delim
		{
			Input: Field{
				Number: 8,
				Type:   2,
				Data:   sampleWireBytes,
			},
			Expected: append(append([]byte{8<<3 | 2}, varInt8...), sampleWireBytes...),
		},
	}

	for i, testCase := range testCases {
		serialized := testCase.Input.Serialize()
		if !bytes.Equal(testCase.Expected, serialized) {
			t.Errorf("[%01d] Expected:\n\t%x\nActual:\n\t%x", i, testCase.Expected, serialized)
		}
	}
}
