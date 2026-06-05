package realtime

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a single frame to the peer.
	writeWait = 10 * time.Second
	// pongWait is how long we wait for a pong before considering the peer dead.
	pongWait = 60 * time.Second
	// pingPeriod must be less than pongWait; ping at this interval to keep the
	// connection alive through proxies and detect half-open sockets.
	pingPeriod = (pongWait * 9) / 10
	// maxClientMessageBytes caps inbound frames. Clients are not expected to
	// send application data (sending happens over REST), so this is small.
	maxClientMessageBytes = 1024
)

// Event is the JSON envelope pushed to WebSocket clients. Type discriminates
// the payload so the channel can carry more than order messages later.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Manager owns the WebSocket upgrade policy and the lifecycle of a single
// connection's read/write pumps. It is transport-agnostic: it talks to a
// Broadcaster, never to the storage or HTTP layers directly.
type Manager struct {
	broadcaster Broadcaster
	upgrader    websocket.Upgrader
}

type closer interface {
	Close() error
}

// OriginChecker decides whether a handshake Origin is allowed.
type OriginChecker func(r *http.Request) bool

// NewManager builds a Manager. allowOrigin gates cross-origin upgrades; pass a
// checker derived from the configured CORS/WS allow-list.
func NewManager(b Broadcaster, allowOrigin OriginChecker) *Manager {
	return &Manager{
		broadcaster: b,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     allowOrigin,
		},
	}
}

// Broadcaster exposes the underlying fan-out so handlers can publish after
// persisting a message.
func (m *Manager) Broadcaster() Broadcaster { return m.broadcaster }

// Close releases optional broadcaster resources such as Redis subscriptions.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	if c, ok := m.broadcaster.(closer); ok {
		return c.Close()
	}
	return nil
}

// Serve upgrades the HTTP request to a WebSocket and bridges the given room's
// subscription to the socket until the client disconnects or ctx-equivalent
// lifecycle ends. It must be called from a Gin handler that has already
// authenticated the user and authorized access to room.
//
// done, when non-nil, is closed once the connection is fully torn down — used
// by tests; production callers may pass nil.
func (m *Manager) Serve(w http.ResponseWriter, r *http.Request, room string) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote an error response on failure.
		return
	}
	defer conn.Close()

	sub := m.broadcaster.Subscribe(room)
	defer sub.Close()

	// Reader pump: we don't expect application messages, but we must drain the
	// socket to process control frames (pong/close) and enforce limits.
	done := make(chan struct{})
	go m.readPump(conn, done)

	m.writePump(conn, sub, done)
}

func (m *Manager) readPump(conn *websocket.Conn, done chan<- struct{}) {
	defer close(done)
	conn.SetReadLimit(maxClientMessageBytes)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (m *Manager) writePump(conn *websocket.Conn, sub Subscription, done <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case payload, ok := <-sub.Messages():
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
