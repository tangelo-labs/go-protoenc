# Protoenc

This package provide a simple way for encoding/decoding protobuf messages from/to go structs.

## Installation

```bash
go get github.com/tangelo-labs/go-protoenc
```

## Usage

Encoding:

```go
encoder := protoenc.NewEncoder()

encoder.Register(func (ts time.Time) *timestamppb.Timestamp {
    return timestamppb.New(ts)
})

out, err := encoder.Encode(time.Now())
if err != nil {
    panic(err)
}

fmt.Println(out.(*timestamppb.Timestamp).AsTime().String())
```

Decoding:

```go
decoder := protoenc.NewDecoder()

decoder.Register(func (ts *timestamppb.Timestamp) time.Time {
    return ts.AsTime()
})

out, err := decoder.Decode(timestamppb.Now())
if err != nil {
    panic(err)
}

fmt.Println(out.(time.Time).String())
```