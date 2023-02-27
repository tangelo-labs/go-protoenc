package protoenc_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-protoenc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDecode(t *testing.T) {
	now := time.Now()

	t.Run("GIVEN a decoder instance WHEN decoding by reference THEN a valid output is returned", func(t *testing.T) {
		decoder := protoenc.NewDecoder()
		decoder.Register(func(ts *timestamppb.Timestamp) time.Time {
			return ts.AsTime()
		})

		out, err := decoder.Decode(timestamppb.New(now))
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, now.Unix(), out.(time.Time).Unix())
	})
}

func TestRegisterDecoder(t *testing.T) {
	tests := []struct {
		decoderFunc interface{}
		panic       error
	}{
		{
			decoderFunc: nil,
			panic:       protoenc.ErrDecoderNil,
		},
		{
			decoderFunc: func() {},
			panic:       protoenc.ErrDecoderInputLengthMissMatch,
		},
		{
			decoderFunc: func(*input) {},
			panic:       protoenc.ErrDecoderOutputLengthMissMatch,
		},
		{
			decoderFunc: func(*input) *input { return nil },
			panic:       protoenc.ErrDecoderInputInvalid,
		},
		{
			decoderFunc: func(*timestamppb.Timestamp) *input { return nil },
			panic:       nil,
		},
		{
			decoderFunc: func(*timestamppb.Timestamp) input { return input{} },
			panic:       nil,
		},
		{
			decoderFunc: func(proto.Message) *input { return nil },
			panic:       nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			encoder := protoenc.NewDecoder()

			if tt.panic != nil {
				require.PanicsWithError(t, tt.panic.Error(), func() {
					encoder.Register(tt.decoderFunc)
				})
			} else {
				encoder.Register(tt.decoderFunc)
			}
		})
	}
}
