package pbdoctor

type Mutator interface {
	// When a field is length delimited (bytes, string, message) this method is called
	// with the field number. If the field is a message type you can return
	// a Mutator that will be called for each of the detected submessage fields.
	MessageMutator(fieldNumber byte) Mutator

	// Called for every field in the message.
	//
	// If a non-nil Field is returned it will replace the original
	// data in the serialized proto.
	Mutate(*Field) (*Field, error)
}
