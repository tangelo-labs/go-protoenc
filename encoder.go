package protoenc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Known errors definition.
var (
	ErrEncoderNil                   = errors.New("encoder can not be nil")
	ErrEncoderNotRegistered         = errors.New("encoder for given message not registered")
	ErrEncoderNotAFunction          = errors.New("encoder must be a function")
	ErrEncoderInputLengthMissMatch  = errors.New("encoder must be a function that takes exactly one pointer argument")
	ErrEncoderOutputLengthMissMatch = errors.New("encoder func must return only one argument")
	ErrEncoderOutputInvalid         = errors.New("encoder function output not a PB message")
	ErrEncoderAlreadyRegistered     = errors.New("encoder function already registered")
	ErrEncoderOutNotAProto          = errors.New("the encoded result it's not a proto message")
)

var protoType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// Encoder defines an element capable of registering encoder functions and
// encoding any given messages.
type Encoder interface {
	// Register records an encoder function, which is a function that takes
	// one pointer argument and transform it to a corresponding protocol buffer
	// message.
	//
	// This method should panic when the provided encoder does not satisfy the
	// following requirements:
	//
	//	- It must one single input argument
	//	- It must return as first argument a pointer to an object implementing
	//    proto.Message interface.
	//
	// Usage Example:
	//
	// ```go
	// encoder := protoenc.NewEncoder()
	//
	// encoder.Register(func(msg *MyMessage) *MyProtoMessage {
	// 		return &MyProtoMessage{}, nil
	// })
	// ```
	//
	// Registration is usually done on init functions, so panic-ing will stop
	// the execution of a program earlier if an invalid encoder is registered.
	Register(encoderFn interface{}) Encoder

	// Encode encodes a value of previously registered type as a protocol
	// buffer using the registered encoder.
	Encode(msg interface{}) (proto.Message, error)
}

type encoder struct {
	encoders  map[reflect.Type]interface{}
	protoType reflect.Type // for optimization purposes
	mu        sync.RWMutex
}

// NewEncoder builds a new encoder instance.
func NewEncoder() Encoder {
	return &encoder{
		encoders:  make(map[reflect.Type]interface{}),
		protoType: reflect.TypeOf((*proto.Message)(nil)).Elem(),
	}
}

func (e *encoder) Register(encoderFn interface{}) Encoder {
	e.mu.Lock()
	defer e.mu.Unlock()

	encoderType := reflectEncoderFunc(encoderFn)
	argType := encoderType.In(0)

	if _, exists := e.encoders[argType]; exists {
		details := fmt.Sprintf("func (%s) %s { ... }", argType.String(), encoderType.Out(0).String())

		panic(fmt.Errorf("%w: %s", ErrEncoderAlreadyRegistered, details))
	}

	e.encoders[argType] = reflect.ValueOf(encoderFn)

	return e
}

func (e *encoder) Encode(msg interface{}) (proto.Message, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	msgType := reflect.TypeOf(msg)
	if msgType.Kind() == reflect.Ptr {
		val := reflect.ValueOf(msg)
		msgType = val.Type()
	}

	enc, ok := e.encoders[msgType]
	if !ok {
		return nil, fmt.Errorf("%w: unable to encode message of type `%T` into its proto equivalent", ErrEncoderNotRegistered, msg)
	}

	args := []reflect.Value{reflect.ValueOf(msg)}
	result := enc.(reflect.Value).Call(args)
	protoMsg, ok := result[0].Interface().(proto.Message)

	if !ok {
		return nil, fmt.Errorf("%w: unable to encode message of type `%T` into its proto equivalent", ErrEncoderOutNotAProto, msg)
	}

	return protoMsg, nil
}

func reflectEncoderFunc(encoderFn interface{}) reflect.Type {
	if encoderFn == nil {
		panic(ErrEncoderNil)
	}

	encoderType := reflect.TypeOf(encoderFn)

	if encoderType.Kind() != reflect.Func {
		panic(ErrEncoderNotAFunction)
	}

	if encoderType.NumIn() != 1 {
		panic(ErrEncoderInputLengthMissMatch)
	}

	if encoderType.NumOut() != 1 {
		panic(ErrEncoderOutputLengthMissMatch)
	}

	if !encoderType.Out(0).Implements(protoType) {
		panic(ErrEncoderOutputInvalid)
	}

	return encoderType
}
