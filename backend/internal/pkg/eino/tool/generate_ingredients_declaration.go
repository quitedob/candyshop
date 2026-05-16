package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateIngredientsDeclRequest struct {
	ProductName     string   `json:"product_name" jsonschema_description:"Name of the candy product"`
	Ingredients     []string `json:"ingredients" jsonschema_description:"List of explicit ingredients"`
	Allergens       []string `json:"allergens" jsonschema_description:"List of known allergens (e.g., peanuts, dairy)"`
	IsHalal         bool     `json:"is_halal" jsonschema_description:"Does the product comply with Halal standards"`
	ContainsGelatin bool     `json:"contains_gelatin" jsonschema_description:"Does the product contain animal gelatin"`
	TradeID         uint     `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateIngredientsDeclResponse struct {
	Status  string `json:"status"`
	DocNo   string `json:"doc_no"`
	Content string `json:"content"`
	SavedTo string `json:"saved_to,omitempty"`
}

func NewGenerateIngredientsDeclarationTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_ingredients_declaration",
		"Generate an ingredient and allergen declaration for food safety compliance.",
		func(ctx context.Context, req *GenerateIngredientsDeclRequest) (*GenerateIngredientsDeclResponse, error) {
			docNo := fmt.Sprintf("ING-DEC-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
	"product": %q,
	"ingredients": %v,
	"allergens": %v,
	"halal": %t,
	"contains_gelatin": %t
}`, req.ProductName, req.Ingredients, req.Allergens, req.IsHalal, req.ContainsGelatin)

			resp := &GenerateIngredientsDeclResponse{
				Status:  "ingredients_declaration_drafted",
				DocNo:   docNo,
				Content: contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          "INGREDIENTS_DECLARATION",
					DocNumber:     docNo,
					Status:        tradeModels.TradeStatusDraft,
					IsAIGenerated: true,
					LineageSource: "ai_draft",
				}
				contentJSON, _ := json.Marshal(contentStr)
				if ue := json.Unmarshal(contentJSON, &doc.Content); ue != nil {
				log.Printf("eino tool: unmarshal doc content failed: %v", ue)
			}
				if err := persister.AddDocument(ctx, doc); err != nil {
					return resp, fmt.Errorf("persist failed: %w", err)
				}
				resp.SavedTo = fmt.Sprintf("DB(trade=%d)", req.TradeID)
			}

			return resp, nil
		})
	if err != nil {
		return nil, err
	}

	return baseTool, nil
}
