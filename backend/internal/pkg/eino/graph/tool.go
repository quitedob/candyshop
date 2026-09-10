package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// DocPipelineTool wraps the document generation Graph as an Eino InvokableTool,
// following the official "Graph as Agent Tool" architecture.
type DocPipelineTool struct {
	runner compose.Runnable[DocGenerationInput, DocGenerationOutput]
}

// NewDocPipelineTool compiles the graph and wraps it as a tool.
func NewDocPipelineTool(ctx context.Context, graph *compose.Graph[DocGenerationInput, DocGenerationOutput]) (*DocPipelineTool, error) {
	runner, err := graph.Compile(ctx,
		compose.WithGraphName("TradeDocPipeline"),
	)
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}
	return &DocPipelineTool{runner: runner}, nil
}

// Info returns tool metadata for LLM function calling.
func (t *DocPipelineTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "generate_trade_documents",
		Desc: "Generate trade documents (PI, CI, Sales Contract, Packing List, COO, Health Certificate, B/L, SLI, Insurance, Ingredients) " +
			"for a given trade transaction. Call this when the user wants to create or generate any trade document. " +
			"Provide trade_id (required), doc_types (list of document types), and optionally buyer_name, seller_name, incoterms, payment_terms, total_amount, currency, port_of_loading, port_of_destination.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"trade_id":       {Type: schema.Integer, Desc: "Transaction ID (required)", Required: true},
			"doc_types":      {Type: schema.Array, Desc: "Document types to generate: PROFORMA_INVOICE, COMMERCIAL_INVOICE, SALES_CONTRACT, PACKING_LIST, ORIGIN_CERTIFICATE, HEALTH_CERTIFICATE, BILL_OF_LADING, INGREDIENTS_DECLARATION, SHIPPER_LETTER_OF_INSTRUCTION, INSURANCE_CERTIFICATE", Required: true},
			"buyer_name":     {Type: schema.String, Desc: "Buyer company name"},
			"seller_name":    {Type: schema.String, Desc: "Seller company name (default: CandyPro OEM)"},
			"incoterms":      {Type: schema.String, Desc: "Incoterms, e.g. FOB Shanghai"},
			"payment_terms":  {Type: schema.String, Desc: "Payment terms"},
			"total_amount":   {Type: schema.Number, Desc: "Total transaction amount"},
			"currency":       {Type: schema.String, Desc: "Currency code (USD, EUR, CNY)"},
			"extra_context":  {Type: schema.String, Desc: "Additional business context"},
			"port_of_loading":     {Type: schema.String, Desc: "Port of loading, e.g. Shanghai"},
			"port_of_destination": {Type: schema.String, Desc: "Port of destination/discharge"},
		}),
	}, nil
}

// InvokableRun executes the document generation graph.
func (t *DocPipelineTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input DocGenerationInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	output, err := t.runner.Invoke(ctx, input)
	if err != nil {
		return "", fmt.Errorf("graph execution failed: %w", err)
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("marshal output: %w", err)
	}

	return string(resultJSON), nil
}
