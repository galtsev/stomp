package server

import (
	"github.com/galtsev/stomp/frame"
)

type ClientConnection struct {
	OutChan   chan frame.Frame
	WriteChan chan frame.Frame
}
