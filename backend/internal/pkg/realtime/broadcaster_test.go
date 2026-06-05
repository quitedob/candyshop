package realtime

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHubPublishDeliversToSubscribers(t *testing.T) {
	h := NewHub()
	sub1 := h.Subscribe("order:1")
	sub2 := h.Subscribe("order:1")
	defer sub1.Close()
	defer sub2.Close()

	h.Publish("order:1", []byte("hello"))

	for i, sub := range []Subscription{sub1, sub2} {
		select {
		case got := <-sub.Messages():
			if string(got) != "hello" {
				t.Fatalf("sub%d: got %q, want %q", i, got, "hello")
			}
		case <-time.After(time.Second):
			t.Fatalf("sub%d: timed out waiting for message", i)
		}
	}
}

func TestHubPublishIsolatesRooms(t *testing.T) {
	h := NewHub()
	other := h.Subscribe("order:2")
	defer other.Close()

	h.Publish("order:1", []byte("for-room-1"))

	select {
	case msg := <-other.Messages():
		t.Fatalf("room isolation broken: received %q on order:2", msg)
	case <-time.After(100 * time.Millisecond):
		// expected: no cross-room delivery
	}
}

func TestSubscriptionCloseRemovesRoom(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe("order:1")
	if h.RoomCount() != 1 {
		t.Fatalf("RoomCount = %d, want 1", h.RoomCount())
	}
	sub.Close()
	if h.RoomCount() != 0 {
		t.Fatalf("RoomCount after close = %d, want 0", h.RoomCount())
	}
	// Double close must be safe.
	sub.Close()
}

func TestHubPublishDoesNotBlockOnSlowConsumer(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe("order:1")
	defer sub.Close()

	// Overfill well beyond the buffer; publisher must never block.
	done := make(chan struct{})
	go func() {
		for i := 0; i < subBufferSize*4; i++ {
			h.Publish("order:1", []byte("x"))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a slow consumer")
	}
}

func TestHubConcurrentSubscribePublishClose(t *testing.T) {
	h := NewHub()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := h.Subscribe("order:race")
			h.Publish("order:race", []byte("y"))
			s.Close()
		}()
	}
	wg.Wait()
	if h.RoomCount() != 0 {
		t.Fatalf("RoomCount after all close = %d, want 0", h.RoomCount())
	}
}

func TestOriginAllowList(t *testing.T) {
	check := OriginAllowList([]string{"https://app.example.com", "http://localhost:3000/"})

	cases := []struct {
		origin string
		want   bool
	}{
		{"", true}, // non-browser / same-origin
		{"https://app.example.com", true},
		{"https://app.example.com/", true},
		{"http://localhost:3000", true},
		{"https://evil.example.com", false},
	}
	for _, tc := range cases {
		r := newRequestWithOrigin(tc.origin)
		if got := check(r); got != tc.want {
			t.Errorf("origin %q: got %v, want %v", tc.origin, got, tc.want)
		}
	}
}

func TestOriginAllowListWildcard(t *testing.T) {
	check := OriginAllowList([]string{"*"})
	if !check(newRequestWithOrigin("https://anything.example.com")) {
		t.Fatal("wildcard should allow any origin")
	}
}

func newRequestWithOrigin(origin string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/ws", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}
