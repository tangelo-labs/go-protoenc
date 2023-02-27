package main

import (
	"fmt"
	"time"

	"github.com/tangelo-labs/go-protoenc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	encoder := protoenc.NewEncoder()
	encoder.Register(func(ts time.Time) *timestamppb.Timestamp {
		return timestamppb.New(ts)
	})

	decoder := protoenc.NewDecoder()
	decoder.Register(func(ts *timestamppb.Timestamp) time.Time {
		return ts.AsTime()
	})

	{
		out, err := encoder.Encode(time.Now())
		if err != nil {
			panic(err)
		}

		fmt.Printf("encoded output: %s\n", out.(*timestamppb.Timestamp).AsTime().String())
	}

	{
		out, err := decoder.Decode(timestamppb.Now())
		if err != nil {
			panic(err)
		}

		fmt.Printf("decoded outout: %s\n", out.(time.Time).String())
	}
}
