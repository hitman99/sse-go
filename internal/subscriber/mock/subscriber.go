package mock

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type Subscriber struct {
	mock.Mock
}

func (m *Subscriber) HttpHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *Subscriber) Notify(msg string) {
	m.Called(msg)
}

func (m *Subscriber) GetIp() string {
	m.Called()
	return "theIP"
}
