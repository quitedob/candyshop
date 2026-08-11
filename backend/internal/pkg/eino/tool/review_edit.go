/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ReviewEditInfo is presented to the user for editing.
type ReviewEditInfo struct {
	ToolName        string
	ArgumentsInJSON string
	ReviewResult    *ReviewEditResult
}

// ReviewEditResult is the result of the user's review.
type ReviewEditResult struct {
	EditedArgumentsInJSON *string
	NoNeedToEdit          bool
	Disapproved           bool
	DisapproveReason      *string
}

func (re *ReviewEditInfo) String() string {
	return fmt.Sprintf("Tool '%s' is about to be called with the following arguments:\n`\n%s\n`\n\n"+
		"Please review and either provide edited arguments in JSON format, "+
		"reply with 'no need to edit', or reply with 'N' to disapprove the tool call.",
		re.ToolName, re.ArgumentsInJSON)
}

func init() {
	schema.Register[*ReviewEditInfo]()
}

// reviewExemptTools are read-only, informational tools that add no value to a
// human review-and-edit gate. Without an exemption the HITL wrapper would
// interrupt them on every call and, because the SSE flow has no resume path,
// they could never complete — leaving e.g. the 17track shipment provider behind
// track_shipment unreachable. The default review policy passes these through so
// they execute immediately.
var reviewExemptTools = map[string]struct{}{
	"track_shipment":        {},
	"translate_content":     {},
	"validate_lc_documents": {},
}

// ReviewPolicyFunc decides whether a tool call must pass through the human
// review-and-edit gate before execution. It returns true when the call requires
// review and false when it should execute immediately. A nil policy (the zero
// value) uses the built-in default, which exempts read-only informational tools.
type ReviewPolicyFunc func(toolName, argumentsInJSON string) bool

func defaultReviewPolicy(toolName, _ string) bool {
	_, exempt := reviewExemptTools[toolName]
	return !exempt
}

// InvokableReviewEditTool is a wrapper that enforces a review-and-edit step.
type InvokableReviewEditTool struct {
	tool.InvokableTool

	// RequiresReview, when non-nil, overrides the default review policy. It lets
	// a caller force a normally-exempt tool through the gate or exempt a tool
	// that the default policy would review.
	RequiresReview ReviewPolicyFunc
}

func (i InvokableReviewEditTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return i.InvokableTool.Info(ctx)
}

func (i InvokableReviewEditTool) reviewRequired(toolName, argumentsInJSON string) bool {
	if i.RequiresReview != nil {
		return i.RequiresReview(toolName, argumentsInJSON)
	}
	return defaultReviewPolicy(toolName, argumentsInJSON)
}

func (i InvokableReviewEditTool) InvokableRun(ctx context.Context, argumentsInJSON string,
	opts ...tool.Option) (string, error) {

	toolInfo, err := i.Info(ctx)
	if err != nil {
		return "", err
	}

	wasInterrupted, _, storedArguments := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		if !i.reviewRequired(toolInfo.Name, argumentsInJSON) {
			// Read-only informational tool: execute directly instead of
			// interrupting, so the call completes and any backing provider
			// stays reachable.
			return i.InvokableTool.InvokableRun(ctx, argumentsInJSON, opts...)
		}
		return "", tool.StatefulInterrupt(ctx, &ReviewEditInfo{
			ToolName:        toolInfo.Name,
			ArgumentsInJSON: argumentsInJSON,
		}, argumentsInJSON)
	}

	isResumeTarget, hasData, data := tool.GetResumeContext[*ReviewEditInfo](ctx)
	if !isResumeTarget {
		return "", tool.StatefulInterrupt(ctx, &ReviewEditInfo{
			ToolName:        toolInfo.Name,
			ArgumentsInJSON: storedArguments,
		}, storedArguments)
	}
	if !hasData || data.ReviewResult == nil {
		return "", fmt.Errorf("tool '%s' resumed with no review data", toolInfo.Name)
	}

	result := data.ReviewResult

	if result.Disapproved {
		if result.DisapproveReason != nil {
			return fmt.Sprintf("tool '%s' disapproved, reason: %s", toolInfo.Name, *result.DisapproveReason), nil
		}
		return fmt.Sprintf("tool '%s' disapproved", toolInfo.Name), nil
	}

	if result.NoNeedToEdit {
		return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
	}

	if result.EditedArgumentsInJSON != nil {
		res, err := i.InvokableTool.InvokableRun(ctx, *result.EditedArgumentsInJSON, opts...)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("after presenting the tool call info to the user, the user explilcitly changed tool call arguments to %s. Tool called, final result: %s",
			*result.EditedArgumentsInJSON, res), nil
	}

	return "", fmt.Errorf("invalid review result for tool '%s'", toolInfo.Name)
}
