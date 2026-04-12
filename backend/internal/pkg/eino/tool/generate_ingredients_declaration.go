package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateIngredientsDeclRequest struct {
	ProductName     string   `json:"product_name" jsonschema_description:"Name of the candy product"`
	Ingredients     []string `json:"ingredients" jsonschema_description:"List of explicit ingredients"`
	Allergens       []string `json:"allergens" jsonschema_description:"List of known allergens (e.g., peanuts, dairy)"`
	IsHalal         bool     `json:"is_halal" jsonschema_description:"Does the product comply with Halal standards"`
	ContainsGelatin bool     `json:"contains_gelatin" jsonschema_description:"Does the product contain animal gelatin"`
}

type GenerateIngredientsDeclResponse struct {
	Status    string `json:"status"`
	DocNo     string `json:"doc_no"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func NewGenerateIngredientsDeclarationTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_ingredients_declaration", "Generate an ingredient and allergen declaration for food safety compliance.",
		func(ctx context.Context, req *GenerateIngredientsDeclRequest) (*GenerateIngredientsDeclResponse, error) {

			docNo := fmt.Sprintf("ING-DEC-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"product": "%s",
				"ingredients": %v,
				"allergens": %v,
				"halal": %t,
				"contains_gelatin": %t
			}`, req.ProductName, req.Ingredients, req.Allergens, req.IsHalal, req.ContainsGelatin)

			return &GenerateIngredientsDeclResponse{
				Status:    "ingredients_declaration_drafted",
				DocNo:     docNo,
				Content:   contentStr,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
