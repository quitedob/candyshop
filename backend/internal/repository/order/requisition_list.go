package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RequisitionListRepository handles requisition list data operations.
type RequisitionListRepository struct {
	db *gorm.DB
}

// NewRequisitionListRepository creates a new RequisitionListRepository.
func NewRequisitionListRepository(db *gorm.DB) *RequisitionListRepository {
	return &RequisitionListRepository{db: db}
}

// Create creates a requisition list with its items.
func (r *RequisitionListRepository) Create(ctx context.Context, list *modelsOrder.RequisitionList, items []modelsOrder.RequisitionListItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(list).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].RequisitionListID = list.ID
			items[i].CreatedAt = time.Now()
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByUserID returns all requisition lists for a user.
func (r *RequisitionListRepository) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.RequisitionList, error) {
	var lists []modelsOrder.RequisitionList
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC").Find(&lists).Error; err != nil {
		return nil, err
	}
	return lists, nil
}

// FindByID returns a single requisition list.
func (r *RequisitionListRepository) FindByID(ctx context.Context, id string) (*modelsOrder.RequisitionList, error) {
	var list modelsOrder.RequisitionList
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&list).Error; err != nil {
		return nil, err
	}
	return &list, nil
}

// FindItems returns all items for a requisition list.
func (r *RequisitionListRepository) FindItems(ctx context.Context, listID string) ([]modelsOrder.RequisitionListItem, error) {
	var items []modelsOrder.RequisitionListItem
	if err := r.db.WithContext(ctx).Where("requisition_list_id = ?", listID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update updates a requisition list's name/notes.
func (r *RequisitionListRepository) Update(ctx context.Context, list *modelsOrder.RequisitionList) error {
	list.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(&modelsOrder.RequisitionList{}).Where("id = ?", list.ID).Updates(map[string]interface{}{
		"name":       list.Name,
		"notes":      list.Notes,
		"updated_at": list.UpdatedAt,
	}).Error
}

// UpdateItems replaces all items for a requisition list atomically.
func (r *RequisitionListRepository) UpdateItems(ctx context.Context, listID string, items []modelsOrder.RequisitionListItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("requisition_list_id = ?", listID).Delete(&modelsOrder.RequisitionListItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].RequisitionListID = listID
			items[i].ID = 0
			items[i].CreatedAt = time.Now()
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		return tx.Model(&modelsOrder.RequisitionList{}).Where("id = ?", listID).Update("updated_at", now).Error
	})
}

// Delete deletes a requisition list and its items.
func (r *RequisitionListRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("requisition_list_id = ?", id).Delete(&modelsOrder.RequisitionListItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&modelsOrder.RequisitionList{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// AddItem adds a single item to a requisition list.
func (r *RequisitionListRepository) AddItem(ctx context.Context, item *modelsOrder.RequisitionListItem) error {
	item.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(item).Error
}

// RemoveItem removes a single item from a requisition list.
func (r *RequisitionListRepository) RemoveItem(ctx context.Context, listID string, itemID uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND requisition_list_id = ?", itemID, listID).Delete(&modelsOrder.RequisitionListItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}
