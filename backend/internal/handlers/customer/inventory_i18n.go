package customer

import (
	modelsProduct "candypro/api/internal/models/product"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/pkg/inventorywarn"

	"github.com/gin-gonic/gin"
)

func formatInventoryWarnings(c *gin.Context, warnings []orderSvc.InventoryWarning, productByID map[string]modelsProduct.Product) []string {
	return inventorywarn.FormatWarnings(c, warnings, productByID)
}
