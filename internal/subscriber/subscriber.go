package subscriber

import (
	"fmt"
	"github.com/hitman99/sse-go/internal/event"
	"net/http"
	"sync"
	"time"
)

type Subscriber interface {
	HttpHandler(w http.ResponseWriter, r *http.Request)
	Notify(msg string)
	GetIp() string
}

type subscriber struct {
	ip          string
	connectedAt *time.Time
	timeout     time.Duration
	notify      chan string
	disconnect  chan struct{}
	wg          *sync.WaitGroup
}

func MakeSubscriber(timeout time.Duration, closer chan<- Subscriber) *subscriber {
	notify := make(chan string, 100)
	disconn := make(chan struct{})
	now := time.Now()
	sub := &subscriber{
		connectedAt: &now,
		timeout:     timeout,
		notify:      notify,
		disconnect:  disconn,
		wg:          &sync.WaitGroup{},
	}
	go sub.dropConnection(closer)
	return sub
}

func (s *subscriber) HttpHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	s.ip = r.RemoteAddr
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		select {
		case msg := <-s.notify:
			_, err := fmt.Fprint(w, msg)
			if err != nil {
				close(s.disconnect)
				break
			}
			flusher.Flush()
		case <-r.Context().Done():
			select {
			case <-s.disconnect:
				return
			default:
			}
			close(s.disconnect)
		case <-s.disconnect:
			return
		}
	}
}

func (s *subscriber) Notify(msg string) {
	select {
	case s.notify <- msg:
	default:
	}
}

func (s *subscriber) dropConnection(closer chan<- Subscriber) {
	for {
		select {
		// in case client disconnected first
		case <-s.disconnect:
			closer <- s
			return
		// in case of long connection
		case <-time.After(s.timeout):
			s.Notify(event.NewEvent(0, event.KIND_TIMEOUT, []byte(s.timeout.String())).Marshal())
			select {
			case <-s.disconnect:
				return
			default:
			}
			close(s.disconnect)
		}
	}
}

func (s *subscriber) GetIp() string {
	return s.ip
}
