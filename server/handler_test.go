package server

import (
	"dan/stomp/frame"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestHandlerSubscribeAndSend(t *testing.T) {
	var (
		subscriptionId string = randomString(8)
		destination    string = "/queue/" + randomString(4)
		body           []byte = randomBytes(16)
	)
	// setup
	server := NewServer()
	client := &ClientConnection{
		WriteChan: make(chan frame.Frame, 4),
	}
	handler := NewHandler(client, server)

	// subscribe
	subscriptionFrame := frame.New()
	subscriptionFrame.Command = frame.CmdSubscribe
	subscriptionFrame.Header.Set(frame.HdrId, subscriptionId)
	subscriptionFrame.Header.Set(frame.HdrDestination, destination)
	handler.Handle(*subscriptionFrame)

	// send
	sendFrame := frame.New()
	sendFrame.Command = frame.CmdSend
	sendFrame.Header.Set(frame.HdrDestination, destination)
	sendFrame.Body = body
	handler.Handle(*sendFrame)

	// check output
	select {
	case outFr := <-client.WriteChan:
		assert.Equal(t, frame.CmdMessage, outFr.Command, "Wrong message type")
		assert.Equal(t, body, outFr.Body, "wrong message body")
		subId, _ := outFr.Header.Get(frame.HdrSubscription)
		assert.Equal(t, subscriptionId, subId, "wrong subscription id")
	case <-time.NewTimer(time.Millisecond).C:
		t.Error("Timeout getting dispatched message")
	}
}

func makeSubscriptionFrame(subId, destination string) *frame.Frame {
	fr := frame.New()
	fr.Command = frame.CmdSubscribe
	fr.Header.Set(frame.HdrId, subId)
	fr.Header.Set(frame.HdrDestination, destination)
	return fr
}

func makeSendFrame(destination, body string) *frame.Frame {
	fr := frame.New()
	fr.Command = frame.CmdSend
	fr.Header.Set(frame.HdrDestination, destination)
	fr.Body = []byte(body)
	return fr
}

type TestMsg struct {
	client int    // client which must receive this message
	msgId  string // for x-msg-id custom header
	sId    string // id of subscription, which must receive this message
	dest   string
	body   string
}

// client 1 subscribed to /queue/1 (sid1) and /queue/2 (sid2)
// client 2 send:
// 1 message to /queue/1
// 2 messages to /queue/2
// 1 message to /queue/3 (no subscriptions yet)
// client 2 subscribed to /queue/3 (sid3)
// check that all 4 messages delivered to correct destination
func TestDispatchThreeSubscriptions(t *testing.T) {
	var data []TestMsg = []TestMsg{
		{client: 1, msgId: "1", sId: "sid1", dest: "/queue/1", body: "body1"},
		{client: 1, msgId: "2", sId: "sid2", dest: "/queue/2", body: "body2"},
		{client: 1, msgId: "3", sId: "sid2", dest: "/queue/2", body: "body3"},
		{client: 2, msgId: "4", sId: "sid3", dest: "/queue/3", body: "body4"},
	}
	// setup
	server := NewServer()

	// client1
	client1 := &ClientConnection{
		WriteChan: make(chan frame.Frame, 4),
	}
	handler1 := NewHandler(client1, server)

	// client2
	client2 := &ClientConnection{
		WriteChan: make(chan frame.Frame, 4),
	}
	handler2 := NewHandler(client2, server)

	// subscribe client 1
	handler1.Handle(*makeSubscriptionFrame("sid1", "/queue/1"))
	handler1.Handle(*makeSubscriptionFrame("sid2", "/queue/2"))

	// send messages
	for _, msg := range data {
		fr := makeSendFrame(msg.dest, msg.body)
		fr.Header.Set("x-msg-id", msg.msgId)
		handler2.Handle(*fr)
	}

	// subscribe client 2
	handler2.Handle(*makeSubscriptionFrame("sid3", "/queue/3"))

	// check dispatch
	received := make(map[string]int)
	checkMsg := func(fr frame.Frame, client int) {
		assert.Equal(t, frame.CmdMessage, fr.Command)
		msgId, ok := fr.Header.Get("x-msg-id")
		assert.Equal(t, true, ok)
		// find original message definition
		var orig TestMsg
		for _, msg := range data {
			if msg.msgId == msgId {
				orig = msg
				break
			}
		}
		assert.Equal(t, msgId, orig.msgId)
		assert.Equal(t, client, orig.client)
		sId, _ := fr.Header.Get(frame.HdrSubscription)
		assert.Equal(t, orig.sId, sId)
		dest, _ := fr.Header.Get(frame.HdrDestination)
		assert.Equal(t, orig.dest, dest)
		received[msgId] = received[msgId] + 1
	}
	for i := 0; i < 4; i++ {
		select {
		case fr := <-client1.WriteChan:
			checkMsg(fr, 1)
		case fr := <-client2.WriteChan:
			checkMsg(fr, 2)
		case <-time.NewTimer(time.Millisecond * 10).C:
			t.Error("timeout!")
		}
	}
	// assert that we got every message exactly once
	for _, msg := range data {
		n, ok := received[msg.msgId]
		assert.Equal(t, true, ok)
		assert.Equal(t, 1, n)
	}

}

// one producer, one consumer, one queue
func BenchmarkHandlerOne(b *testing.B) {
	server := NewServer()

	// client1
	client1 := &ClientConnection{
		WriteChan: make(chan frame.Frame, 4),
	}
	handler1 := NewHandler(client1, server)

	// client2
	client2 := &ClientConnection{
		WriteChan: make(chan frame.Frame, 4),
	}
	handler2 := NewHandler(client2, server)

	// subscribe client 1
	handler1.Handle(*makeSubscriptionFrame("sid1", "/queue/1"))

	// setup listener
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < b.N; i++ {
			<-client1.WriteChan
		}
		wg.Done()
	}()

	b.ResetTimer()
	// send messages
	for i := 0; i < b.N; i++ {
		handler2.Handle(*makeSendFrame("/queue/1", "bench"))
	}
	wg.Wait()

}
