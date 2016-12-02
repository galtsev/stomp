package server

import (
	"encoding/hex"
	"math/rand"
)

func randomBytes(size int) []byte {
	buf := make([]byte, size)
	out := make([]byte, size*2)
	rand.Read(buf)
	hex.Encode(out, buf)
	return out
}

func randomString(size int) string {
	return string(randomBytes(size))
}
