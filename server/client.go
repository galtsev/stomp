package server

import (
	"dan/stomp/frame"
)

type ClientConnection struct {
	OutChan   chan frame.Frame
	WriteChan chan frame.Frame
}
