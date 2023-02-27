package protoenc

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Codec is a component capable of encoding and decoding proto messages
// into/from Go objects.
type Codec interface {
	// Register records an encoder and a decoder functions pair, responsible for
	// encoding and decoding a concrete Go object. This method will panic if the
	// provided functions are inconsistent of the input/output types.
	//
	// For example, if an encoder function expects a "MyMessage" as input, then
	// its counterpart decoder function must return a "MyMessage" as output:
	//
	// 	func encodeMyMessage(msg *MyMessage) *pb.MyMessage {
	// 		return &pb.MyMessage{}
	// 	}
	//
	// 	func decodeMyMessage(msg *pb.MyMessage) *MyMessage {
	// 		return &MyMessage{}
	// 	}
	Register(encoderFn interface{}, decoderFn interface{})

	// Encode encodes a Go object into a proto message.
	Encode(msg interface{}) (proto.Message, error)

	// Decode decodes a proto message into a Go object.
	Decode(msg proto.Message) (interface{}, error)
}

type codec struct {
	encoder Encoder
	decoder Decoder
}

// NewCodec builds a new codec instance.
func NewCodec() Codec {
	return &codec{
		encoder: NewEncoder(),
		decoder: NewDecoder(),
	}
}

func (c *codec) Register(encoderFn interface{}, decoderFn interface{}) {
	c.validatePair(encoderFn, decoderFn)

	c.encoder.Register(encoderFn)
	c.decoder.Register(decoderFn)
}

func (c *codec) Encode(msg interface{}) (proto.Message, error) {
	return c.encoder.Encode(msg)
}

func (c *codec) Decode(msg proto.Message) (interface{}, error) {
	return c.decoder.Decode(msg)
}

func (c *codec) validatePair(encoderFn interface{}, decoderFn interface{}) {
	encoderR := reflectEncoderFunc(encoderFn)
	decoderR := reflectDecoderFunc(decoderFn)

	if encoderR.In(0) != decoderR.Out(0) {
		panic(fmt.Errorf("input/output types do not match: %s != %s", encoderR.In(0), decoderR.Out(0)))
	}
}
