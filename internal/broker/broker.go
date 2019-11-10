package broker

import (
	"github.com/gorilla/mux"
	"github.com/hitman99/sse-go/internal/event"
	"github.com/hitman99/sse-go/internal/subscriber"
	"github.com/hitman99/sse-go/internal/topic"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Broker interface {
	Notify() func(w http.ResponseWriter, r *http.Request)
	RegisterSubscriber() func(w http.ResponseWriter, r *http.Request)
}

type broker struct {
	topics     map[string]*topic.Topic
	topicMutex *sync.Mutex
	shutdown   chan bool
	closer     chan subscriber.Subscriber
	idGen      *event.IdGenerator
	logger     *log.Logger
	timeout    time.Duration
}

func NewBroker(timeout time.Duration) Broker {
	closer := make(chan subscriber.Subscriber)
	b := &broker{
		topics:     map[string]*topic.Topic{},
		topicMutex: &sync.Mutex{},
		idGen:      event.NewIdGenerator(),
		logger:     log.New(os.Stdout, "[broker] ", log.Ltime),
		closer:     closer,
		timeout:    timeout,
	}
	go b.CleanSubscribers()
	return b
}

func (b *broker) Notify() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "empty body", http.StatusBadRequest)
				return
			}
			ev := event.NewEvent(b.idGen.GetId(), event.KIND_MSG, body)
			b.topicMutex.Lock()
			defer b.topicMutex.Unlock()
			t := b.getOrCreateTopic(mux.Vars(r)["topic"])
			t.Notify(ev.Marshal())
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (b *broker) RegisterSubscriber() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tp := mux.Vars(r)["topic"]
		b.logger.Printf("new subsciber with ip %s for topic %s", r.RemoteAddr, tp)
		t := b.getOrCreateTopic(tp)
		sub := subscriber.MakeSubscriber(b.timeout, b.closer)
		t.Subscribe(sub)
		b.logger.Printf("subscribers for topic %s: %d", tp, len(t.Subscribers))
		sub.HttpHandler(w, r)
	}
}

func (b *broker) getOrCreateTopic(name string) *topic.Topic {
	if t, found := b.topics[name]; found {
		return t
	} else {
		t := topic.MakeTopic(name)
		b.topics[name] = t
		return t
	}
}

func (b *broker) CleanSubscribers() {
	for {
		select {
		case sub := <-b.closer:
			b.logger.Printf("removing subscriber %s", sub.GetIp())
			for _, t := range b.topics {
				t.Unsubscribe(sub)
			}
		}
	}
}
