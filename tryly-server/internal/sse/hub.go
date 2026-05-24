package sse

import "sync"

// Message is a single Server-Sent Events payload.
type Message struct {
	Event string
	Data  string
	ID    string
}

// Hub routes SSE events to subscribed clients per user.
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[chan Message]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]map[chan Message]struct{}),
	}
}

func (h *Hub) Subscribe(userID int64) chan Message {
	ch := make(chan Message, 32)
	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan Message]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(userID int64, ch chan Message) {
	h.mu.Lock()
	delete(h.clients[userID], ch)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Push(userID int64, event, data string) {
	h.push(userID, Message{Event: event, Data: data})
}

func (h *Hub) PushMessage(userID int64, msg Message) {
	h.push(userID, msg)
}

func (h *Hub) push(userID int64, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[userID] {
		select {
		case ch <- msg:
		default:
		}
	}
}
