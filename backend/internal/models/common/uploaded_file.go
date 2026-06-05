package common

import "time"

const (
	UploadedFileStatusPending  = "pending_upload"
	UploadedFileStatusUploaded = "uploaded"
	UploadedFileStatusDeleted  = "deleted"

	UploadScanStatusNotConfigured = "not_configured"
	UploadScanStatusPending       = "pending"
	UploadScanStatusClean         = "clean"
	UploadScanStatusInfected      = "infected"
	UploadScanStatusError         = "error"
)

// UploadedFile records object metadata independently of the storage driver.
type UploadedFile struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	StorageKey   string     `json:"storageKey" gorm:"uniqueIndex;not null"`
	URL          string     `json:"url" gorm:"type:text"`
	OriginalName string     `json:"originalName"`
	ContentType  string     `json:"contentType"`
	SizeBytes    int64      `json:"sizeBytes"`
	Folder       string     `json:"folder" gorm:"index"`
	OwnerID      string     `json:"ownerId,omitempty" gorm:"index"`
	Visibility   string     `json:"visibility" gorm:"default:'private'"`
	Status       string     `json:"status" gorm:"index"`
	ScanStatus   string     `json:"scanStatus" gorm:"index"`
	ScanMessage  string     `json:"scanMessage,omitempty" gorm:"type:text"`
	UploadedAt   *time.Time `json:"uploadedAt,omitempty"`
	ScannedAt    *time.Time `json:"scannedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
