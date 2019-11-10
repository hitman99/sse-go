package subscriber_test

import (
	"context"
	"fmt"
	"github.com/hitman99/sse-go/internal/subscriber"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type DoneContext struct {
	context.Context
	Closer chan struct{}
}

func (d *DoneContext) Done() <-chan struct{} {
	return d.Closer
}

type SubscriberTestSuite struct {
	suite.Suite
}

func (s *SubscriberTestSuite) TestMake() {
	closer := make(chan subscriber.Subscriber)
	sub := subscriber.MakeSubscriber(time.Second, closer)
	s.Assert().NotNil(sub)
}

func (s *SubscriberTestSuite) TestHttpHandler() {
	wg := &sync.WaitGroup{}
	ctx := &DoneContext{
		Closer: make(chan struct{}),
	}
	close(ctx.Closer)
	r, _ := http.NewRequestWithContext(ctx, "GET", "nevermind", nil)
	r.RemoteAddr = "1.2.3.4:64646"
	w := httptest.NewRecorder()

	closer := make(chan subscriber.Subscriber)
	sub := subscriber.MakeSubscriber(time.Second, closer)
	wg.Add(1)
	go func(c <-chan subscriber.Subscriber, wg *sync.WaitGroup) {
		defer wg.Done()
		<-closer
	}(closer, wg)
	sub.HttpHandler(w, r)
	wg.Wait()
	s.Assert().Equal(r.RemoteAddr, sub.GetIp())
	s.Assert().Equal("text/event-stream", w.Header().Get("Content-Type"))
	s.Assert().Equal("no-cache", w.Header().Get("Cache-Control"))
	s.Assert().Equal("keep-alive", w.Header().Get("Connection"))
	s.Assert().Equal("*", w.Header().Get("Access-Control-Allow-Origin"))
	s.Assert().Equal("chunked", w.Header().Get("Transfer-Encoding"))
}

func (s *SubscriberTestSuite) TestNotify() {
	msg := "the message"
	wg := &sync.WaitGroup{}
	ctx := &DoneContext{
		Closer: make(chan struct{}),
	}
	r, _ := http.NewRequestWithContext(ctx, "GET", "nevermind", nil)
	r.RemoteAddr = "1.2.3.4:64646"
	w := httptest.NewRecorder()

	closer := make(chan subscriber.Subscriber)
	sub := subscriber.MakeSubscriber(time.Second, closer)
	wg.Add(3)
	go func(c <-chan subscriber.Subscriber, wg *sync.WaitGroup) {
		defer wg.Done()
		<-closer
	}(closer, wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		sub.Notify(msg)
		close(ctx.Closer)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		sub.HttpHandler(w, r)
	}(wg)
	wg.Wait()

	s.Assert().Equal(msg, w.Body.String())

}

func (s *SubscriberTestSuite) TestTimeout() {
	wg := &sync.WaitGroup{}
	ctx := &DoneContext{
		Closer: make(chan struct{}),
	}
	r, _ := http.NewRequestWithContext(ctx, "GET", "nevermind", nil)
	r.RemoteAddr = "1.2.3.4:64646"
	w := httptest.NewRecorder()

	closer := make(chan subscriber.Subscriber)
	then := time.Now()
	timeout := time.Second * 2
	sub := subscriber.MakeSubscriber(timeout, closer)

	wg.Add(2)
	go func(c <-chan subscriber.Subscriber, wg *sync.WaitGroup) {
		defer wg.Done()
		<-closer
		timeSpent := time.Now().Sub(then).Round(time.Second)
		s.Assert().Equal(timeSpent, timeout)
	}(closer, wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		sub.HttpHandler(w, r)
	}(wg)
	wg.Wait()

	s.Assert().Equal(fmt.Sprintf("event: timeout\ndata: 2s\n\n"), w.Body.String())
}

func (s *SubscriberTestSuite) TestTimeoutAndNotify() {
	msg := "test"
	wg := &sync.WaitGroup{}
	ctx := &DoneContext{
		Closer: make(chan struct{}),
	}
	r, _ := http.NewRequestWithContext(ctx, "GET", "nevermind", nil)
	r.RemoteAddr = "1.2.3.4:64646"
	w := httptest.NewRecorder()

	closer := make(chan subscriber.Subscriber)
	timeout := time.Millisecond * 10
	sub := subscriber.MakeSubscriber(timeout, closer)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		sub.HttpHandler(w, r)
	}(wg)

	time.Sleep(time.Millisecond * 100)
	sub.Notify(msg)
	select {
	case <-ctx.Closer:
	default:
		close(ctx.Closer)
	}

	wg.Wait()

	s.Assert().Equal("event: timeout\ndata: 10ms\n\n", w.Body.String())
}

func TestSubscriberSuite(t *testing.T) {
	suite.Run(t, new(SubscriberTestSuite))
}
