package event

import (
	"fmt"
	"sync"
)

type EventKind int

const (
	KIND_MSG EventKind = iota
	KIND_TIMEOUT
)

func (e EventKind) String() string {
	return [...]string{"msg", "timeout", "unknown"}[e]
}

type IdGenerator struct {
	m sync.Mutex
	n uint64
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{
		m: sync.Mutex{},
	}
}

func (i *IdGenerator) GetId() uint64 {
	i.m.Lock()
	defer i.m.Unlock()
	i.n++
	return i.n
}

type Event struct {
	Data []byte
	Kind EventKind
	Id   uint64
}

func (e *Event) Marshal() string {
	// omit ID if zero
	if e.Id != 0 {
		return fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", e.Id, e.Kind.String(), e.Data)
	} else {
		return fmt.Sprintf("event: %s\ndata: %s\n\n", e.Kind.String(), e.Data)
	}

}

func NewEvent(id uint64, kind EventKind, data []byte) *Event {
	return &Event{
		Data: data,
		Kind: kind,
		Id:   id,
	}
}
