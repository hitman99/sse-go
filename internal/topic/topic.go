package topic

import (
	"github.com/hitman99/sse-go/internal/subscriber"
	"sync"
)

func MakeTopic(name string) *Topic {
	topicChan := make(chan []byte)
	newTopic := &Topic{
		Channel:     topicChan,
		Name:        name,
		Subscribers: map[subscriber.Subscriber]bool{},
		m:           sync.Mutex{},
	}

	return newTopic
}

type Topic struct {
	Channel     chan []byte
	Name        string
	Subscribers map[subscriber.Subscriber]bool
	m           sync.Mutex
}

func (t *Topic) Notify(msg string) {
	t.m.Lock()
	defer t.m.Unlock()
	for sub := range t.Subscribers {
		sub.Notify(msg)
	}
}

func (t *Topic) Subscribe(sub subscriber.Subscriber) {
	t.m.Lock()
	defer t.m.Unlock()
	t.Subscribers[sub] = true
}

func (t *Topic) Unsubscribe(sub subscriber.Subscriber) {
	t.m.Lock()
	defer t.m.Unlock()
	delete(t.Subscribers, sub)
}
