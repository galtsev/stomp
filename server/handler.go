package server

import (
	"errors"
	"github.com/galtsev/stomp/frame"
)

func Err(msg string) {
	panic(errors.New(msg))
}

type Handler struct {
	Client        *ClientConnection
	Server        *Server
	subscriptions map[string]Dispatcher
	waitingAcks   map[string]func()
}

func NewHandler(client *ClientConnection, server *Server) *Handler {
	return &Handler{
		Client:        client,
		Server:        server,
		subscriptions: make(map[string]Dispatcher),
		waitingAcks:   make(map[string]func()),
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
	h.Client.WriteChan <- fr
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
			ClientWriteChan: h.Client.WriteChan,
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
		h.Client.WriteChan <- *recFrame
	}

}
