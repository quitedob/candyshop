package customer

import (
	"context"
	"net/http"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
)

type cartCostProductRepository struct {
	*fakeProductRepo
	unitCost float64
}

func (repository *cartCostProductRepository) ComputeWeightedAvgCost(context.Context, string) float64 {
	return repository.unitCost
}

func TestCartCheckout_PersistsCOGSWithConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	products := map[string]modelsProduct.Product{
		"costed-cart-product": {
			ID: "costed-cart-product", Status: "active", BasePrice: 10, MOQ: 1,
			StockQuantity: 100, Ingredients: "Sugar, Pectin", Allergens: "none",
		},
	}
	orderRepository := &fakeOrderRepo{}
	handler := buildCartCheckoutHandler(products,
		[]modelsOrder.CartItem{{ProductID: "costed-cart-product", Quantity: 2, UnitPrice: 10, Currency: "USD"}},
		orderService.NewOrderService(orderRepository), userService.NewUserService(&fakeUserRepo{}), nil, nil)
	handler.services.Product = productService.NewProductService(&cartCostProductRepository{
		fakeProductRepo: &fakeProductRepo{products: products}, unitCost: 4,
	})
	response := performCartCheckout(handler, "cart-cost-buyer", map[string]interface{}{
		"currency": "USD",
		"shippingAddress": map[string]interface{}{
			"street": "1 Main St", "city": "New York", "state": "NY", "zipCode": "10001", "country": "USA",
		},
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, body = %s", response.Code, response.Body.String())
	}
	committedOrder := orderRepository.createdOrder
	if committedOrder == nil || committedOrder.ConfirmedAt == nil || committedOrder.COGS != 8 {
		t.Fatalf("expected confirmed checkout to persist 2 × 4 cost atomically; got %+v", committedOrder)
	}
}
