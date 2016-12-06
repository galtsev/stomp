package main

import (
	"github.com/galtsev/stomp/server"
	"github.com/go-stomp/stomp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const addr string = "localhost:1620"

func makeServer() *server.Server {
	srv := server.NewServer()
	srv.NotifyChan = make(chan struct{}, 1)
	// run server
	go srv.ListenAndServe(addr)
	<-srv.NotifyChan
	return srv
}

func _TestConnectDisconnect(t *testing.T) {
	srv := makeServer()
	defer srv.Stop()

	conn, err := stomp.Dial("tcp", addr)
	assert.NoError(t, err)
	err = conn.Disconnect()
	assert.NoError(t, err)
}

// conn1 subscribe to /queue/1
// conn2 send message to /queue/1
func TestSubscribeAndSend(t *testing.T) {
	srv := makeServer()
	defer srv.Stop()

	conn1, err := stomp.Dial("tcp", addr)
	assert.NoError(t, err)
	defer conn1.Disconnect()
	sub, err := conn1.Subscribe("/queue/1", stomp.AckAuto)
	assert.NoError(t, err)

	conn2, err := stomp.Dial("tcp", addr)
	assert.NoError(t, err)
	defer conn2.Disconnect()

	body := "from conn1"
	conn2.Send("/queue/1", "text/plain", []byte(body))

	select {
	case msg := <-sub.C:
		assert.Equal(t, []byte(body), msg.Body)
	case <-time.NewTimer(time.Millisecond * 10).C:
		t.Error("timeout receiving message")

	}
}
