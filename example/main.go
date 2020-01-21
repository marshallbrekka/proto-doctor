package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	pbdoctor "github.com/marshallbrekka/proto-doctor"
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

func (d Dr) Mutate(f *pbdoctor.Field) (*pbdoctor.Field, error) {
	n := f.Number
	ft := f.Type
	data := f.Data
	if ft == 2 {
		fmt.Printf("field: %d, type: %d, length: %d, value: %s\n", n, ft, len(data), string(data))
	} else {
		fmt.Printf("field: %d, type: %d, length: %d, value: %x\n", n, ft, len(data), data)
	}
	if f.Number == 3 {
		data := make([]byte, 0)
		data = append(data, f.Data...)
		return &pbdoctor.Field{
			// list value is 7
			Number: 6,
			Type:   2,
			Data: pbdoctor.Field{
				Number: 1,
				Type:   2,
				Data: pbdoctor.Field{
					Number: f.Number,
					Type:   f.Type,
					Data:   append(data, []byte(" more")...),
				}.Serialize(),
			}.Serialize(),
		}, nil
	}
	return nil, nil
}

func main() {
	test := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"str": &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: "string value",
				},
			},
		},
	}

	data, _ := proto.Marshal(test)
	fmt.Printf("org: %x\n", data)
	spew.Dump(test)
	mutator := Dr{
		Sub: map[byte]Dr{
			// Field 1 is Struct.fields
			1: Dr{
				Sub: map[byte]Dr{
					// Field 2 is the value of a map field (Value type)
					2: Dr{},
				},
			},
		},
	}
	mutated, _ := pbdoctor.Doctor(data, mutator)
	fmt.Printf("mut: %x\n", mutated)

	err := proto.Unmarshal(mutated, test)
	if err != nil {
		panic(err)
	}
	spew.Dump(test)
}
