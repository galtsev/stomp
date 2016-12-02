package server

import (
	"github.com/go-stomp/stomp/frame"
	"strings"
)

type Server struct {
	InChan      chan frame.Frame
	Dispatchers map[string]Dispatcher
}

func NewServer() *Server {
	return &Server{
		InChan:      make(chan frame.Frame),
		Dispatchers: make(map[string]Dispatcher),
	}
}

func (s *Server) GetDispatcher(destination string) Dispatcher {
	dispatcher, ok := s.Dispatchers[destination]
	if !ok {
		if strings.HasPrefix(destination, "/queue") {
			dispatcher = NewQueue(destination)
		} else {
			dispatcher = NewTopic(destination)
		}
		s.Dispatchers[destination] = dispatcher
	}
	return dispatcher
}
