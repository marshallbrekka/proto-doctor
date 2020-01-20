package mutator

import (
	pbdoctor "github.com/marshallbrekka/protobuf-doctor"
)

type TreeMutator struct {
	Children map[byte]TreeMutator
	Fields   []byte
	Mutator  func(f *pbdoctor.Field) *pbdoctor.Field
}

func (m TreeMutator) MessageMutator(fieldNumber byte) pbdoctor.Mutator {
	child, ok := m.Children[fieldNumber]
	if ok {
		return child
	}
	return nil
}

func (m TreeMutator) Mutate(f *pbdoctor.Field) *pbdoctor.Field {
	for _, num := range m.Fields {
		if num == f.Number {
			return m.Mutator(f)
		}
	}
	return nil
}
