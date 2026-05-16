package eino

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// PostgresCheckPointStore implements compose.CheckPointStore backed by a PostgreSQL table.
type PostgresCheckPointStore struct {
	db *gorm.DB
}

// checkpointRow is the database row.
type checkpointRow struct {
	CheckPointID string `gorm:"primaryKey;column:checkpoint_id"`
	Data         []byte `gorm:"type:bytea"`
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
	return s.db.WithContext(ctx).Save(&row).Error
}
