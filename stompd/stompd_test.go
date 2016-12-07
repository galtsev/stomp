package main

import (
	"bytes"
	"github.com/galtsev/stomp/server"
	"github.com/go-stomp/stomp"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

const addr string = "localhost:1620"

func connect(srv *server.Server) (*stomp.Conn, error) {
	srvConn, clientConn := net.Pipe()
	srv.AddHandler(server.NewHandler(srv, srvConn, srvConn))
	conn, err := stomp.Connect(clientConn)
	return conn, err
}

func TestConnectDisconnect(t *testing.T) {
	srv := server.NewServer()
	defer srv.Stop()

	conn, err := connect(srv)
	assert.NoError(t, err)
	err = conn.Disconnect()
	assert.NoError(t, err)
}

// conn1 subscribe to /queue/1
// conn2 send message to /queue/1
func TestSubscribeAndSend(t *testing.T) {
	srv := server.NewServer()
	defer srv.Stop()

	conn1, err := connect(srv)
	assert.NoError(t, err)
	defer conn1.Disconnect()
	sub, err := conn1.Subscribe("/queue/1", stomp.AckAuto)
	assert.NoError(t, err)

	conn2, err := connect(srv)
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

func BenchmarkSubscribeSend(b *testing.B) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	srv := server.NewServer()
	defer srv.Stop()

	var wg sync.WaitGroup

	conn1, err := connect(srv)
	assert.NoError(b, err)
	defer conn1.Disconnect()
	sub, err := conn1.Subscribe("/queue/1", stomp.AckAuto)
	assert.NoError(b, err)

	wg.Add(1)
	go func() {
		for msg := range sub.C {
			if bytes.Equal(msg.Body, []byte("stop")) {
				break
			}
		}
		conn1.Disconnect()
		wg.Done()
	}()

	conn2, err := connect(srv)
	assert.NoError(b, err)
	defer conn2.Disconnect()

	body := []byte("from conn1")
	for i := 0; i < b.N; i++ {
		conn2.Send("/queue/1", "text/plain", body)
	}
	conn2.Send("/queue/1", "text/plain", []byte("stop"))
	wg.Wait()

}

func BenchmarkSubscribeSendExternal1(b *testing.B) {
	var wg sync.WaitGroup
	conn1, err := stomp.Dial("tcp", addr)
	assert.NoError(b, err)
	if err != nil {
		panic(err)
	}
	defer conn1.Disconnect()
	sub, err := conn1.Subscribe("/queue/1", stomp.AckAuto)
	assert.NoError(b, err)

	wg.Add(1)
	go func() {
		cnt := 0
		for msg := range sub.C {
			if bytes.Equal(msg.Body, []byte("stop")) {
				break
			}
			cnt += 1
		}
		assert.Equal(b, b.N, cnt)
		wg.Done()
	}()

	conn2, err := stomp.Dial("tcp", addr)
	assert.NoError(b, err)
	defer conn2.Disconnect()

	body := []byte("from conn1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn2.Send("/queue/1", "text/plain", body)
	}
	conn2.Send("/queue/1", "text/plain", []byte("stop"))
	wg.Wait()
}
