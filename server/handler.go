package server

import (
	"errors"
	"github.com/galtsev/stomp/frame"
	"io"
)

func Err(msg string) {
	panic(errors.New(msg))
}

type Handler struct {
	Server        *Server
	inChan        chan frame.Frame
	outChan       chan frame.Frame
	subscriptions map[string]Dispatcher
	waitingAcks   map[string]func()
}

func NewHandler(server *Server, reader io.Reader, writer io.Writer) *Handler {
	handler := Handler{
		Server:        server,
		inChan:        make(chan frame.Frame),
		outChan:       make(chan frame.Frame),
		subscriptions: make(map[string]Dispatcher),
		waitingAcks:   make(map[string]func()),
	}
	if writer != nil {
		go handler.writeLoop(writer)
	}
	if reader != nil {
		go handler.readLoop(reader)
	}
	go handler.processLoop()
	return &handler
}

func (h *Handler) readLoop(r io.Reader) {
	reader := frame.NewReader(r)
	for {
		fr, err := reader.Read()
		if err != nil {
			h.Err(err.Error())
		}
		h.inChan <- *fr
	}
}

func (h *Handler) writeLoop(w io.Writer) {
	writer := frame.NewWriter(w)
	for fr := range h.outChan {
		writer.Write(&fr)
	}
}

func (h *Handler) processLoop() {
	for fr := range h.inChan {
		h.Handle(fr)
	}
}

func (h *Handler) addAckCallBack(msgId string, cb func()) {
	h.waitingAcks[msgId] = cb
}

func (h *Handler) Disconnect() {
	for subscriptionId, dispatcher := range h.subscriptions {
		dispatcher.Unsubscribe(subscriptionId)
	}
}

func (h *Handler) Err(msg string) {
	fr := frame.New()
	fr.Command = frame.CmdError
	fr.Header.Set(frame.HdrMessage, msg)
	h.outChan <- *fr
	h.Disconnect()
}

func (h *Handler) Handle(fr frame.Frame) {
	switch fr.Command {

	case frame.CmdConnect:
		//h.state = StateConnected

	case frame.CmdDisconnect:
		for subscriptionId, dispatcher := range h.subscriptions {
			dispatcher.Unsubscribe(subscriptionId)
		}

	case frame.CmdSubscribe:
		destination, ok := fr.Header.Get(frame.HdrDestination)
		if !ok {
			h.Err("Missing destination header")
		}
		subscriptionId, ok := fr.Header.Get(frame.HdrId)
		if !ok {
			h.Err("Missing subscription id header")
		}
		dispatcher := h.Server.GetDispatcher(destination)
		h.subscriptions[subscriptionId] = dispatcher
		options := SubscriptionOptions{
			ClientWriteChan: h.outChan,
			AddAckCallback:  h.addAckCallBack,
		}
		dispatcher.Subscribe(fr, options)

	case frame.CmdUnsubscribe:
		subscriptionId, ok := fr.Header.Get(frame.HdrId)
		if !ok {
			h.Err("Missing subscription id header")
		}
		if dispatcher, ok := h.subscriptions[subscriptionId]; ok {
			delete(h.subscriptions, subscriptionId)
			dispatcher.Unsubscribe(subscriptionId)
		}

	case frame.CmdSend:
		destination, ok := fr.Header.Get(frame.HdrDestination)
		if !ok {
			h.Err("Missing destination header")
		}
		dispatcher := h.Server.GetDispatcher(destination)
		outFr := fr.Clone()
		outFr.Command = frame.CmdMessage
		dispatcher.Send(*outFr)

	case frame.CmdAck:
		id, ok := fr.Header.Get(frame.HdrId)
		if !ok {
			h.Err("Missing Id header")
		}
		wh, ok := h.waitingAcks[id]
		if ok {
			delete(h.waitingAcks, id)
			wh()
		}
	}

	if receiptId, ok := fr.Header.Get(frame.HdrReceipt); ok {
		recFrame := frame.New()
		recFrame.Command = frame.CmdReceipt
		recFrame.Header.Set(frame.HdrReceipt, receiptId)
		h.outChan <- *recFrame
	}

}
