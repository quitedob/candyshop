package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const redisChannelPrefix = "candypro:realtime:"

type redisEnvelope struct {
	InstanceID string `json:"instanceId"`
	Payload    []byte `json:"payload"`
}

// RedisBroadcaster fans out locally and mirrors events over Redis pub/sub so
// subscribers connected to another API instance receive the same payload.
type RedisBroadcaster struct {
	client     *redis.Client
	pubsub     *redis.PubSub
	hub        *Hub
	instanceID string
	cancel     context.CancelFunc
	closeOnce  sync.Once
}

// NewRedisBroadcaster connects to Redis and starts the cross-instance relay.
func NewRedisBroadcaster(redisURL string) (*RedisBroadcaster, error) {
	opts, err := redis.ParseURL(strings.TrimSpace(redisURL))
	if err != nil {
		return nil, fmt.Errorf("realtime: parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("realtime: redis ping: %w", err)
	}

	relayCtx, relayCancel := context.WithCancel(context.Background())
	pubsub := client.PSubscribe(relayCtx, redisChannelPrefix+"*")
	if _, err := pubsub.Receive(relayCtx); err != nil {
		relayCancel()
		_ = pubsub.Close()
		_ = client.Close()
		return nil, fmt.Errorf("realtime: redis subscribe: %w", err)
	}
	b := &RedisBroadcaster{
		client:     client,
		pubsub:     pubsub,
		hub:        NewHub(),
		instanceID: uuid.NewString(),
		cancel:     relayCancel,
	}
	go b.relay(relayCtx)
	return b, nil
}

// NewBroadcaster uses Redis when configured and the in-process Hub otherwise.
func NewBroadcaster(redisURL string) (Broadcaster, error) {
	if strings.TrimSpace(redisURL) == "" {
		return NewHub(), nil
	}
	return NewRedisBroadcaster(redisURL)
}

func (b *RedisBroadcaster) Subscribe(room string) Subscription {
	return b.hub.Subscribe(room)
}

func (b *RedisBroadcaster) Publish(room string, payload []byte) {
	b.hub.Publish(room, payload)
	raw, err := json.Marshal(redisEnvelope{InstanceID: b.instanceID, Payload: payload})
	if err != nil {
		return
	}
	_ = b.client.Publish(context.Background(), redisChannelPrefix+room, raw).Err()
}

func (b *RedisBroadcaster) relay(ctx context.Context) {
	ch := b.pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var env redisEnvelope
			if json.Unmarshal([]byte(msg.Payload), &env) != nil || env.InstanceID == b.instanceID {
				continue
			}
			room := strings.TrimPrefix(msg.Channel, redisChannelPrefix)
			if room != "" {
				b.hub.Publish(room, env.Payload)
			}
		}
	}
}

// Close stops the relay and releases Redis resources.
func (b *RedisBroadcaster) Close() error {
	var closeErr error
	b.closeOnce.Do(func() {
		b.cancel()
		if err := b.pubsub.Close(); err != nil {
			closeErr = err
		}
		if err := b.client.Close(); closeErr == nil {
			closeErr = err
		}
	})
	return closeErr
}
