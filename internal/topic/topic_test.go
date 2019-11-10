package topic_test

import (
	"fmt"
	"github.com/hitman99/sse-go/internal/subscriber"
	"github.com/hitman99/sse-go/internal/subscriber/mock"
	"github.com/hitman99/sse-go/internal/topic"
	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type TopicTestSuite struct {
	suite.Suite
	TopicName string
}

func (s *TopicTestSuite) SetupTest() {
	s.TopicName = "test-topic"
}

func (s *TopicTestSuite) TestMake() {
	t := topic.MakeTopic(s.TopicName)
	s.Assert().Equal(t.Name, s.TopicName)
	s.Assert().Equal(len(t.Subscribers), 0)
}

func (s *TopicTestSuite) TestSubscribe() {
	name := "test-topic"
	t := topic.MakeTopic(name)
	closer := make(chan subscriber.Subscriber)
	sub := subscriber.MakeSubscriber(time.Second*2, closer)
	t.Subscribe(sub)
	s.Assert().Equal(len(t.Subscribers), 1)
}

func (s *TopicTestSuite) TestNotify() {
	msg := "message"
	sub := new(mock.Subscriber)
	sub.On("Notify", msg).Return()
	t := topic.MakeTopic("test")
	t.Subscribe(sub)
	t.Notify(msg)
	sub.AssertCalled(s.T(), "Notify", msg)
}

func (s *TopicTestSuite) TestUnsubscribe() {
	sub := new(mock.Subscriber)
	t := topic.MakeTopic("test")
	t.Subscribe(sub)
	s.Assert().Equal(len(t.Subscribers), 1)
	t.Unsubscribe(sub)
	s.Assert().Equal(len(t.Subscribers), 0)
}

func (s *TopicTestSuite) TestConcurrentAccess() {
	sub := new(mock.Subscriber)
	t := topic.MakeTopic("test")
	sub.On("Notify", tmock.Anything).Return()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond)
			t.Notify(fmt.Sprintf("message #%d", i))
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond)
			t.Subscribe(sub)
		}
	}(wg)

	wg.Wait()

}

func TestTopicSuite(t *testing.T) {
	suite.Run(t, new(TopicTestSuite))
}
