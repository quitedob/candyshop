package trade

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// ShipmentInput is the starting point of the workflow
type ShipmentInput struct {
	Documents []*RawDocument `json:"documents"`
}

// ShipmentOutput is the result of the workflow
type ShipmentOutput struct {
	Report *VerificationReport `json:"report"`
	Data   []*ExtractedData    `json:"data"`
}

// BuildTradeWorkflow orchestrates classification, extraction, validation, and HITL
func BuildTradeWorkflow(ctx context.Context, cm model.ChatModel) (compose.Runnable[*ShipmentInput, *HitlRequest], error) {
	graph := compose.NewGraph[*ShipmentInput, *HitlRequest]()

	// Node 1: Process documents in parallel (Classification -> Extraction)
	processDocsNode := compose.InvokableLambda(func(ctx context.Context, input *ShipmentInput) (*HitlRequest, error) {
		var extracted []*ExtractedData

		for _, rawDoc := range input.Documents {
			// A. Classify
			classed, err := ClassifyDocument(ctx, cm, rawDoc)
			if err != nil {
				return nil, fmt.Errorf("classification error for %s: %v", rawDoc.Metadata.FileName, err)
			}

			if classed.DocumentType == "Unknown" {
				continue // skip unknown docs
			}

			// B. Extract
			extData, err := ExtractDocument(ctx, cm, classed)
			if err != nil {
				return nil, fmt.Errorf("extraction error for %s: %v", rawDoc.Metadata.FileName, err)
			}

			extracted = append(extracted, extData)
		}

		return &HitlRequest{Data: extracted}, nil
	})

	err := graph.AddLambdaNode("ProcessDocuments", processDocsNode)
	if err != nil {
		return nil, err
	}

	// Node 2: Validate the combined structured data
	err = graph.AddLambdaNode("ValidateDocuments", NewValidatorNode(cm))
	if err != nil {
		return nil, err
	}

	// Node 3: Check for HITL interrupts
	err = graph.AddLambdaNode("HumanInTheLoop", NewHitlNode())
	if err != nil {
		return nil, err
	}

	// Stitch them together
	err = graph.AddEdge(compose.START, "ProcessDocuments")
	if err != nil {
		return nil, err
	}

	err = graph.AddEdge("ProcessDocuments", "ValidateDocuments")
	if err != nil {
		return nil, err
	}

	err = graph.AddEdge("ValidateDocuments", "HumanInTheLoop")
	if err != nil {
		return nil, err
	}

	err = graph.AddEdge("HumanInTheLoop", compose.END)
	if err != nil {
		return nil, err
	}

	// Compile the graph
	runner, err := graph.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return runner, nil
}
