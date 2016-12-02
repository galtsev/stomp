package server

import (
	"encoding/hex"
	"math/rand"
)

func genId() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
