package internal

import (
	"errors"
	"fmt"
	"github.com/peter-svintsitskyi/go-gesclient/client"
	"github.com/peter-svintsitskyi/go-gesclient/log"
	"reflect"
	"sync"
)

type eventHandlers struct {
	mu sync.Mutex
	handlers []client.EventHandler
}

func newEventHandlers() *eventHandlers {
	return &eventHandlers{
		handlers: make([]client.EventHandler, 0, 10),
	}
}

func (h *eventHandlers) Add(handler client.EventHandler) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, handler)
	return nil
}

func (h *eventHandlers) Remove(handler client.EventHandler) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	pos := -1
	for i, h := range h.handlers {
		if fmt.Sprintf("%v", h) == fmt.Sprintf("%v", handler) {
			pos = i
			break
		}
	}
	if pos == -1 {
		return errors.New("handler not found")
	}
	h.handlers = append(h.handlers[:pos], h.handlers[pos+1:]...)
	return nil
}

func (h *eventHandlers) Raise(evt client.Event) {
	go func() {
		for _, h := range h.handlers {
			if err := h(evt); err != nil {
				log.Errorf("Error occurred while raising event %s: %v", reflect.TypeOf(evt), err)
			}
		}
	}()
}
