package product

import "time"

// ProductMarketCostStack 目的国成本栈（物流/关税/标签摊销等），供定价与 AI 建议引用
type ProductMarketCostStack struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProductID         string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_prod_mkt_cost" json:"productId"`
	MarketCode        string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_prod_mkt_cost" json:"marketCode"`
	LogisticsPerUnit  float64   `json:"logisticsPerUnit"`
	DutyRate          float64   `json:"dutyRate"` // 0~1
	LabelCostPerUnit  float64   `json:"labelCostPerUnit"`
	CompliancePerUnit float64   `json:"compliancePerUnit"`
	TargetGrossMargin float64   `json:"targetGrossMargin"` // 0~1，可选目标毛利率
	Currency          string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Notes             string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
