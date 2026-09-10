package eino

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	timeutil "candypro/api/internal/pkg/timeutil"

	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresCheckPointStore implements compose.CheckPointStore backed by a PostgreSQL table.
type PostgresCheckPointStore struct {
	db *gorm.DB
}

// checkpointRow is the database row.
type checkpointRow struct {
	CheckPointID  string    `gorm:"primaryKey;column:checkpoint_id"`
	Data          []byte    `gorm:"type:bytea"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
	ResumeClaimed bool      `gorm:"not null;default:false"`
}

func (checkpointRow) TableName() string { return "_eino_checkpoints" }

var (
	_autoMigrateOnce sync.Once
	_autoMigrateErr  error
)

// NewPostgresCheckPointStore creates a PostgreSQL-backed checkpoint store.
// It auto-migrates the checkpoint table on first use.
func NewPostgresCheckPointStore(db *gorm.DB) *PostgresCheckPointStore {
	store := &PostgresCheckPointStore{db: db}
	_autoMigrateOnce.Do(func() {
		_autoMigrateErr = db.AutoMigrate(&checkpointRow{})
	})
	return store
}

// AutoMigrateErr returns any error from the auto-migration.
func AutoMigrateErr() error {
	return _autoMigrateErr
}

// NewOwnedCheckPointID creates a fresh run ID in the authenticated user's
// namespace. Client-provided conversation/trade IDs must never become store keys.
func NewOwnedCheckPointID(userID string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(userID)) + ":" + uuid.NewString()
}

// CheckPointOwnedBy rejects legacy unowned keys and checkpoints of other users.
func CheckPointOwnedBy(checkPointID, userID string) bool {
	owner, runID, found := strings.Cut(checkPointID, ":")
	if !found || userID == "" || owner != base64.RawURLEncoding.EncodeToString([]byte(userID)) {
		return false
	}
	_, err := uuid.Parse(runID)
	return err == nil
}

// ClaimResume consumes a checkpoint once across all server instances. The claim
// stays consumed on completion, cancellation, or failure because effects may
// already have committed. A subsequent interrupt is saved under a fresh key.
func (store *PostgresCheckPointStore) ClaimResume(ctx context.Context, checkPointID string) (bool, error) {
	result := store.db.WithContext(ctx).Model(&checkpointRow{}).
		Where("checkpoint_id = ? AND resume_claimed = ?", checkPointID, false).
		Update("resume_claimed", true)
	return result.RowsAffected != 0, result.Error
}

// WithContinuationID redirects a resumed runner's next checkpoint to a fresh ID,
// leaving the consumed checkpoint unavailable to duplicate resume requests.
func (store *PostgresCheckPointStore) WithContinuationID(checkPointID string) compose.CheckPointStore {
	return &continuationCheckPointStore{PostgresCheckPointStore: store, checkPointID: checkPointID}
}

type continuationCheckPointStore struct {
	*PostgresCheckPointStore
	checkPointID string
}

func (store *continuationCheckPointStore) Set(ctx context.Context, _ string, checkPoint []byte) error {
	return store.PostgresCheckPointStore.Set(ctx, store.checkPointID, checkPoint)
}

// Get retrieves a checkpoint by ID.
func (s *PostgresCheckPointStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	var row checkpointRow
	err := s.db.WithContext(ctx).Where("checkpoint_id = ?", checkPointID).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return row.Data, true, nil
}

// Set stores a checkpoint. It upserts by checkPointID.
func (s *PostgresCheckPointStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	row := checkpointRow{
		CheckPointID: checkPointID,
		Data:         checkPoint,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "checkpoint_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"data", "updated_at"}),
	}).Create(&row).Error
}

// CleanupOlderThan deletes checkpoints not updated within the given duration.
// Returns the number of deleted rows.
func (s *PostgresCheckPointStore) CleanupOlderThan(ctx context.Context, maxAge time.Duration) (int64, error) {
	cutoff := timeutil.Now().Add(-maxAge)
	result := s.db.WithContext(ctx).
		Where("updated_at < ?", cutoff).
		Delete(&checkpointRow{})
	return result.RowsAffected, result.Error
}
