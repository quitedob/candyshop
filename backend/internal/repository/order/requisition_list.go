package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/money"
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

// SumOpenOrderTotalsByCompany returns the sum of total_amount for all open
// (non-terminal) orders of every user belonging to a buyer company, excluding a
// single order ID (the order being confirmed, whose confirmed total the caller
// adds separately). Orders carry only user_id, so the company is resolved
// through users.company_id.
//
// G20 r3: the previous cumulative credit check summed open orders PER USER while
// the credit limit is per COMPANY. Two buyer users of one company could each
// confirm up to the full limit, leaving outstanding exposure exceeding the
// limit with no check catching it. This single aggregate query replaces the
// per-user pagination walk for the credit path — it also avoids the
// created_at-tie skip/double-count hazard of offset pagination. Statuses
// cancelled / returned / expired are terminal and carry no outstanding exposure
// (mirrors openOrderTerminalStatuses in the order service). Soft-deleted rows
// are excluded by the GORM scope on Model(&Order{}).
func (r *OrderRepository) SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error) {
	companyUsers := r.db.WithContext(ctx).
		Model(&modelsUser.User{}).
		Select("id").
		Where("company_id = ?", companyID)
	var result struct {
		Amount float64
	}
	err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Where("id <> ?", excludeOrderID).
		Where("status NOT IN ?", []string{
			modelsOrder.OrderStatusCancelled,
			modelsOrder.OrderStatusReturned,
			modelsOrder.OrderStatusExpired,
		}).
		Where("user_id IN (?)", companyUsers).
		Select("COALESCE(SUM(total_amount), 0) AS amount").
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return money.RoundMoney(result.Amount), nil
}
