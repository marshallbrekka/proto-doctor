# Proto Doctor [![GoDoc](https://godoc.org/github.com/marshallbrekka/proto-doctor?status.svg)](https://godoc.org/github.com/marshallbrekka/proto-doctor) [![CircleCI](https://circleci.com/gh/marshallbrekka/proto-doctor.svg?style=svg)](https://circleci.com/gh/marshallbrekka/proto-doctor)

A tool for manipulating serialized protocol buffers.

Not production ready, use at your own risk :)

## Why?

This is mostly a proof of concept to handle a very specific issue/design constraint, so I don't expect this to be particulary useful outside my own usecase.

The protos below are not exactly my use-case, but should be a good enough example to get the point across.

### My Usecase

```proto
message InternalString {
  string value = 1;
  map<string, string> metadata = 2;
}

message InternalInt32 {
  int32 value = 1;
  map<string, string> metadata = 2;
}


message MyRPCRequest {
  InternalString field_a = 1;
  InternalString field_b = 2;
  InternalInt32 field_c = 3;
}

message MyRPCReply {
  InternalString response_a = 1;
  InternalInt32 response_b = 2;
}
```

When requests are received from the public internet, those `Internal<Type>` messages have their `metadata` fields populated by an API-Gateway before being forwarded to the downstream service.

The metadata are not HTTP headers or anything, but state that is contextual to the user making the request.

Similarly on responses back to the public internet, the API- Gatewayany uses any `metadata` that is present to update some internal state, and then zero's out the `metadata` before sending the response out to the net.

Internally this works quite well, as all downstream systems have access to a bunch of information automatically, without having to invoke additional APIs or build the metadata retrieval access themselves.

Externally is another story...

#### Improving the Ergonomics
Externally this system makes little sense and has a couple drawbacks:
1. The API is more cumbersome to use, as fields that would have normally been a `string` or `int32` type are now nested inside wrapper objects
2. It leaks a little bit of implementation details if we were to start sharing our API protos with third parties who wanted to call our APIs.

Viewed as JSON, a normal request/response would look like:

```json
{
  "field_a": {
    "value": "value a"
  },
  "field_b": {
    "value": "value b"
  },
  "field_c": {
    "value": 123
  }
}
```

The goal is to have a public representation that unwrappes those fields:

```json
{
  "field_a": "value a",
  "field_b": "value b",
  "field_c": 123
}
```

In order to accomplish the above, I need two things:
1. Convert the internal .proto files into an external version that unwrappes the scalar types
2. Code that converts from the external proto to the wrapped internal protos (and vice versa).

Accomplishing the first is pretty straight forward (slightly smarter find/replace).

The second can be less straight forward.

#### Translating the protos

One way to accomplish the translation could be using reflection, or a compiler plugin that adds methods to convert from one type to the other, etc.

There is one issue with the Go proto implimentation that makes one part of this difficult:
*Messages are registered in a global namespace.*

So if we wanted to do the translation using the proto generated Go objects, we would also need to re-write the proto packages so that both the internal and external generated structs could be registered in memory at the same time.

This could be a very acceptable tradeoff, but I wanted to see if I could accomplish it without having to change package names.

This lead me to manipulating the serialized protos directly.

Looking again at the slightly simplified example above, its actually pretty straight forward.

If these are our internal protos:
```proto
message InternalString {
  string value = 1;
}

message MyRPCRequest {
  InternalString field_a = 3;
}
```

If these are the external protos:
```proto
message MyRPCRequest {
  string field_a = 3;
}
```

And we have an external request of:
```json
{
  "field_a": "value a"
}
```

Translating from _external_ to _internal_ is as simple as re-encoding the string in `field_a` as a proto length delimited field with a tag of `field number 1` and `wire type 2`.

Or by using the `Field` struct type in this library

```go
externalFieldA := Field{
  Number: 3,
  Type: proto.WireBytes,
  Data: []byte("value a"),
}

internalFieldA := Field{
  Number: 3,
  Type: proto.WireBytes,
  Data: Field{
    Number: 1,
    Type: proto.WireBytes,
    Data: []byte("value a"),
  }.Serialize(),
}
```
