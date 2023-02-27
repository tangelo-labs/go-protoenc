package protoenc_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-protoenc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEncode(t *testing.T) {
	now := time.Now()

	t.Run("GIVEN an encoder instance WHEN encoding by reference THEN a valid output is returned", func(t *testing.T) {
		encoder := protoenc.NewEncoder()
		encoder.Register(func(*input) *timestamppb.Timestamp {
			return timestamppb.New(now)
		})

		out, err := encoder.Encode(&input{})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, now.Unix(), out.(*timestamppb.Timestamp).AsTime().Unix())
	})

	t.Run("GIVEN an encoder instance WHEN encoding by value THEN a valid output is returned", func(t *testing.T) {
		encoder := protoenc.NewEncoder()
		encoder.Register(func(input) *timestamppb.Timestamp {
			return timestamppb.New(now)
		})

		out, err := encoder.Encode(input{})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, now.Unix(), out.(*timestamppb.Timestamp).AsTime().Unix())
	})

	t.Run("GIVEN an encoder instance WHEN encoder expects a value AND a pointer is given THEN encoding fails", func(t *testing.T) {
		encoder := protoenc.NewEncoder()
		encoder.Register(func(input) *timestamppb.Timestamp {
			return timestamppb.New(now)
		})

		out, err := encoder.Encode(&input{})
		require.Error(t, err)
		require.Nil(t, out)
	})

	t.Run("GIVEN an encoder instance WHEN encoder expects a pinter AND a value is given THEN encoding fails", func(t *testing.T) {
		encoder := protoenc.NewEncoder()
		encoder.Register(func(*input) *timestamppb.Timestamp {
			return timestamppb.New(now)
		})

		out, err := encoder.Encode(input{})
		require.Error(t, err)
		require.Nil(t, out)
	})
}

func TestRegisterEncoder(t *testing.T) {
	tests := []struct {
		encoderFunc interface{}
		panic       error
	}{
		{
			encoderFunc: nil,
			panic:       protoenc.ErrEncoderNil,
		},
		{
			encoderFunc: func() {},
			panic:       protoenc.ErrEncoderInputLengthMissMatch,
		},
		{
			encoderFunc: func(*input) {},
			panic:       protoenc.ErrEncoderOutputLengthMissMatch,
		},
		{
			encoderFunc: func(*input) *input { return nil },
			panic:       protoenc.ErrEncoderOutputInvalid,
		},
		{
			encoderFunc: func(*input) *emptypb.Empty { return nil },
			panic:       nil,
		},
		{
			encoderFunc: func(input) *emptypb.Empty { return nil },
			panic:       nil,
		},
		{
			encoderFunc: func(*input) proto.Message { return nil },
			panic:       nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			encoder := protoenc.NewEncoder()

			if tt.panic != nil {
				require.PanicsWithError(t, tt.panic.Error(), func() {
					encoder.Register(tt.encoderFunc)
				})
			} else {
				encoder.Register(tt.encoderFunc)
			}
		})
	}
}

type input struct{}
