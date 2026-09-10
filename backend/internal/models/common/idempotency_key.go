package common

import "time"

// IdempotencyStatusPending reserves a key before the handler performs effects.
// It is never an HTTP response status and must not be replayed to a client.
const IdempotencyStatusPending = 0

// IdempotencyKey records the result of a previous mutating request keyed by a
// client-supplied `Idempotency-Key` header. Replays return the cached response
// without re-executing the operation (M-17).
//
// Scope: per (user_id, method, path, key). Different users, different routes,
// or different methods do NOT collide; users cannot replay across endpoints.
type IdempotencyKey struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Key           string    `gorm:"type:varchar(128);not null;uniqueIndex:idx_idem_user_route_key,priority:4" json:"key"`
	UserID        string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_idem_user_route_key,priority:1;index" json:"userId"`
	Method        string    `gorm:"type:varchar(8);not null;uniqueIndex:idx_idem_user_route_key,priority:2" json:"method"`
	Path          string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_idem_user_route_key,priority:3" json:"path"`
	RequestHash   string    `gorm:"type:varchar(64);not null" json:"-"`
	StatusCode    int       `json:"statusCode"`
	ResponseBody  string    `gorm:"type:text" json:"responseBody"`
	ResponseCType string    `gorm:"type:varchar(128)" json:"responseContentType"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
	ExpiresAt     time.Time `gorm:"index" json:"expiresAt"`
}
