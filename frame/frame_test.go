package frame

import (
	"testing"
)

func makeHeaders() [][]byte {
	data := []string{"id:194-197497987", "destination:/queue/foo", "content-length:4000", "receipt:bazzzz"}
	res := make([][]byte, 4)
	for i, s := range data {
		res[i] = []byte(s)
	}
	return res
}

func BenchmarkParseHeader(b *testing.B) {
	binHeaders := makeHeaders()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fr := New()
		fr.Command = CmdMessage
		for _, h := range binHeaders {
			fr.Header.Parse(h)
		}
	}
}
