package inventorywarn

import (
	"fmt"
	"strconv"
	"strings"

	modelsProduct "candypro/api/internal/models/product"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/pkg/i18n"

	"github.com/gin-gonic/gin"
)

func productDisplayName(p modelsProduct.Product) string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	if slug := strings.TrimSpace(p.Slug); slug != "" {
		return slug
	}
	return p.ID
}

// FormatWarnings 将结构化库存警告翻译为面向用户的字符串
func FormatWarnings(c *gin.Context, warnings []orderSvc.InventoryWarning, productByID map[string]modelsProduct.Product) []string {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]string, 0, len(warnings))
	for _, w := range warnings {
		p, ok := productByID[w.ProductID]
		name := w.ProductID
		if ok {
			name = productDisplayName(p)
		}
		vars := map[string]string{
			"productName": name,
			"remaining":   strconv.Itoa(w.Remaining),
			"available":   strconv.Itoa(w.Available),
			"ordered":     strconv.Itoa(w.Ordered),
			"moq":         strconv.Itoa(w.MOQ),
		}
		key := "errors." + w.Code
		msg := i18n.TWithVars(c, key, vars)
		if msg == key || strings.HasPrefix(msg, "errors.") {
			switch w.Code {
			case orderSvc.InventoryWarningSoldOutAdmin:
				msg = fmt.Sprintf("Product %s will be sold out after fulfillment (available: %d, ordered: %d)", name, w.Available, w.Ordered)
			case orderSvc.InventoryWarningBelowMOQAfterOrder:
				msg = fmt.Sprintf("Product %s remaining stock (%d) will be below MOQ (%d) after this order", name, w.Remaining, w.MOQ)
			case orderSvc.InventoryWarningSoldOut:
				msg = fmt.Sprintf("Product %s will be sold out after this order.", name)
			case orderSvc.InventoryWarningLowStock:
				msg = fmt.Sprintf("Product %s has low remaining stock %d after this order.", name, w.Remaining)
			default:
				msg = name
			}
		}
		out = append(out, msg)
	}
	return out
}
