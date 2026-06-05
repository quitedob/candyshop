// Package realtime provides an in-process publish/subscribe fan-out for
// WebSocket delivery (order conversation messages today, extensible later).
//
// The Broadcaster interface is the seam that keeps the transport swappable:
// the default Hub fans out via Go channels (single-instance deployments),
// and a Redis pub/sub implementation can be dropped in for multi-instance
// deployments without changing any handler or the WebSocket Manager.
package realtime

import "sync"

// Subscription is a single consumer's view of a room's message stream.
type Subscription interface {
	// Messages returns the receive-only channel of payloads for this room.
	// The channel is closed when Close is called.
	Messages() <-chan []byte
	// Close detaches the subscription and releases its resources. It is safe
	// to call Close more than once.
	Close()
}

// Broadcaster fans payloads out to every active Subscription of a room.
// Implementations must be safe for concurrent use.
type Broadcaster interface {
	// Publish delivers payload to every current subscriber of room. It must
	// never block on a slow consumer.
	Publish(room string, payload []byte)
	// Subscribe registers a new consumer for room.
	Subscribe(room string) Subscription
}

// subBufferSize bounds per-subscriber buffering. A consumer that falls this
// far behind is treated as slow and has messages dropped rather than blocking
// the publisher; the client recovers via reconnect + REST refetch.
const subBufferSize = 32

// Hub is an in-process Broadcaster backed by Go channels. It requires no
// central goroutine: publishing iterates the room's subscribers and performs
// non-blocking sends.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*channelSub]struct{}
}

// NewHub creates an empty Hub ready for use.
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*channelSub]struct{})}
}

type channelSub struct {
	hub  *Hub
	room string
	ch   chan []byte
	mu   sync.Mutex
	dead bool
	once sync.Once
}

func (s *channelSub) Messages() <-chan []byte { return s.ch }

func (s *channelSub) Close() {
	s.once.Do(func() {
		s.hub.remove(s)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.dead = true
		close(s.ch)
	})
}

func (s *channelSub) publish(payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		return
	}
	select {
	case s.ch <- payload:
	default:
		// Buffer full: drop. The client reconnects and refetches via REST.
	}
}

// Subscribe registers a new consumer for room and returns its Subscription.
func (h *Hub) Subscribe(room string) Subscription {
	sub := &channelSub{hub: h, room: room, ch: make(chan []byte, subBufferSize)}
	h.mu.Lock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*channelSub]struct{})
	}
	h.rooms[room][sub] = struct{}{}
	h.mu.Unlock()
	return sub
}

func (h *Hub) remove(sub *channelSub) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs, ok := h.rooms[sub.room]
	if !ok {
		return
	}
	delete(subs, sub)
	if len(subs) == 0 {
		delete(h.rooms, sub.room)
	}
}

// Publish delivers payload to every current subscriber of room. Slow consumers
// (full buffers) are skipped so one stalled client cannot block the others.
func (h *Hub) Publish(room string, payload []byte) {
	h.mu.RLock()
	subs := h.rooms[room]
	targets := make([]*channelSub, 0, len(subs))
	for s := range subs {
		targets = append(targets, s)
	}
	h.mu.RUnlock()

	for _, s := range targets {
		s.publish(payload)
	}
}

// RoomCount reports the number of active rooms. Intended for tests/metrics.
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}
