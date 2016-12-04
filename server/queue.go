package server

import (
	"github.com/galtsev/stomp/frame"
	"sync"
)

// Implement server.Dispatcher
type queueSubscription struct {
	stop chan struct{}
}

type Queue struct {
	Destination   string
	ch            chan frame.Frame
	Subscriptions map[string]*queueSubscription
}

func NewQueue(destination string) *Queue {
	return &Queue{
		Destination:   destination,
		ch:            make(chan frame.Frame, 8),
		Subscriptions: make(map[string]*queueSubscription),
	}
}

func (q *Queue) Send(fr frame.Frame) {
	q.ch <- fr
}

func (q *Queue) Subscribe(fr frame.Frame, options SubscriptionOptions) {
	subscriptionId, _ := fr.Header.Get(frame.HdrId)
	ack, ok := fr.Header.Get(frame.HdrAck)
	if !ok {
		ack = frame.AckAuto
	}
	sub := queueSubscription{
		stop: make(chan struct{}, 0),
	}
	q.Subscriptions[subscriptionId] = &sub
	go func() {
		var wg sync.WaitGroup
		for {
			select {
			case <-sub.stop:
				return
			case fr := <-q.ch:
				var msgId string
				if ack != frame.AckAuto {
					msgId := genId()
					fr.Header.Set(frame.HdrAck, msgId)
				}
				fr.Header.Set(frame.HdrSubscription, subscriptionId)
				options.ClientWriteChan <- fr
				if ack == "client" {
					wg.Add(1)
					options.AddAckCallback(msgId, wg.Done)
					wg.Wait()
				}
			}
		}
	}()
}

func (q *Queue) Unsubscribe(subscriptionId string) {
	if sub, ok := q.Subscriptions[subscriptionId]; ok {
		close(sub.stop)
		delete(q.Subscriptions, subscriptionId)
	}
}
