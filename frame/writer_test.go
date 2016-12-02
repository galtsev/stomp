package frame

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestData struct {
	cmd      string
	headers  []string
	body     string
	expected string
}

func makeFrame(data *TestData) *Frame {
	fr := New()
	fr.Command = data.cmd
	for i := 0; i < len(data.headers); i += 2 {
		fr.Header.Set(data.headers[i], data.headers[i+1])
	}
	fr.Body = []byte(data.body)
	return fr
}

func TestWriter(t *testing.T) {
	data := []TestData{
		{
			cmd:      CmdAck,
			headers:  []string{HdrId, "one4123"},
			body:     "msg bopy987:^256",
			expected: "ACK\nid:one4123\n\nmsg bopy987:^256\x00",
		},
		{
			cmd:      CmdSend,
			headers:  []string{"x-code", "abc:def\n\\oop"},
			body:     "body",
			expected: "SEND\nx-code:abc\\cdef\\n\\\\oop\n\nbody\x00",
		},
	}
	for _, d := range data {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		fr := makeFrame(&d)
		err := writer.Write(fr)
		assert.NoError(t, err)
		assert.Equal(t, d.expected, string(buf.Bytes()))
	}
}

type NullWriter struct{}

func (n *NullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkWriter(b *testing.B) {
	data := TestData{
		cmd:     CmdMessage,
		headers: []string{"destination", "/queue/foo98769", "content-type", "application/json", "content-length", "20480"},
		body:    `{"one":1,"two":2, "fine_key": [1,2,3,4,5,6,7,8,9]}`,
	}
	fr := makeFrame(&data)
	writer := NewWriter(&NullWriter{})
	for i := 0; i < b.N; i++ {
		err := writer.Write(fr)
		if err != nil {
			b.Error(err)
		}

	}
}
