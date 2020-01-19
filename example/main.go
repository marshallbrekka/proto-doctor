package main

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	pbdoctor "github.com/marshallbrekka/protobuf-doctor"
)

type Dr struct {
	Sub map[byte]Dr
}

func (d Dr) MessageMutator(n byte) pbdoctor.Mutator {
	s, ok := d.Sub[n]
	if ok {
		fmt.Printf("sub field: %d\n", n)
		return s
	}
	return nil
}

func (d Dr) Mutate(f *pbdoctor.Field) *pbdoctor.Field {
	n := f.Number
	ft := f.Type
	data := f.Data
	if ft == 2 {
		fmt.Printf("field: %d, type: %d, length: %d, value: %s\n", n, ft, len(data), string(data))
	} else {
		fmt.Printf("field: %d, type: %d, length: %d, value: %x\n", n, ft, len(data), data)
	}
	return nil
}

func main() {
	test := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"str": &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: "string value",
				},
			},
			"lst": &structpb.Value{
				Kind: &structpb.Value_ListValue{
					ListValue: &structpb.ListValue{
						Values: []*structpb.Value{
							&structpb.Value{
								Kind: &structpb.Value_StringValue{
									StringValue: "list value",
								},
							},
						},
					},
				},
			},
		},
	}

	data, _ := proto.Marshal(test)
	fmt.Printf("%x\n", data)
	mutator := Dr{
		Sub: map[byte]Dr{
			// Field 1 is Struct.fields
			1: Dr{
				Sub: map[byte]Dr{
					// Field 2 is the value of a map field (Value type)
					2: Dr{
						Sub: map[byte]Dr{
							// Field 6 is Value.list_value
							6: Dr{
								Sub: map[byte]Dr{
									// Field 1 is Value.values
									// At this point we are back at a Value type.
									1: Dr{},
								},
							},
						},
					},
				},
			},
		},
	}
	pbdoctor.Doctor(mutator, data)
}
