package database

import (
	modelsProduct "candypro/api/internal/models/product"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

var legacyLeadTimePattern = regexp.MustCompile(`(?i)(\d+\s*[-–—~]\s*\d+|\d+)\s*(days?|weeks?|months?|天|周|月|工作日|business\s+days?)`)

var legacyGenericLeadCopy = []string{
	"varies by order volume",
	"depends on order complexity and production schedule",
	"varies by scope",
	"confirmed at order time",
	"视订单量而定",
	"视订单复杂度与排产情况而定",
	"视具体情况而定",
}

// NormalizeLegacyLeadTimeCopy 清除数据库中遗留的具体交期天数及通用占位文案，改由管理员在订单/产品中维护。
func NormalizeLegacyLeadTimeCopy(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	productResult := db.Exec(`
		UPDATE products
		SET lead_time = '', updated_at = NOW()
		WHERE lead_time ~* '(\d+\s*[-–—~]?\s*\d*|\d+)\s*(days?|weeks?|months?|天|周|月|business\s+days?)'
		   OR lead_time ~* '^\d'
		   OR lower(lead_time) LIKE '%varies by%'
		   OR lower(lead_time) LIKE '%depends on%'
		   OR lead_time LIKE '%视订单%'
	`)
	if productResult.Error != nil {
		return productResult.Error
	}

	sampleResult := db.Exec(`
		UPDATE products
		SET sample_lead_time = '', updated_at = NOW()
		WHERE sample_lead_time ~* '(\d+\s*[-–—~]?\s*\d*|\d+)\s*(days?|weeks?|months?|天|周|月|business\s+days?)'
		   OR sample_lead_time ~* '^\d'
		   OR lower(sample_lead_time) LIKE '%varies by%'
		   OR lower(sample_lead_time) LIKE '%depends on%'
		   OR sample_lead_time LIKE '%视订单%'
	`)
	if sampleResult.Error != nil {
		return sampleResult.Error
	}

	flowResult := db.Exec(`
		UPDATE oem_flows
		SET timeline = '', updated_at = NOW()
		WHERE timeline ~* '(\d+\s*[-–—~]?\s*\d*|\d+)\s*(days?|weeks?|months?|天|周|月|business\s+days?)'
		   OR timeline ~* '^\d'
		   OR lower(timeline) LIKE '%varies by%'
		   OR lower(timeline) LIKE '%depends on%'
	`)
	if flowResult.Error != nil {
		return flowResult.Error
	}

	caseResult := db.Exec(`
		UPDATE case_studies
		SET timeline = '', updated_at = NOW()
		WHERE timeline ~* '(\d+\s*[-–—~]?\s*\d*|\d+)\s*(days?|weeks?|months?|天|周|月|business\s+days?)'
		   OR timeline ~* '^\d'
		   OR lower(timeline) LIKE '%varies by%'
		   OR lower(timeline) LIKE '%depends on%'
	`)
	if caseResult.Error != nil {
		return caseResult.Error
	}

	translationUpdates, err := normalizeProductTranslationLeadTimes(db)
	if err != nil {
		return err
	}

	stepUpdates, err := normalizeOEMFlowSteps(db)
	if err != nil {
		return err
	}

	total := productResult.RowsAffected + sampleResult.RowsAffected +
		flowResult.RowsAffected + caseResult.RowsAffected + translationUpdates + stepUpdates
	if total > 0 {
		log.Printf("Normalized legacy lead-time copy (%d row updates)", total)
	}
	return nil
}

func normalizeOEMFlowSteps(db *gorm.DB) (int64, error) {
	var flows []modelsProduct.OEMFlow
	if err := db.Find(&flows).Error; err != nil {
		return 0, err
	}

	var updated int64
	for _, flow := range flows {
		changed := false
		steps := flow.Steps
		for i := range steps {
			if shouldClearLeadCopy(steps[i].Duration) {
				steps[i].Duration = ""
				changed = true
			}
		}
		if shouldClearLeadCopy(flow.Timeline) {
			flow.Timeline = ""
			changed = true
		}
		if !changed {
			continue
		}
		flow.Steps = steps
		flow.UpdatedAt = time.Now()
		if err := db.Save(&flow).Error; err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func normalizeProductTranslationLeadTimes(db *gorm.DB) (int64, error) {
	var products []modelsProduct.Product
	if err := db.Select("id", "translations").Find(&products).Error; err != nil {
		return 0, err
	}

	var updated int64
	for _, product := range products {
		if len(product.Translations) == 0 {
			continue
		}
		changed := false
		for locale, fields := range product.Translations {
			value, ok := fields["leadTime"]
			if !ok || !shouldClearLeadCopy(value) {
				continue
			}
			fields["leadTime"] = ""
			product.Translations[locale] = fields
			changed = true
		}
		if !changed {
			continue
		}
		if err := db.Model(&product).Update("translations", product.Translations).Error; err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func shouldClearLeadCopy(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if legacyLeadTimePattern.MatchString(value) {
		return true
	}
	lower := strings.ToLower(value)
	for _, phrase := range legacyGenericLeadCopy {
		if strings.Contains(lower, phrase) || strings.Contains(value, phrase) {
			return true
		}
	}
	for _, r := range value {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
