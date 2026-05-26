package orderwarehouse

import (
	"context"
)

// ProductWarehouseReader 读取默认仓库 ID。
type ProductWarehouseReader interface {
	GetDefaultWarehouseID(ctx context.Context) (string, error)
}

// ChannelWarehouseResolver 解析渠道默认仓（多仓模式可选）。
type ChannelWarehouseResolver interface {
	ResolveWebstoreDefaultWarehouse(ctx context.Context) (string, error)
}

// AssignDefault 为订单设置默认仓库 ID（单仓/多仓 + 可选渠道默认仓）。
func AssignDefault(
	ctx context.Context,
	enableMultiWarehouse bool,
	product ProductWarehouseReader,
	channel ChannelWarehouseResolver,
	orderWarehouseID **string,
) {
	if product == nil || orderWarehouseID == nil {
		return
	}
	if !enableMultiWarehouse {
		wid, err := product.GetDefaultWarehouseID(ctx)
		if err != nil || wid == "" {
			return
		}
		*orderWarehouseID = &wid
		return
	}
	if channel != nil {
		if wh, err := channel.ResolveWebstoreDefaultWarehouse(ctx); err == nil && wh != "" {
			*orderWarehouseID = &wh
			return
		}
	}
	wid, err := product.GetDefaultWarehouseID(ctx)
	if err != nil || wid == "" {
		return
	}
	*orderWarehouseID = &wid
}
