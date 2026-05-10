package trade

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func init() {
	schema.Register[*HitlRequest]()
	schema.Register[*VerificationReport]()
}

// ReviewDecision is the result returned by human operator
type ReviewDecision struct {
	Approved bool   `json:"approved"`
	Notes    string `json:"notes"`
	// Provide a way to supply overridden data if needed
	OverriddenData []*ExtractedData `json:"overridden_data,omitempty"`
}

func init() {
	schema.Register[*ReviewDecision]()
}

// NewHitlNode creates a lambda node that conditionally interrupts for human review
func NewHitlNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, req *HitlRequest) (*HitlRequest, error) {

		wasInterrupted, _, storedReq := compose.GetInterruptState[*HitlRequest](ctx)

		if !wasInterrupted {
			// First run
			if req.Report.ManualReviewReq {
				// Needs review, trigger interrupt
				return nil, compose.StatefulInterrupt(ctx, req, req)
			}
			// Automatic pass
			return req, nil
		}

		// Resuming from interrupt
		isTarget, hasData, decision := compose.GetResumeContext[*ReviewDecision](ctx)
		if isTarget && hasData {
			if decision.Approved {
				// Human operator approved it
				storedReq.Report.Passed = true
				storedReq.Report.ManualReviewReq = false
				if decision.OverriddenData != nil {
					storedReq.Data = decision.OverriddenData
				}
				return storedReq, nil
			}
			// Human rejected it, we could pass an error or return failed state
			storedReq.Report.Passed = false
			return storedReq, nil
		}

		// Re-interrupt if we aren't the target or no data was provided
		return nil, compose.StatefulInterrupt(ctx, storedReq, storedReq)
	})
}
