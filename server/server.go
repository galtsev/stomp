package server

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Dispatchers map[string]Dispatcher
	Handlers    map[string]*Handler
	listener    net.Listener
	hLock       sync.Mutex
	dispLock    sync.RWMutex
	// temporary hook for testing
	// send message to this channel when listener is ready
	NotifyChan chan struct{}
}

func NewServer() *Server {
	return &Server{
		Dispatchers: make(map[string]Dispatcher),
		Handlers:    make(map[string]*Handler),
	}
}

func (s *Server) AddHandler(h *Handler) {
	s.hLock.Lock()
	s.Handlers[h.id] = h
	s.hLock.Unlock()
}

func (s *Server) RemoveHandler(h *Handler) {
	s.hLock.Lock()
	if _, ok := s.Handlers[h.id]; ok {
		delete(s.Handlers, h.id)
	}
	s.hLock.Unlock()
}

func (s *Server) ListenAndServe(addr string) {
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Panic(err)
	}
	s.listener = listener
	log.Println("Listen on", addr)
	// for testing
	if s.NotifyChan != nil {
		s.NotifyChan <- struct{}{}
	}
	for {
		conn, err := s.listener.Accept()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err.Error())
		}
		s.AddHandler(NewHandler(s, conn, conn))
	}
}

func (s *Server) Connect() *net.Conn {
	srvConn, clientConn := net.Pipe()
	s.AddHandler(NewHandler(s, srvConn, srvConn))
	return &clientConn
}

func (s *Server) Stop() {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			log.Println("Error closing ")
		}
	}
	for _, handler := range s.Handlers {
		handler.Disconnect()
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
