package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// TranslateFunc is the function signature for batch translation.
// sourceData: field name → source text
// Returns: locale → field name → translated text
type TranslateFunc func(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, error)

// TranslateContentRequest is the input for the translate_content tool.
type TranslateContentRequest struct {
	Fields        map[string]string `json:"fields" jsonschema_description:"Map of field names to source text to translate"`
	TargetLocales []string          `json:"target_locales" jsonschema_description:"List of target locale codes, e.g. [en, ar, ja]"`
}

// TranslateContentResponse is the output from the translate_content tool.
type TranslateContentResponse struct {
	Translations map[string]map[string]string `json:"translations"`
	LocalesDone  int                          `json:"locales_done"`
}

// NewTranslateContentTool creates an Eino InvokableTool for content translation.
// translateFn is injected by the caller to avoid circular dependencies between
// the tool package and the service package.
func NewTranslateContentTool(ctx context.Context, translateFn TranslateFunc) (tool.BaseTool, error) {
	if translateFn == nil {
		return nil, fmt.Errorf("translateFn is required")
	}

	baseTool, err := utils.InferTool("translate_content",
		"Translate named text fields into multiple target languages. Use this when you need to localize product names, descriptions, summaries, or other content fields for international markets. Returns translations keyed by locale code.",
		func(ctx context.Context, req *TranslateContentRequest) (*TranslateContentResponse, error) {
			if req == nil || len(req.Fields) == 0 {
				return nil, fmt.Errorf("fields is required and must not be empty")
			}
			if len(req.TargetLocales) == 0 {
				return nil, fmt.Errorf("target_locales is required and must not be empty")
			}

			translations, err := translateFn(ctx, req.Fields, req.TargetLocales)
			if err != nil {
				return nil, fmt.Errorf("translation failed: %w", err)
			}

			return &TranslateContentResponse{
				Translations: translations,
				LocalesDone:  len(translations),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
