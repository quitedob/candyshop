package product

import (
	"time"

	"gorm.io/datatypes"
)

// ProductMarketProfile 市场版本合规画像（EU/US/GCC 等），用于规则引擎输入
type ProductMarketProfile struct {
	ID                        uint           `gorm:"primaryKey" json:"id"`
	ProductID                 string         `gorm:"type:varchar(100);not null;index:idx_prod_market,unique" json:"productId"`
	MarketCode                string         `gorm:"type:varchar(20);not null;index:idx_prod_market,unique" json:"marketCode"` // EU, US, GCC, SA
	DestinationCountries      datatypes.JSON `gorm:"type:jsonb" json:"destinationCountries,omitempty"`                         // ["DE","FR"]
	BlockedIngredientPatterns datatypes.JSON `gorm:"type:jsonb" json:"blockedIngredientPatterns,omitempty"`                    // 小写子串列表
	RequiredCertKeywords      datatypes.JSON `gorm:"type:jsonb" json:"requiredCertKeywords,omitempty"`
	LabelTemplateID           string         `gorm:"type:varchar(100)" json:"labelTemplateId,omitempty"`
	Notes                     string         `gorm:"type:text" json:"notes,omitempty"`
	// RuleVersion 画像规则集版本号（人工维护，便于审计）
	RuleVersion string `gorm:"type:varchar(40)" json:"ruleVersion,omitempty"`
	// EffectiveFrom 规则对该 SKU 生效时间（可选）
	EffectiveFrom *time.Time `json:"effectiveFrom,omitempty"`
	// RuleSourceSummary 规则来源摘要（法规编号、内控文档链接等，非法律依据声明由接口返回）
	RuleSourceSummary string    `gorm:"type:text" json:"ruleSourceSummary,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
