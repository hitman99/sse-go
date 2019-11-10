package broker_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hitman99/sse-go/internal/broker"
	"github.com/hitman99/sse-go/internal/event"
	"github.com/hitman99/sse-go/internal/subscriber/mock"
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

type BrokerTestSuite struct {
	suite.Suite
}

func (s *BrokerTestSuite) TestNewBroker() {
	b := broker.NewBroker(time.Second)
	s.Assert().NotNil(b)
}

func (s *BrokerTestSuite) TestNotifyEmptyBody() {
	b := broker.NewBroker(time.Second)

	r, _ := http.NewRequestWithContext(context.TODO(), "GET", "nevermind", nil)
	r.RemoteAddr = "1.2.3.4:64646"
	w := httptest.NewRecorder()

	handler := b.Notify()
	handler(w, r)
	s.Assert().Equal(400, w.Code)
}

func (s *BrokerTestSuite) TestNotifyWithBody() {
	wg := &sync.WaitGroup{}
	body := "message"
	buf := bytes.NewBuffer([]byte(body))

	b := broker.NewBroker(time.Second * 3)
	sub := mock.Subscriber{}
	sub.On("Notify", body).Return()

	ctx := &DoneContext{
		Context: context.TODO(),
		Closer:  make(chan struct{}),
	}
	rNotify, _ := http.NewRequestWithContext(context.TODO(), "GET", "nevermind", buf)
	wNotify := httptest.NewRecorder()

	rReg, _ := http.NewRequestWithContext(ctx, "GET", "localhost/infocenter/test", nil)
	rReg.RemoteAddr = "1.2.3.4:64646"
	wReg := httptest.NewRecorder()

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		b.RegisterSubscriber()(wReg, rReg)
	}(wg)
	time.Sleep(time.Millisecond * 100)
	b.Notify()(wNotify, rNotify)
	close(ctx.Closer)
	wg.Wait()

	s.Assert().Equal(204, wNotify.Code)
	s.Assert().Equal(fmt.Sprintf("id: 1\nevent: %s\ndata: %s\n\n", event.KIND_MSG.String(), body), wReg.Body.String())
}

func TestBrokerSuite(t *testing.T) {
	suite.Run(t, new(BrokerTestSuite))
}
