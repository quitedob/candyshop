package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	tradeModels "candypro/api/internal/models/trade"
)

type fakeSaver struct {
	saved []*tradeModels.QuotationReview
	err   error
}

func (f *fakeSaver) SaveQuotationReview(_ context.Context, qr *tradeModels.QuotationReview) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, qr)
	return nil
}

func unmarshalReviewResponse(t *testing.T, raw string) QuotationReviewResponse {
	t.Helper()
	var resp QuotationReviewResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func TestQuotationHumanReviewTool_PersistsViaSaver(t *testing.T) {
	saver := &fakeSaver{}
	tl, err := NewQuotationHumanReviewTool(context.Background(), saver)
	if err != nil {
		t.Fatalf("NewQuotationHumanReviewTool: %v", err)
	}
	out := invokeTool(t, tl,
		`{"customer_ref":"INQ-1","currency":"USD","total_amount":1250.5,"line_items_json":"[{\"sku\":\"gum\",\"qty\":10}]","strategy_notes":"volume discount","suggested_discount_pct":5}`)
	resp := unmarshalReviewResponse(t, out)
	if resp.Status != "pending_human_review" {
		t.Fatalf("expected pending_human_review, got %q", resp.Status)
	}
	if len(saver.saved) != 1 {
		t.Fatalf("expected 1 saved review, got %d", len(saver.saved))
	}
	qr := saver.saved[0]
	if qr.Status != tradeModels.QuotationReviewStatusPending {
		t.Fatalf("expected pending status, got %q", qr.Status)
	}
	if qr.CustomerRef != "INQ-1" || qr.TotalAmount != 1250.5 || qr.SuggestedDiscountPct != 5 {
		t.Fatalf("unexpected review fields: %+v", qr)
	}
	if len(qr.LineItems) == 0 {
		t.Fatal("expected line items persisted")
	}
}

func TestQuotationHumanReviewTool_SaverErrorFailsQueue(t *testing.T) {
	saver := &fakeSaver{err: errors.New("db down")}
	tl, _ := NewQuotationHumanReviewTool(context.Background(), saver)
	out := invokeTool(t, tl, `{"customer_ref":"INQ-2","currency":"USD","total_amount":10}`)
	resp := unmarshalReviewResponse(t, out)
	if resp.Status != "review_queue_failed" {
		t.Fatalf("expected review_queue_failed, got %q", resp.Status)
	}
}

func TestQuotationHumanReviewTool_NilSaverDegrades(t *testing.T) {
	tl, _ := NewQuotationHumanReviewTool(context.Background(), nil)
	out := invokeTool(t, tl, `{"customer_ref":"INQ-3","currency":"USD","total_amount":10}`)
	resp := unmarshalReviewResponse(t, out)
	if resp.Status != "pending_human_review" {
		t.Fatalf("expected pending_human_review, got %q", resp.Status)
	}
}
