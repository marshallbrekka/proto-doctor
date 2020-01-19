package pbdoctor

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

func TestDoctor(t *testing.T) {
	input := &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "string value",
		},
	}
	output := &structpb.Value{}
	expected := &structpb.Value{
		Kind: &structpb.Value_ListValue{
			ListValue: &structpb.ListValue{
				Values: []*structpb.Value{
					&structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: "string valuu",
						},
					},
				},
			},
		},
	}

	data, _ := proto.Marshal(input)
	mutated := Doctor(mutator, Dr{})

	err := proto.Unmarshal(mutated, output)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Errorf("Expected:\n\t%#+v\nActual:\n\t%#+v", expected, output)
	}
}

type Dr struct {
}

func (d Dr) MessageMutator(n byte) Mutator {
	return nil
}

func (d Dr) Mutate(f *Field) *Field {
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
		}
	}
	return nil
}