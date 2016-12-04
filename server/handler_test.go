package server

import (
	"bytes"
	"github.com/galtsev/stomp/frame"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

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

func TestHandlerSubscribeAndSend(t *testing.T) {
	var (
		subscriptionId string = randomString(8)
		destination    string = "/queue/" + randomString(4)
		body           string = randomString(16)
	)
	// setup
	server := NewServer()
	handler := NewHandler(server, nil, nil)

	// subscribe
	handler.Handle(*makeSubscriptionFrame(subscriptionId, destination))

	// send
	handler.Handle(*makeSendFrame(destination, body))

	// check output
	select {
	case outFr := <-handler.outChan:
		assert.Equal(t, frame.CmdMessage, outFr.Command, "Wrong message type")
		assert.Equal(t, []byte(body), outFr.Body, "wrong message body")
		subId, _ := outFr.Header.Get(frame.HdrSubscription)
		assert.Equal(t, subscriptionId, subId, "wrong subscription id")
	case <-time.NewTimer(time.Millisecond).C:
		t.Error("Timeout getting dispatched message")
	}
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
	handler1 := NewHandler(server, nil, nil)

	// client2
	handler2 := NewHandler(server, nil, nil)

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
		case fr := <-handler1.outChan:
			checkMsg(fr, 1)
		case fr := <-handler2.outChan:
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

// this test use frame.Reader and frame.Writer
// one client subscribe to queue /queue/1
// another one send three messages to this queue
func TestHandlerReaderWriter(t *testing.T) {
	server := NewServer()
	makeConn := func() (reader io.Reader, writer io.Writer, h *Handler) {
		reader, hWriter := io.Pipe()
		hReader, writer := io.Pipe()
		h = NewHandler(server, hReader, hWriter)
		return
	}
	_, pWriter, producer := makeConn()
	cReader, cWriter, consumer := makeConn()

	send := func(w io.Writer, msg string) {
		n, err := w.Write([]byte(msg))
		assert.NoError(t, err)
		assert.Equal(t, len(msg), n)
	}
	// subscribe
	send(cWriter, "SUBSCRIBE\ndestination:/queue/1\nid:sub1\n\n\x00")

	// setup listener
	var wg sync.WaitGroup
	frameReader := frame.NewReader(cReader)
	expect := func(body string) {
		fr, err := frameReader.Read()
		assert.NoError(t, err)
		assert.Equal(t, []byte(body), fr.Body)
	}
	wg.Add(1)
	go func() {
		expect("msg1")
		expect("msg2")
		expect("another message")
		wg.Done()
	}()

	// send messages
	send(pWriter, "SEND\ndestination:/queue/1\n\nmsg1\x00")
	send(pWriter, "SEND\ndestination:/queue/1\n\nmsg2\x00")
	send(pWriter, "SEND\ndestination:/queue/1\n\nanother message\x00")
	wg.Wait()
	producer.Disconnect()
	consumer.Disconnect()
}

func BenchmarkHandlerReaderWriter(b *testing.B) {
	server := NewServer()
	makeConn := func() (reader io.Reader, writer io.Writer, h *Handler) {
		reader, hWriter := io.Pipe()
		hReader, writer := io.Pipe()
		h = NewHandler(server, hReader, hWriter)
		return
	}
	_, pWriter, producer := makeConn()
	cReader, cWriter, consumer := makeConn()

	send := func(w io.Writer, msg string) {
		n, err := w.Write([]byte(msg))
		assert.NoError(b, err)
		assert.Equal(b, len(msg), n)
	}
	// subscribe
	send(cWriter, "SUBSCRIBE\ndestination:/queue/1\nid:sub1\n\n\x00")

	// setup listener
	var wg sync.WaitGroup
	frameReader := frame.NewReader(cReader)
	wg.Add(1)
	go func() {
		for {
			fr, err := frameReader.Read()
			if err != nil {
				b.Error(err.Error())
			}
			if bytes.Equal(fr.Body, []byte("done")) {
				break
			}
		}
		wg.Done()
	}()

	// send messages
	for i := 0; i < b.N; i++ {
		send(pWriter, "SEND\ndestination:/queue/1\n\nmessage body\x00")
	}
	send(pWriter, "SEND\ndestination:/queue/1\n\ndone\x00")
	wg.Wait()
	producer.Disconnect()
	consumer.Disconnect()
}

// one producer, one consumer, one queue
func BenchmarkHandlerOne(b *testing.B) {
	server := NewServer()

	// client1
	handler1 := NewHandler(server, nil, nil)

	// client2
	handler2 := NewHandler(server, nil, nil)

	// subscribe client 1
	handler1.Handle(*makeSubscriptionFrame("sid1", "/queue/1"))

	// setup listener
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < b.N; i++ {
			<-handler1.outChan
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
