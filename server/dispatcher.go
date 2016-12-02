package server

import (
	"github.com/galtsev/stomp/frame"
)

type SubscriptionOptions struct {
	ClientWriteChan chan frame.Frame
	AddAckCallback  func(msgId string, cb func())
}

type Dispatcher interface {
	Send(fr frame.Frame)
	Subscribe(fr frame.Frame, options SubscriptionOptions)
	Unsubscribe(subscriptionId string)
}
