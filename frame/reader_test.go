package frame

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var messages []string = []string{
	"ACK\nid:someid\n\n\x00",
	"SEND\ndestination:/queue/foo124\n\nmessage body\x00",
	"SUBSCRIBE\ndestination:/topic/hello/some.new\nid:9789709\nack:client\n\n\x00",
}

func TestReadAckOk(t *testing.T) {
	msg := "ACK\nid:someid\n\n\x00"
	reader := NewReader(bytes.NewReader([]byte(msg)))
	fr, err := reader.Read()
	assert.NoError(t, err, "Unexpected parsing error")
	assert.Equal(t, fr.Command, CmdAck)
	id, _ := fr.Header.Get(HdrId)
	assert.Equal(t, "someid", id, "Id don't match")
}

func TestReadSendOk(t *testing.T) {
	msg := "SEND\ndestination:/queue/foo124\n\nmessage body\x00"
	reader := NewReader(bytes.NewReader([]byte(msg)))
	fr, err := reader.Read()
	assert.NoError(t, err, "Unexpected parsing error")
	assert.Equal(t, fr.Command, CmdSend, "Command don't match")
	destination, _ := fr.Header.Get(HdrDestination)
	assert.Equal(t, "/queue/foo124", destination, "destination don't match")
	assert.Equal(t, []byte("message body"), fr.Body, "body don't match")
}

func TestReadSendErr(t *testing.T) {
	// destination header miss semicolon
	msg := "SEND\ndestination/queue/foo124\n\nmessage body\x00"
	reader := NewReader(bytes.NewReader([]byte(msg)))
	_, err := reader.Read()
	assert.Error(t, err, "Expected error")
}

func TestReadManyFrames(t *testing.T) {
	messages := []string{
		"SUBSCRIBE\nid:19876\ndestination:/queue/9283\n\n\x00",
		"SEND\ndestination:/queue/5602\n\nmsg body\x00",
	}
	reader := NewReader(bytes.NewReader([]byte(strings.Join(messages, ""))))
	fr1, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, CmdSubscribe, fr1.Command)
	fr2, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, []byte("msg body"), fr2.Body)
}
