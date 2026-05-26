package shipmenttrack

import (
	"strings"
	"time"
)

// ResolveTrackingStatus 根据最新物流事件描述与 ETA 推断追踪状态。
func ResolveTrackingStatus(lastEvent string, eta *time.Time) string {
	event := strings.ToLower(strings.TrimSpace(lastEvent))
	if strings.Contains(event, "delivered") || strings.Contains(event, "arrived at destination") {
		return "DELIVERED"
	}
	if strings.Contains(event, "customs hold") || strings.Contains(event, "delay") {
		return "DELAYED"
	}
	if eta != nil && time.Now().UTC().After(eta.UTC()) {
		return "PAST_ETA"
	}
	if event != "" {
		return "IN_TRANSIT"
	}
	return "TRACKING_PENDING"
}

// MapToDBStatus 将追踪推断结果映射为 ShipmentTracking 数据库存储状态；空串表示无需变更。
func MapToDBStatus(resolved, current string) string {
	switch resolved {
	case "DELIVERED":
		if current != "DELIVERED" {
			return "DELIVERED"
		}
	case "DELAYED":
		if current != "DELIVERED" && current != "EXCEPTION" {
			return "EXCEPTION"
		}
	case "IN_TRANSIT":
		if current == "PENDING" || current == "DISPATCHED" {
			return "IN_TRANSIT"
		}
	}
	return ""
}
