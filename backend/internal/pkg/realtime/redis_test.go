package realtime

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// TestNewRedisBroadcasterHappyPathWithRedis verifies the bounded handshake does
// not break the normal path: with a healthy Redis the subscription is confirmed
// and the broadcaster is immediately usable for in-process fan-out.
func TestNewRedisBroadcasterHappyPathWithRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	b, err := NewRedisBroadcaster("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRedisBroadcaster: %v", err)
	}
	defer b.Close()

	sub := b.Subscribe("room:1")
	defer sub.Close()
	b.Publish("room:1", []byte("hi"))

	select {
	case got := <-sub.Messages():
		if string(got) != "hi" {
			t.Fatalf("got %q, want %q", got, "hi")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for in-process delivery")
	}
}

// TestNewRedisBroadcasterSubscribeHandshakeIsBounded is the regression test for
// the startup hang. The fake Redis answers PING but accepts the PSUBSCRIBE
// without ever sending the confirmation — exactly the "connection dropped right
// after Ping" case. On the pre-fix path Receive blocks on a background context
// forever; with the fix startup must fail within the bounded handshake timeout.
func TestNewRedisBroadcasterSubscribeHandshakeIsBounded(t *testing.T) {
	addr := startSilentSubscribeServer(t)

	type result struct {
		b   *RedisBroadcaster
		err error
	}
	resCh := make(chan result, 1)
	go func() {
		b, err := NewRedisBroadcaster("redis://" + addr)
		resCh <- result{b, err}
	}()

	select {
	case res := <-resCh:
		if res.err == nil {
			defer res.b.Close()
			t.Fatal("expected an error when Redis never confirms the subscription")
		}
	case <-time.After(8 * time.Second):
		t.Fatal("NewRedisBroadcaster hung on the subscribe handshake; startup is unbounded")
	}
}

// TestNewRedisBroadcasterPingFailsBounded covers the other startup failure
// mode: Redis completely unreachable. The ping timeout must bound the failure.
func TestNewRedisBroadcasterPingFailsBounded(t *testing.T) {
	start := time.Now()
	_, err := NewRedisBroadcaster("redis://127.0.0.1:59999")
	if err == nil {
		t.Fatal("expected an error for an unreachable Redis")
	}
	if elapsed := time.Since(start); elapsed > 6*time.Second {
		t.Fatalf("NewRedisBroadcaster took %s to fail; startup is unbounded", elapsed)
	}
}

// startSilentSubscribeServer listens on an ephemeral port and answers PING with
// PONG, but after receiving PSUBSCRIBE keeps the connection open without
// sending the subscription confirmation.
func startSilentSubscribeServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleSilentConn(conn)
		}
	}()
	return ln.Addr().String()
}

func handleSilentConn(conn net.Conn) {
	defer conn.Close()
	br := bufio.NewReader(conn)
	for {
		cmd, _, err := readRESPCommand(br)
		if err != nil {
			return
		}
		switch cmd {
		case "HELLO":
			// go-redis runs HELLO as part of connection init; reply with a valid
			// RESP2 flat-array map (server/version) so initConn proceeds to the
			// subscription and we can exercise the silent-PSUBSCRIBE path below.
			_, _ = io.WriteString(conn, "*4\r\n$6\r\nserver\r\n$5\r\nredis\r\n$7\r\nversion\r\n$5\r\n7.0.0\r\n")
		case "PING":
			_, _ = io.WriteString(conn, "+PONG\r\n")
		case "PSUBSCRIBE", "SUBSCRIBE":
			// Never send the *3 psubscribe confirmation; this is what makes the
			// pre-fix handshake block forever.
			select {}
		default:
			// Swallow other handshake commands (AUTH, CLIENT SETINFO) with a
			// bare OK so the connection stays usable for the next command.
			_, _ = io.WriteString(conn, "+OK\r\n")
		}
	}
}

// readRESPCommand reads a single RESP array command ("*N\r\n$len\r\n...") and
// returns its upper-cased command name and arguments.
func readRESPCommand(r *bufio.Reader) (string, []string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "*") {
		return "", nil, fmt.Errorf("expected RESP array, got %q", line)
	}
	n, err := strconv.Atoi(line[1:])
	if err != nil {
		return "", nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return "", nil, err
		}
		hdr = strings.TrimSpace(hdr)
		if !strings.HasPrefix(hdr, "$") {
			return "", nil, fmt.Errorf("expected RESP bulk string, got %q", hdr)
		}
		ln, err := strconv.Atoi(hdr[1:])
		if err != nil {
			return "", nil, err
		}
		buf := make([]byte, ln+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", nil, err
		}
		args = append(args, string(buf[:ln]))
	}
	cmd := ""
	if len(args) > 0 {
		cmd = strings.ToUpper(args[0])
	}
	return cmd, args, nil
}
