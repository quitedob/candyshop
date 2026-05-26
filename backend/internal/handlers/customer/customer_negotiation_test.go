package customer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	servicesCommon "candypro/api/internal/services/common"
	inquiryService "candypro/api/internal/services/inquiry"
	orderService "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
)

// fakeNegotiationInquiryRepo returns a fixed inquiry for ownership checks.
type fakeNegotiationInquiryRepo struct {
	inquiry *modelsProduct.Inquiry
	fakeInquiryRepo
}

func (f *fakeNegotiationInquiryRepo) FindByID(ctx context.Context, id string) (*modelsProduct.Inquiry, error) {
	if f.inquiry != nil && f.inquiry.ID == id {
		cp := *f.inquiry
		return &cp, nil
	}
	return nil, context.Canceled
}

// fakeNegotiationRepo stores offers in memory for testing negotiation flow.
type fakeNegotiationRepo struct {
	offers []modelsOrder.NegotiationOffer
}

func (r *fakeNegotiationRepo) FindByInquiryID(ctx context.Context, inquiryID string) ([]modelsOrder.NegotiationOffer, error) {
	var result []modelsOrder.NegotiationOffer
	for _, o := range r.offers {
		if o.InquiryID == inquiryID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *fakeNegotiationRepo) FindByID(ctx context.Context, id string) (*modelsOrder.NegotiationOffer, error) {
	for i, o := range r.offers {
		if o.ID == id {
			return &r.offers[i], nil
		}
	}
	return nil, context.Canceled
}

func (r *fakeNegotiationRepo) Create(ctx context.Context, offer *modelsOrder.NegotiationOffer) error {
	r.offers = append(r.offers, *offer)
	return nil
}

func (r *fakeNegotiationRepo) Update(ctx context.Context, offer *modelsOrder.NegotiationOffer) error {
	for i, o := range r.offers {
		if o.ID == offer.ID {
			r.offers[i] = *offer
			return nil
		}
	}
	return nil
}

func (r *fakeNegotiationRepo) FindPendingByInquiryID(ctx context.Context, inquiryID string) (*modelsOrder.NegotiationOffer, error) {
	for _, o := range r.offers {
		if o.InquiryID == inquiryID && o.Status == "pending" {
			return &o, nil
		}
	}
	return nil, context.Canceled
}

// TransitionStatus 模拟 A-4 的条件 UPDATE：仅当当前 status==fromStatus 时才改为 toStatus。
// 返回受影响行数，0 表示并发竞争 / 已被改写。
func (r *fakeNegotiationRepo) TransitionStatus(ctx context.Context, id, fromStatus, toStatus string, updatedAt time.Time) (int64, error) {
	for i, o := range r.offers {
		if o.ID == id {
			if o.Status != fromStatus {
				return 0, nil
			}
			r.offers[i].Status = toStatus
			r.offers[i].UpdatedAt = updatedAt
			return 1, nil
		}
	}
	return 0, nil
}

func TestCustomerCreateNegotiationOffer_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-1"
	inquiryID := "inq-nego-1"
	negoRepo := &fakeNegotiationRepo{}
	inquiryRepo := &fakeNegotiationInquiryRepo{
		inquiry: &modelsProduct.Inquiry{ID: inquiryID, UserID: &userID},
	}
	handler := buildTestNegotiationHandler(negoRepo, inquiryRepo)

	rec := performCustomerCreateNegotiation(handler, userID, inquiryID, map[string]interface{}{
		"unitPrice":   2.5,
		"quantity":    100,
		"totalAmount": 250.0,
		"currency":    "USD",
		"message":     "Can we get a better price?",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if len(negoRepo.offers) != 1 {
		t.Fatalf("expected 1 offer created, got %d", len(negoRepo.offers))
	}
	o := negoRepo.offers[0]
	if o.SenderType != "customer" {
		t.Fatalf("expected senderType=customer, got %s", o.SenderType)
	}
	if o.Status != "pending" {
		t.Fatalf("expected status=pending, got %s", o.Status)
	}
	if o.InquiryID != inquiryID {
		t.Fatalf("expected inquiryID=%s, got %s", inquiryID, o.InquiryID)
	}
}

func TestCustomerCreateNegotiationOffer_InvalidAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-2"
	inquiryID := "inq-nego-2"
	negoRepo := &fakeNegotiationRepo{}
	inquiryRepo := &fakeNegotiationInquiryRepo{
		inquiry: &modelsProduct.Inquiry{ID: inquiryID, UserID: &userID},
	}
	handler := buildTestNegotiationHandler(negoRepo, inquiryRepo)

	rec := performCustomerCreateNegotiation(handler, userID, inquiryID, map[string]interface{}{
		"unitPrice":   2.5,
		"quantity":    100,
		"totalAmount": 0,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if len(negoRepo.offers) != 0 {
		t.Fatalf("expected 0 offers created, got %d", len(negoRepo.offers))
	}
}

func TestCustomerCreateNegotiationOffer_ForbiddenInquiry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-3"
	otherUserID := "u-other"
	inquiryID := "inq-nego-3"
	negoRepo := &fakeNegotiationRepo{}
	inquiryRepo := &fakeNegotiationInquiryRepo{
		inquiry: &modelsProduct.Inquiry{ID: inquiryID, UserID: &otherUserID},
	}
	handler := buildTestNegotiationHandler(negoRepo, inquiryRepo)

	rec := performCustomerCreateNegotiation(handler, userID, inquiryID, map[string]interface{}{
		"unitPrice":   2.5,
		"quantity":    100,
		"totalAmount": 250.0,
	})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if len(negoRepo.offers) != 0 {
		t.Fatalf("expected 0 offers created, got %d", len(negoRepo.offers))
	}
}

func TestCustomerGetNegotiationOffers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-4"
	inquiryID := "inq-nego-4"
	now := time.Now()
	negoRepo := &fakeNegotiationRepo{
		offers: []modelsOrder.NegotiationOffer{
			{ID: "offer-1", InquiryID: inquiryID, UserID: userID, SenderType: "admin", Status: "pending", TotalAmount: 200, Currency: "USD", CreatedAt: now, UpdatedAt: now},
			{ID: "offer-2", InquiryID: inquiryID, UserID: userID, SenderType: "customer", Status: "accepted", TotalAmount: 180, Currency: "USD", CreatedAt: now, UpdatedAt: now},
		},
	}
	inquiryRepo := &fakeNegotiationInquiryRepo{
		inquiry: &modelsProduct.Inquiry{ID: inquiryID, UserID: &userID},
	}
	handler := buildTestNegotiationHandler(negoRepo, inquiryRepo)

	rec := performCustomerGetNegotiation(handler, userID, inquiryID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array in response, got %v", resp)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(data))
	}
}

func TestCustomerAcceptNegotiationOffer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-5"
	offerID := "offer-accept-1"
	now := time.Now()
	negoRepo := &fakeNegotiationRepo{
		offers: []modelsOrder.NegotiationOffer{
			{ID: offerID, InquiryID: "inq-nego-5", UserID: userID, SenderType: "admin", Status: "pending", TotalAmount: 200, Currency: "USD", CreatedAt: now, UpdatedAt: now},
		},
	}
	handler := buildTestNegotiationHandler(negoRepo, &fakeNegotiationInquiryRepo{})

	rec := performCustomerAcceptNegotiation(handler, userID, offerID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	for _, o := range negoRepo.offers {
		if o.ID == offerID && o.Status != "accepted" {
			t.Fatalf("expected offer status=accepted, got %s", o.Status)
		}
	}
}

func TestCustomerAcceptNegotiationOffer_NotPending(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-6"
	offerID := "offer-re-accept"
	now := time.Now()
	negoRepo := &fakeNegotiationRepo{
		offers: []modelsOrder.NegotiationOffer{
			{ID: offerID, InquiryID: "inq-nego-6", UserID: userID, SenderType: "admin", Status: "accepted", TotalAmount: 200, Currency: "USD", CreatedAt: now, UpdatedAt: now},
		},
	}
	handler := buildTestNegotiationHandler(negoRepo, &fakeNegotiationInquiryRepo{})

	rec := performCustomerAcceptNegotiation(handler, userID, offerID)
	// 409 Conflict reflects "offer is not in pending state" (state-machine conflict),
	// the previous 400 was returned by the generic err.Error() handler before C-10
	// mapped the sentinel error to a stable code.
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestCustomerRejectNegotiationOffer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "u-nego-7"
	offerID := "offer-reject-1"
	now := time.Now()
	negoRepo := &fakeNegotiationRepo{
		offers: []modelsOrder.NegotiationOffer{
			{ID: offerID, InquiryID: "inq-nego-7", UserID: userID, SenderType: "admin", Status: "pending", TotalAmount: 500, Currency: "USD", CreatedAt: now, UpdatedAt: now},
		},
	}
	handler := buildTestNegotiationHandler(negoRepo, &fakeNegotiationInquiryRepo{})

	rec := performCustomerRejectNegotiation(handler, userID, offerID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	for _, o := range negoRepo.offers {
		if o.ID == offerID && o.Status != "rejected" {
			t.Fatalf("expected offer status=rejected, got %s", o.Status)
		}
	}
}

func buildTestNegotiationHandler(negoRepo *fakeNegotiationRepo, inquiryRepo *fakeNegotiationInquiryRepo) *Handler {
	cfg := &config.Config{}
	svcs := &servicesCommon.UserPortalServices{
		Negotiation: orderService.NewNegotiationService(negoRepo),
		Inquiry:     inquiryService.NewInquiryService(inquiryRepo, cfg),
	}
	return NewHandler(cfg, svcs, nil)
}

func performCustomerCreateNegotiation(handler *Handler, userID, inquiryID string, body map[string]interface{}) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/user/inquiries/"+inquiryID+"/negotiations", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: inquiryID}}
	c.Set("userID", userID)
	handler.CustomerCreateNegotiationOffer(c)
	return rec
}

func performCustomerGetNegotiation(handler *Handler, userID, inquiryID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/user/inquiries/"+inquiryID+"/negotiations", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: inquiryID}}
	c.Set("userID", userID)
	handler.CustomerGetNegotiationOffers(c)
	return rec
}

func performCustomerAcceptNegotiation(handler *Handler, userID, offerID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/user/inquiries/inq-x/negotiations/"+offerID+"/accept", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "offerId", Value: offerID}}
	c.Set("userID", userID)
	handler.CustomerAcceptNegotiationOffer(c)
	return rec
}

func performCustomerRejectNegotiation(handler *Handler, userID, offerID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/user/inquiries/inq-x/negotiations/"+offerID+"/reject", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "offerId", Value: offerID}}
	c.Set("userID", userID)
	handler.CustomerRejectNegotiationOffer(c)
	return rec
}
