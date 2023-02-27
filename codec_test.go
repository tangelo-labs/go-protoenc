package protoenc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-protoenc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCodec(t *testing.T) {
	codec := protoenc.NewCodec()
	now := time.Now()

	encoder := func(msg *myMessage) *timestamppb.Timestamp {
		return timestamppb.New(msg.ts)
	}

	decoder := func(msg *timestamppb.Timestamp) *myMessage {
		return &myMessage{
			ts: msg.AsTime(),
		}
	}

	codec.Register(encoder, decoder)

	p, err := codec.Encode(&myMessage{ts: now})
	require.NoError(t, err)

	d, err := codec.Decode(p)
	require.NoError(t, err)

	require.EqualValues(t, now.Unix(), d.(*myMessage).ts.Unix())
}

type myMessage struct {
	ts time.Time
}
