package order

import (
	"context"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
)

type channelRepository interface {
	GetByCode(ctx context.Context, code string) (*modelsProduct.Channel, error)
}

// ChannelService applies channel-level business rules for orders.
type ChannelService struct {
	repo channelRepository
}

func NewChannelService(repo channelRepository) *ChannelService {
	return &ChannelService{repo: repo}
}

// ResolveWebstoreChannel returns the webstore channel config or sensible defaults.
func (s *ChannelService) ResolveWebstoreChannel(ctx context.Context) (*modelsProduct.Channel, error) {
	if s.repo == nil {
		return &modelsProduct.Channel{Code: modelsProduct.ChannelWebstore, PriceMultiplier: 1.0}, nil
	}
	ch, err := s.repo.GetByCode(ctx, modelsProduct.ChannelWebstore)
	if err != nil {
		return &modelsProduct.Channel{Code: modelsProduct.ChannelWebstore, PriceMultiplier: 1.0}, nil
	}
	return ch, nil
}

// ApplyPriceMultiplier adjusts unit price by channel multiplier.
func (s *ChannelService) ApplyPriceMultiplier(price float64, ch *modelsProduct.Channel) float64 {
	if ch == nil || ch.PriceMultiplier <= 0 {
		return price
	}
	return price * ch.PriceMultiplier
}

// DefaultWarehouseID returns channel default warehouse or empty string.
func (s *ChannelService) DefaultWarehouseID(ch *modelsProduct.Channel) string {
	if ch != nil && ch.DefaultWarehouse != nil {
		return *ch.DefaultWarehouse
	}
	return ""
}

// DraftExpireMinutes returns channel draft expiry or fallback default.
func (s *ChannelService) DraftExpireMinutes(ch *modelsProduct.Channel, fallback int) int {
	if ch != nil && ch.DraftExpireMinutes > 0 {
		return ch.DraftExpireMinutes
	}
	if fallback > 0 {
		return fallback
	}
	return 60
}

// ComputeTaxAmount applies channel tax config when client did not supply tax.
func (s *ChannelService) ComputeTaxAmount(subtotal float64, ch *modelsProduct.Channel, clientTax float64) float64 {
	if clientTax > 0 || ch == nil || ch.TaxConfig.TaxRate <= 0 {
		return clientTax
	}
	if ch.TaxConfig.TaxIncluded {
		return clientTax
	}
	return subtotal * ch.TaxConfig.TaxRate
}

// ResolveInitialOrderStatus picks pending_confirmation vs confirmed based on channel + approval.
func (s *ChannelService) ResolveInitialOrderStatus(ch *modelsProduct.Channel, needsApproval bool) string {
	if needsApproval {
		return modelsOrder.OrderStatusPendingApproval
	}
	if ch != nil && ch.AutoConfirm {
		return modelsOrder.OrderStatusConfirmed
	}
	return modelsOrder.OrderStatusPendingConfirm
}

// ResolvePaymentStatus 根据渠道配置解析订单初始支付状态
func (s *ChannelService) ResolvePaymentStatus(ch *modelsProduct.Channel) string {
	if ch == nil {
		return modelsOrder.PaymentStatusUnpaid
	}
	// 不允许未付款下单：必须先支付/授权（仍为 unpaid，但业务上 requires prepayment）
	if !ch.AllowUnpaid {
		return modelsOrder.PaymentStatusUnpaid
	}
	// B2B 账期渠道：允许未付进入确认/履约，标记 partial 表示账期进行中（非全款已付）
	if strings.EqualFold(ch.Type, "b2b") || ch.AutoConfirm {
		return modelsOrder.PaymentStatusPartial
	}
	return modelsOrder.PaymentStatusUnpaid
}
