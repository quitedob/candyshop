package order

import "time"

// RequisitionList is a saved shopping template (reorder list) for B2B buyers.
type RequisitionList struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"userId" gorm:"index;not null"`
	Name      string    `json:"name" gorm:"not null"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RequisitionListItem is a single line in a requisition list.
type RequisitionListItem struct {
	ID                uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RequisitionListID string    `json:"requisitionListId" gorm:"index;not null"`
	ProductID         string    `json:"productId" gorm:"not null"`
	Quantity          int       `json:"quantity" gorm:"not null"`
	CreatedAt         time.Time `json:"createdAt"`
}

// TableName overrides the default table name for RequisitionListItem.
func (RequisitionListItem) TableName() string {
	return "requisition_list_items"
}
