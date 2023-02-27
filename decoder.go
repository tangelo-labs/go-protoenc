package protoenc

import (
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Known errors definition
var (
	ErrDecoderNil                   = errors.New("decoder can not be nil")
	ErrDecoderNotRegistered         = errors.New("decoder for given message not registered")
	ErrDecoderNotAFunction          = errors.New("decoder must be a function")
	ErrDecoderInputLengthMissMatch  = errors.New("decoder must be a function that takes exactly one input argument")
	ErrDecoderOutputLengthMissMatch = errors.New("decoder func must return only one argument")
	ErrDecoderInputInvalid          = errors.New("decoder function input does not implements proto.Message interface")
	ErrDecoderAlreadyRegistered     = errors.New("decoder function already registered")
)

// Decoder defines an element capable of registering decoder functions and
// decoding proto messages into Golang objects.
//
// This interface does the opposite of its counterpart Encoder.
type Decoder interface {
	// Register records a decoder, which is a function that takes one proto
	// message as input and transform it to a Golang object.
	//
	// This method WILL PANIC when the decoder being registered does not satisfy
	// the requirements.
	//
	// Registration is usually done in init functions, so panic-ing will stop
	// the execution of a program earlier if an invalid decoder is registered.
	//
	// Provided function must be a function that takes exactly one argument of
	// type proto.Message, and returns exactly one argument of any type.
	// Examples:
	//
	// 	func decodeMyMessage(msg *pb.MyMessage) *MyMessage {
	//		return &MyMessage{}
	//	}
	//
	// Above, the type "pb.MyMessage" is a struct that implements proto.Message
	// interface.
	Register(decoderFn interface{})

	// Decode decodes a proto messages into a Golang object using the registered
	// decoders.
	Decode(msg proto.Message) (interface{}, error)
}

type decoder struct {
	decoders map[reflect.Type]reflect.Value
}

// NewDecoder builds a new decoder instance.
func NewDecoder() Decoder {
	return &decoder{
		decoders: make(map[reflect.Type]reflect.Value),
	}
}

func (e *decoder) Register(decoderFn interface{}) {
	if decoderFn == nil {
		panic(ErrDecoderNil)
	}

	decoderType := reflectDecoderFunc(decoderFn)
	argType := decoderType.In(0)

	if _, exists := e.decoders[argType]; exists {
		details := fmt.Sprintf("fn(%s)", argType.Name())

		panic(fmt.Errorf("%w: %s", ErrDecoderAlreadyRegistered, details))
	}

	e.decoders[argType] = reflect.ValueOf(decoderFn)
}

func (e *decoder) Decode(msg proto.Message) (interface{}, error) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil {
		return nil, fmt.Errorf("%w: cannot decode a nil message", ErrDecoderNotRegistered)
	}

	if msgType.Kind() == reflect.Ptr {
		val := reflect.ValueOf(msg)
		msgType = val.Type()
	}

	dec, ok := e.decoders[msgType]
	if !ok {
		return nil, fmt.Errorf("%w: message type was `%T`", ErrDecoderNotRegistered, msg)
	}

	args := []reflect.Value{reflect.ValueOf(msg)}
	result := dec.Call(args)

	return result[0].Interface(), nil
}

func reflectDecoderFunc(decoderFn interface{}) reflect.Type {
	decoderType := reflect.TypeOf(decoderFn)

	if decoderType.Kind() != reflect.Func {
		panic(ErrDecoderNotAFunction)
	}

	if decoderType.NumIn() != 1 {
		panic(ErrDecoderInputLengthMissMatch)
	}

	if decoderType.NumOut() != 1 {
		panic(ErrDecoderOutputLengthMissMatch)
	}

	if !decoderType.In(0).Implements(protoType) {
		panic(ErrDecoderInputInvalid)
	}

	return decoderType
}
