package pbdoctor

import (
	"testing"

	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

func TestDoctor(t *testing.T) {
	// This will transform the following serialized proto from
	// Value{ string_value = "string value" }
	//
	// into
	// Value{
	//   list_value = ListValue{
	//     values = [
	//       Value{ string_value = "string value" }
	//   ]
	// }
	input := &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "string value",
		},
	}
	expected := &structpb.Value{
		Kind: &structpb.Value_ListValue{
			ListValue: &structpb.ListValue{
				Values: []*structpb.Value{
					&structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: "string value",
						},
					},
				},
			},
		},
	}

	data, _ := proto.Marshal(input)
	mutated, err := Doctor(data, Dr{})
	if err != nil {
		t.Fatal(err)
	}

	output := &structpb.Value{}
	err = proto.Unmarshal(mutated, output)

	if err != nil {
		t.Fatal(err)
	}

	if !proto.Equal(expected, output) {
		t.Errorf("Expected:\n\t%#+v\nActual:\n\t%#+v", expected, output)
	}
}

type Dr struct {
}

func (d Dr) MessageMutator(n byte) Mutator {
	return nil
}

func (d Dr) Mutate(f *Field) (*Field, error) {
	if f.Number == 3 {
		return &Field{
			// list value is 7
			Number: 6,
			Type:   2,
			Data: Field{
				Number: 1,
				Type:   2,
				Data: Field{
					Number: f.Number,
					Type:   f.Type,
					Data:   f.Data,
				}.Serialize(),
			}.Serialize(),
		}, nil
	}
	return nil, nil
}
