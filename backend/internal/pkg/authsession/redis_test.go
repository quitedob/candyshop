package authsession

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsAuth "candypro/api/internal/models/auth"

	"github.com/alicebob/miniredis/v2"
)

func TestEnsureAvailable(t *testing.T) {
	if err := EnsureAvailable(""); err != nil {
		t.Fatal(err)
	}
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	if err := EnsureAvailable("redis://" + mr.Addr()); err != nil {
		t.Fatal(err)
	}
	if err := EnsureAvailable("redis://127.0.0.1:59999"); err == nil {
		t.Fatal("expected error for unreachable redis")
	}
}

func TestRedisStore_RefreshAndAccess(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	store, err := NewRedisStore("redis://" + mr.Addr())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	exp := time.Now().Add(7 * 24 * time.Hour)
	if err := store.SaveRefreshToken(ctx, &modelsAuth.RefreshToken{
		UserID: "u1", Token: "rtok", ExpiresAt: exp, IPAddress: "127.0.0.1",
	}); err != nil {
		t.Fatal(err)
	}
	tok, err := store.GetRefreshToken(ctx, "rtok")
	if err != nil || tok.UserID != "u1" {
		t.Fatalf("get refresh: %v %+v", err, tok)
	}
	if err := store.SaveAccessSession(ctx, "jti1", AccessSession{UserID: "u1", Email: "a@b.com", Role: "customer"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	sess, err := store.ValidateAccessSession(ctx, "jti1")
	if err != nil || sess.UserID != "u1" {
		t.Fatalf("validate access: %v %+v", err, sess)
	}
	if err := store.RevokeAccessSession(ctx, "jti1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ValidateAccessSession(ctx, "jti1"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound after revoke, got %v", err)
	}
}
