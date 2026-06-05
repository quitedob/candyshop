package realtime

import (
	"encoding/json"
	"time"
)

// OrderRoom returns the room key for an order's conversation thread.
func OrderRoom(orderID string) string { return "order:" + orderID }

// EventOrderMessage is the event type for a newly created order message.
const EventOrderMessage = "order_message"

// EventOrderMessagesRead is emitted when one side reads the opposite party's messages.
const EventOrderMessagesRead = "order_messages_read"

// OrderMessagesReadPayload describes a read receipt for an order room.
type OrderMessagesReadPayload struct {
	ReaderType string    `json:"readerType"`
	ReadAt     time.Time `json:"readAt"`
	Count      int64     `json:"count"`
}

// PublishOrderMessage marshals msg into the standard Event envelope and fans it
// out to the order's room. A marshalling failure is silently ignored — the
// client still recovers via REST refetch on its next poll/reconnect.
func (m *Manager) PublishOrderMessage(orderID string, msg interface{}) {
	if m == nil {
		return
	}
	payload, err := json.Marshal(Event{Type: EventOrderMessage, Payload: msg})
	if err != nil {
		return
	}
	m.broadcaster.Publish(OrderRoom(orderID), payload)
}

// PublishOrderMessagesRead fans a read receipt out to the order room.
func (m *Manager) PublishOrderMessagesRead(orderID, readerType string, readAt time.Time, count int64) {
	if m == nil || count == 0 {
		return
	}
	payload, err := json.Marshal(Event{
		Type: EventOrderMessagesRead,
		Payload: OrderMessagesReadPayload{
			ReaderType: readerType,
			ReadAt:     readAt,
			Count:      count,
		},
	})
	if err != nil {
		return
	}
	m.broadcaster.Publish(OrderRoom(orderID), payload)
}
