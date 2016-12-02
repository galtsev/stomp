package server

import (
	"github.com/galtsev/stomp/frame"
)

type topicSubscription struct {
	clientWriteChan chan frame.Frame
}

type Topic struct {
	Subscribers map[string]*topicSubscription
}

func NewTopic(destination string) *Topic {
	return &Topic{
		Subscribers: make(map[string]*topicSubscription),
	}
}

func (t *Topic) Send(fr frame.Frame) {
	for _, sub := range t.Subscribers {
		sub.clientWriteChan <- fr
	}
}

func (t *Topic) Subscribe(fr frame.Frame, options SubscriptionOptions) {
	subscriptionId, _ := fr.Header.Get(frame.HdrId)
	sub := topicSubscription{
		clientWriteChan: options.ClientWriteChan,
	}
	t.Subscribers[subscriptionId] = &sub
}

func (t *Topic) Unsubscribe(subscriptionId string) {
	if _, ok := t.Subscribers[subscriptionId]; ok {
		delete(t.Subscribers, subscriptionId)
	}
}
