package nats

import "github.com/nats-io/nats.go/jetstream"

type BatchResult struct {
	AckMsgs []jetstream.Msg
	NakMsgs []jetstream.Msg
	Error   error
}

func (br *BatchResult) HandleBatchError(err error, index int, msgs ...jetstream.Msg) {
	br.Error = err
	br.NakMsgs = msgs[index:]
}
