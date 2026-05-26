package trade

import (
	"fmt"
	"strings"
)

// TradeDocument 状态常量。
//
// M-12: 之前 TradeDocument.Status 是无约束字符串，唯一已使用的值是 'DRAFT'。
// 把允许的状态正式化并提供转换矩阵，避免新代码引入意外字符串。
const (
	TradeDocumentStatusDraft     = "DRAFT"
	TradeDocumentStatusSent      = "SENT"
	TradeDocumentStatusConfirmed = "CONFIRMED"
	TradeDocumentStatusVoided    = "VOIDED"
)

// validTradeDocumentTransitions 列出每个状态可达的目标状态。
var validTradeDocumentTransitions = map[string]map[string]bool{
	TradeDocumentStatusDraft:     {TradeDocumentStatusSent: true, TradeDocumentStatusVoided: true},
	TradeDocumentStatusSent:      {TradeDocumentStatusConfirmed: true, TradeDocumentStatusVoided: true},
	TradeDocumentStatusConfirmed: {TradeDocumentStatusVoided: true},
	TradeDocumentStatusVoided:    {},
}

// ValidateTradeDocumentStatusTransition 校验 current → target 是否被允许。
// 若 current == target 直接通过，避免空转写入 false negative。
func ValidateTradeDocumentStatusTransition(current, target string) error {
	cur := strings.ToUpper(strings.TrimSpace(current))
	tgt := strings.ToUpper(strings.TrimSpace(target))
	if cur == "" {
		cur = TradeDocumentStatusDraft
	}
	if cur == tgt {
		return nil
	}
	allowed, known := validTradeDocumentTransitions[cur]
	if !known {
		return fmt.Errorf("unknown trade document status %q", current)
	}
	if !allowed[tgt] {
		return fmt.Errorf("cannot transition trade document from %q to %q", current, target)
	}
	return nil
}

// SettlementRecord 状态常量（M-13）。
const (
	SettlementStatusUnpaid  = "UNPAID"
	SettlementStatusPartial = "PARTIAL"
	SettlementStatusPaid    = "PAID"
	SettlementStatusVoided  = "VOIDED"
)

// validSettlementTransitions 列出 SettlementRecord 允许的状态流。
var validSettlementTransitions = map[string]map[string]bool{
	SettlementStatusUnpaid:  {SettlementStatusPartial: true, SettlementStatusPaid: true, SettlementStatusVoided: true},
	SettlementStatusPartial: {SettlementStatusPaid: true, SettlementStatusVoided: true},
	SettlementStatusPaid:    {SettlementStatusVoided: true},
	SettlementStatusVoided:  {},
}

// ValidateSettlementStatusTransition 校验 SettlementRecord 状态转换。
func ValidateSettlementStatusTransition(current, target string) error {
	cur := strings.ToUpper(strings.TrimSpace(current))
	tgt := strings.ToUpper(strings.TrimSpace(target))
	if cur == "" {
		cur = SettlementStatusUnpaid
	}
	if cur == tgt {
		return nil
	}
	allowed, known := validSettlementTransitions[cur]
	if !known {
		return fmt.Errorf("unknown settlement status %q", current)
	}
	if !allowed[tgt] {
		return fmt.Errorf("cannot transition settlement from %q to %q", current, target)
	}
	return nil
}
