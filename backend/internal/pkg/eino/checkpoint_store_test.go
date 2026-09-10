package eino

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestOwnedCheckPoints_SingleConsumerAndContinuation(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	if err := database.AutoMigrate(&checkpointRow{}); err != nil {
		t.Fatal(err)
	}
	store := &PostgresCheckPointStore{db: database}
	checkpointID := NewOwnedCheckPointID("admin-owner")
	if !CheckPointOwnedBy(checkpointID, "admin-owner") || CheckPointOwnedBy(checkpointID, "admin-other") || CheckPointOwnedBy("trade:5", "admin-owner") {
		t.Fatal("checkpoint ownership boundary failed")
	}
	if err := store.Set(context.Background(), checkpointID, []byte("pending review")); err != nil {
		t.Fatal(err)
	}
	const concurrentResumes = 8
	var workers sync.WaitGroup
	var successfulClaims atomic.Int32
	for attempt := 0; attempt < concurrentResumes; attempt++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			claimed, err := store.ClaimResume(context.Background(), checkpointID)
			if err != nil {
				t.Error(err)
				return
			}
			if claimed {
				successfulClaims.Add(1)
			}
		}()
	}
	workers.Wait()
	if successfulClaims.Load() != 1 {
		t.Fatalf("successful claims=%d", successfulClaims.Load())
	}
	continuationID := NewOwnedCheckPointID("admin-owner")
	continuationStore := store.WithContinuationID(continuationID)
	if err := continuationStore.Set(context.Background(), checkpointID, []byte("next review")); err != nil {
		t.Fatal(err)
	}
	checkpoint, found, err := continuationStore.Get(context.Background(), checkpointID)
	if err != nil || !found || string(checkpoint) != "pending review" {
		t.Fatalf("consumed checkpoint changed: %q found=%v err=%v", checkpoint, found, err)
	}
	if claimed, err := store.ClaimResume(context.Background(), checkpointID); err != nil || claimed {
		t.Fatalf("old checkpoint replayed: claimed=%v err=%v", claimed, err)
	}
	if claimed, err := store.ClaimResume(context.Background(), continuationID); err != nil || !claimed {
		t.Fatalf("new interrupt unavailable: claimed=%v err=%v", claimed, err)
	}
}
