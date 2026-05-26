package order

import (
	common "candypro/api/internal/models/common"
	"time"
)

// CartItem represents a single product line in a user's shopping cart.
type CartItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_cart_user_product" json:"userId"`
	ProductID      string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_cart_user_product" json:"productId"`
	ProductName    string    `gorm:"type:varchar(255)" json:"productName"`
	Quantity       int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice      float64   `json:"unitPrice"`
	Currency       string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Specifications string    `gorm:"type:text" json:"specifications"` // JSON or free-text OEM specs
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CartItemResponse 返回给前端的购物车行（含产品展示字段）
type CartItemResponse struct {
	ID             uint           `json:"id"`
	UserID         string         `json:"userId"`
	ProductID      string         `json:"productId"`
	ProductName    string         `json:"productName"`
	Name           string         `json:"name"`
	Quantity       int            `json:"quantity"`
	UnitPrice      float64        `json:"unitPrice"`
	Currency       string         `json:"currency"`
	Specifications string         `json:"specifications"`
	Thumbnail      string         `json:"thumbnail"`
	HalalCertified bool           `json:"halalCertified"`
	OEMAvailable   bool           `json:"oemAvailable"`
	MOQ            int            `json:"moq"`
	MaxQuantity    int            `json:"maxQuantity"`
	Category       string         `json:"category"`
	Translations   common.JSONMap `json:"translations"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// CartItemResponseFromItem 由购物车行构造 API 响应（默认 MOQ=1）
func CartItemResponseFromItem(item CartItem) CartItemResponse {
	name := item.ProductName
	return CartItemResponse{
		ID:             item.ID,
		UserID:         item.UserID,
		ProductID:      item.ProductID,
		ProductName:    item.ProductName,
		Name:           name,
		Quantity:       item.Quantity,
		UnitPrice:      item.UnitPrice,
		Currency:       item.Currency,
		Specifications: item.Specifications,
		MOQ:            1,
		MaxQuantity:    ResolveCartMaxQuantity(item.Quantity, 0),
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

// ResolveCartMaxQuantity 计算数量输入上限，确保当前数量不会触发 HTML5 invalid
func ResolveCartMaxQuantity(currentQty, sellable int) int {
	maxQty := sellable
	if maxQty <= 0 {
		maxQty = 999999
	}
	if currentQty > maxQty {
		return currentQty
	}
	return maxQty
}
