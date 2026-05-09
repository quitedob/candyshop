package order

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// FindInquiryTradeHints 读取询盘中的 Incoterms 与商务备注片段（供 Outbox Relay 无 Inquiry 服务时使用）
func (r *OrderRepository) FindInquiryTradeHints(ctx context.Context, inquiryID string) (incoterms, commercialNotes string, err error) {
	type row struct {
		Incoterms              string
		NegotiatedPaymentTerms string
		PackagingRequirements  string
		FlavorRequirements     string
	}
	var rw row
	err = r.db.WithContext(ctx).Table("inquiries").Where("id = ?", inquiryID).First(&rw).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", nil
		}
		return "", "", err
	}
	inc := strings.TrimSpace(rw.Incoterms)
	var parts []string
	if s := strings.TrimSpace(rw.PackagingRequirements); s != "" {
		parts = append(parts, "Packaging: "+s)
	}
	if s := strings.TrimSpace(rw.NegotiatedPaymentTerms); s != "" {
		parts = append(parts, "Payment terms: "+s)
	}
	if s := strings.TrimSpace(rw.FlavorRequirements); s != "" {
		parts = append(parts, "Product specs: "+s)
	}
	return inc, strings.Join(parts, "; "), nil
}
