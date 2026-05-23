package notification

import "sync"

type sseMessage struct {
	Event string
	Data  string
	ID    string
}

type sseHub struct {
	mu      sync.RWMutex
	clients map[int64]map[chan sseMessage]struct{}
}

var globalHub = &sseHub{
	clients: make(map[int64]map[chan sseMessage]struct{}),
}

func (h *sseHub) subscribe(userID int64) chan sseMessage {
	ch := make(chan sseMessage, 32)
	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan sseMessage]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *sseHub) unsubscribe(userID int64, ch chan sseMessage) {
	h.mu.Lock()
	delete(h.clients[userID], ch)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
	h.mu.Unlock()
	close(ch)
}

// PushEvent sends an SSE event to a specific user (for cross-package use).
func PushEvent(userID int64, event, data string) {
	globalHub.push(userID, sseMessage{Event: event, Data: data})
}

func (h *sseHub) push(userID int64, msg sseMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[userID] {
		select {
		case ch <- msg:
		default: // drop if channel full; client will catch up on next poll
		}
	}
}
