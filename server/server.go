package server

import (
	"net"
	"strings"
	"sync"
)

type Server struct {
	Dispatchers map[string]Dispatcher
	dispLock    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		Dispatchers: make(map[string]Dispatcher),
	}
}

func (s *Server) ListenAndServe(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		NewHandler(s, conn, conn)
	}
}

func (s *Server) GetDispatcher(destination string) Dispatcher {
	var dispatcher Dispatcher
	var ok bool
	s.dispLock.RLock()
	dispatcher, ok = s.Dispatchers[destination]
	s.dispLock.RUnlock()
	if !ok {
		s.dispLock.Lock()
		dispatcher, ok = s.Dispatchers[destination]
		if !ok {
			if strings.HasPrefix(destination, "/queue") {
				dispatcher = NewQueue(destination)
			} else {
				dispatcher = NewTopic(destination)
			}
			s.Dispatchers[destination] = dispatcher
		}
		s.dispLock.Unlock()
	}
	return dispatcher
}
