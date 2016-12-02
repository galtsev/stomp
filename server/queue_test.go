package server

import (
	"dan/stomp/frame"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSubscribeAndSend(t *testing.T) {
	// setup
	var (
		subscriptionId string = randomString(8)
		msgBody        []byte = randomBytes(8)
	)
	subscribeFrame := frame.New()
	subscribeFrame.Command = frame.CmdSubscribe
	subscribeFrame.Header.Set(frame.HdrId, subscriptionId)
	ch := make(chan frame.Frame, 4)
	options := SubscriptionOptions{
		ClientWriteChan: ch,
	}
	queue := NewQueue("fake")
	queue.Subscribe(*subscribeFrame, options)

	// frame to send
	fr := frame.New()
	fr.Command = frame.CmdMessage
	fr.Body = msgBody
	queue.Send(*fr)

	// check resulting frame
	select {
	case outFr := <-ch:
		assert.Equal(t, msgBody, outFr.Body, "body don't match")
		subsId, ok := outFr.Header.Get(frame.HdrSubscription)
		assert.Equal(t, true, ok, "Missing subscription header of MESSAGE message type")
		assert.Equal(t, subscriptionId, subsId, "SubscriptionId didn't match")
	case <-time.NewTimer(time.Millisecond).C:
		t.Error("timeout receiving message")
	}
}
