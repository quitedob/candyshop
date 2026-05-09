package admin

import (
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Trade Document CRUD (TradeDocument envelope) ---

type adminCreateTradeDocumentRequest struct {
	Type            string  `json:"type" binding:"required"`
	DocNumber       string  `json:"docNumber"`
	Status          string  `json:"status"`
	SourceOrderID   *string `json:"sourceOrderId"`
	LineageSource   string  `json:"lineageSource"`
}

// AdminCreateTradeDocument creates a new trade document for a transaction.
func (h *Handler) AdminCreateTradeDocument(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}

	var req adminCreateTradeDocumentRequest
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	docNumber := strings.TrimSpace(req.DocNumber)
	if docNumber == "" {
		docNumber = generateDocNo(docTypePrefix(req.Type))
	}

	doc := &modelsTrade.TradeDocument{
		TransactionID: id,
		Type:          req.Type,
		DocNumber:     docNumber,
		Status:        req.Status,
	}
	if doc.Status == "" {
		doc.Status = modelsTrade.TradeStatusDraft
	}
	if req.SourceOrderID != nil {
		if so := strings.TrimSpace(*req.SourceOrderID); so != "" {
			doc.SourceOrderID = &so
		}
	}
	if ls := strings.TrimSpace(req.LineageSource); ls != "" {
		doc.LineageSource = ls
	} else if doc.SourceOrderID != nil {
		doc.LineageSource = "order_derived"
	}

	if err := h.services.Trade.AddDocument(c.Request.Context(), doc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_doc_create_failed")
		return
	}
	c.JSON(http.StatusCreated, doc)
}

// AdminDeleteTradeDocument removes a trade document.
func (h *Handler) AdminDeleteTradeDocument(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	docID, err := parseUintParam(c, "docId")
	if err != nil {
		utils.InvalidResp(c, "invalid_document_id")
		return
	}
	if _, fetchErr := h.services.Trade.GetDocument(c.Request.Context(), docID); fetchErr != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	if err := h.services.Trade.DeleteDocument(c.Request.Context(), docID); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_doc_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted", "id": docID})
}

// --- SalesContract ---

// AdminGetSalesContract returns the sales contract for a trade transaction.
func (h *Handler) AdminGetSalesContract(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	sc, err := h.services.TradeDocDetail.GetSalesContract(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "sales_contract_not_found")
		return
	}
	c.JSON(http.StatusOK, sc)
}

// AdminCreateSalesContract creates a sales contract for a trade transaction.
func (h *Handler) AdminCreateSalesContract(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	var req struct {
		ContractNo      string   `json:"contractNo"`
		BuyerName       string   `json:"buyerName"`
		SellerName      string   `json:"sellerName"`
		Incoterms       string   `json:"incoterms"`
		TermsOfPayment  string   `json:"termsOfPayment"`
		QualityStandard string   `json:"qualityStandard"`
		TotalAmount     float64  `json:"totalAmount"`
		Currency        string   `json:"currency"`
		SignDate         *string `json:"signDate"`
		ValidUntil      *string `json:"validUntil"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	sc := &modelsTrade.SalesContract{
		TransactionID:   id,
		ContractNo:      req.ContractNo,
		BuyerName:       req.BuyerName,
		SellerName:      req.SellerName,
		Incoterms:       req.Incoterms,
		TermsOfPayment:  req.TermsOfPayment,
		QualityStandard: req.QualityStandard,
		TotalAmount:     req.TotalAmount,
		Currency:        req.Currency,
	}
	if req.SignDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.SignDate); parseErr == nil {
			sc.SignDate = &t
		}
	}
	if req.ValidUntil != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil {
			sc.ValidUntil = &t
		}
	}
	if err := h.services.TradeDocDetail.CreateSalesContract(c.Request.Context(), sc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "sales_contract_create_failed")
		return
	}
	c.JSON(http.StatusCreated, sc)
}

// AdminUpdateSalesContract updates an existing sales contract.
func (h *Handler) AdminUpdateSalesContract(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	sc, err := h.services.TradeDocDetail.GetSalesContract(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "sales_contract_not_found")
		return
	}
	var req struct {
		BuyerName       string  `json:"buyerName"`
		SellerName      string  `json:"sellerName"`
		Incoterms       string  `json:"incoterms"`
		TermsOfPayment  string  `json:"termsOfPayment"`
		QualityStandard string  `json:"qualityStandard"`
		TotalAmount     float64 `json:"totalAmount"`
		Currency        string  `json:"currency"`
		SignDate        *string `json:"signDate"`
		ValidUntil      *string `json:"validUntil"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.BuyerName != "" {
		sc.BuyerName = req.BuyerName
	}
	if req.SellerName != "" {
		sc.SellerName = req.SellerName
	}
	if req.Incoterms != "" {
		sc.Incoterms = req.Incoterms
	}
	if req.TermsOfPayment != "" {
		sc.TermsOfPayment = req.TermsOfPayment
	}
	if req.QualityStandard != "" {
		sc.QualityStandard = req.QualityStandard
	}
	if req.TotalAmount > 0 {
		sc.TotalAmount = req.TotalAmount
	}
	if req.Currency != "" {
		sc.Currency = req.Currency
	}
	if req.SignDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.SignDate); parseErr == nil {
			sc.SignDate = &t
		}
	}
	if req.ValidUntil != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil {
			sc.ValidUntil = &t
		}
	}
	if err := h.services.TradeDocDetail.UpdateSalesContract(c.Request.Context(), sc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "sales_contract_update_failed")
		return
	}
	c.JSON(http.StatusOK, sc)
}

// --- PackingList ---

// AdminGetPackingList returns the packing list for a trade transaction.
func (h *Handler) AdminGetPackingList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	pl, err := h.services.TradeDocDetail.GetPackingList(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "packing_list_not_found")
		return
	}
	c.JSON(http.StatusOK, pl)
}

// AdminCreatePackingList creates a packing list for a trade transaction.
func (h *Handler) AdminCreatePackingList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	var req struct {
		PLNumber         string  `json:"plNumber"`
		TotalCartons     int     `json:"totalCartons"`
		TotalPallets     int     `json:"totalPallets"`
		TotalGrossWeight float64 `json:"totalGrossWeight"`
		TotalNetWeight   float64 `json:"totalNetWeight"`
		TotalVolume      float64 `json:"totalVolume"`
		MarksAndNumbers  string  `json:"marksAndNumbers"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	pl := &modelsTrade.PackingList{
		TransactionID:    id,
		PLNumber:         req.PLNumber,
		TotalCartons:     req.TotalCartons,
		TotalPallets:     req.TotalPallets,
		TotalGrossWeight: req.TotalGrossWeight,
		TotalNetWeight:   req.TotalNetWeight,
		TotalVolume:      req.TotalVolume,
		MarksAndNumbers:  req.MarksAndNumbers,
	}
	if err := h.services.TradeDocDetail.CreatePackingList(c.Request.Context(), pl); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "packing_list_create_failed")
		return
	}
	c.JSON(http.StatusCreated, pl)
}

// AdminUpdatePackingList updates an existing packing list.
func (h *Handler) AdminUpdatePackingList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	pl, err := h.services.TradeDocDetail.GetPackingList(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "packing_list_not_found")
		return
	}
	var req struct {
		TotalCartons     int     `json:"totalCartons"`
		TotalPallets     int     `json:"totalPallets"`
		TotalGrossWeight float64 `json:"totalGrossWeight"`
		TotalNetWeight   float64 `json:"totalNetWeight"`
		TotalVolume      float64 `json:"totalVolume"`
		MarksAndNumbers  string  `json:"marksAndNumbers"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.TotalCartons > 0 {
		pl.TotalCartons = req.TotalCartons
	}
	if req.TotalPallets > 0 {
		pl.TotalPallets = req.TotalPallets
	}
	if req.TotalGrossWeight > 0 {
		pl.TotalGrossWeight = req.TotalGrossWeight
	}
	if req.TotalNetWeight > 0 {
		pl.TotalNetWeight = req.TotalNetWeight
	}
	if req.TotalVolume > 0 {
		pl.TotalVolume = req.TotalVolume
	}
	if req.MarksAndNumbers != "" {
		pl.MarksAndNumbers = req.MarksAndNumbers
	}
	if err := h.services.TradeDocDetail.UpdatePackingList(c.Request.Context(), pl); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "packing_list_update_failed")
		return
	}
	c.JSON(http.StatusOK, pl)
}

// --- CertificateOfOrigin ---

// AdminGetCertificateOfOrigin returns the certificate of origin for a transaction.
func (h *Handler) AdminGetCertificateOfOrigin(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	coo, err := h.services.TradeDocDetail.GetCertificateOfOrigin(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "coo_not_found")
		return
	}
	c.JSON(http.StatusOK, coo)
}

// AdminCreateCertificateOfOrigin creates a COO for a trade transaction.
func (h *Handler) AdminCreateCertificateOfOrigin(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	var req struct {
		CertificateNo      string  `json:"certificateNo"`
		CountryOfOrigin    string  `json:"countryOfOrigin" binding:"required"`
		DestinationCountry string  `json:"destinationCountry" binding:"required"`
		FTAType            string  `json:"ftaType"`
		HSCode             string  `json:"hsCode"`
		IssueDate          *string `json:"issueDate"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	coo := &modelsTrade.CertificateOfOrigin{
		TransactionID:      id,
		CertificateNo:      req.CertificateNo,
		CountryOfOrigin:    req.CountryOfOrigin,
		DestinationCountry: req.DestinationCountry,
		FTAType:            req.FTAType,
		HSCode:             req.HSCode,
	}
	if req.IssueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.IssueDate); parseErr == nil {
			coo.IssueDate = &t
		}
	}
	if err := h.services.TradeDocDetail.CreateCertificateOfOrigin(c.Request.Context(), coo); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "coo_create_failed")
		return
	}
	c.JSON(http.StatusCreated, coo)
}

// AdminUpdateCertificateOfOrigin updates a COO.
func (h *Handler) AdminUpdateCertificateOfOrigin(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	coo, err := h.services.TradeDocDetail.GetCertificateOfOrigin(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "coo_not_found")
		return
	}
	var req struct {
		CountryOfOrigin    string  `json:"countryOfOrigin"`
		DestinationCountry string  `json:"destinationCountry"`
		FTAType            string  `json:"ftaType"`
		HSCode             string  `json:"hsCode"`
		IssueDate          *string `json:"issueDate"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.CountryOfOrigin != "" {
		coo.CountryOfOrigin = req.CountryOfOrigin
	}
	if req.DestinationCountry != "" {
		coo.DestinationCountry = req.DestinationCountry
	}
	if req.FTAType != "" {
		coo.FTAType = req.FTAType
	}
	if req.HSCode != "" {
		coo.HSCode = req.HSCode
	}
	if req.IssueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.IssueDate); parseErr == nil {
			coo.IssueDate = &t
		}
	}
	if err := h.services.TradeDocDetail.UpdateCertificateOfOrigin(c.Request.Context(), coo); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "coo_update_failed")
		return
	}
	c.JSON(http.StatusOK, coo)
}

// --- HealthCertificate ---

// AdminGetHealthCertificate returns the health certificate for a transaction.
func (h *Handler) AdminGetHealthCertificate(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	hc, err := h.services.TradeDocDetail.GetHealthCertificate(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "health_cert_not_found")
		return
	}
	c.JSON(http.StatusOK, hc)
}

// AdminCreateHealthCertificate creates a health certificate for a transaction.
func (h *Handler) AdminCreateHealthCertificate(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	var req struct {
		CertificateNo      string  `json:"certificateNo"`
		IssuingAuthority   string  `json:"issuingAuthority"`
		Consignor          string  `json:"consignor"`
		Consignee          string  `json:"consignee"`
		DestinationCountry string  `json:"destinationCountry" binding:"required"`
		IssueDate          *string `json:"issueDate"`
		ValidUntil         *string `json:"validUntil"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	hc := &modelsTrade.HealthCertificate{
		TransactionID:      id,
		CertificateNo:      req.CertificateNo,
		IssuingAuthority:   req.IssuingAuthority,
		Consignor:          req.Consignor,
		Consignee:          req.Consignee,
		DestinationCountry: req.DestinationCountry,
	}
	if req.IssueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.IssueDate); parseErr == nil {
			hc.IssueDate = &t
		}
	}
	if req.ValidUntil != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil {
			hc.ValidUntil = &t
		}
	}
	if err := h.services.TradeDocDetail.CreateHealthCertificate(c.Request.Context(), hc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "health_cert_create_failed")
		return
	}
	c.JSON(http.StatusCreated, hc)
}

// AdminUpdateHealthCertificate updates a health certificate.
func (h *Handler) AdminUpdateHealthCertificate(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	hc, err := h.services.TradeDocDetail.GetHealthCertificate(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "health_cert_not_found")
		return
	}
	var req struct {
		IssuingAuthority   string  `json:"issuingAuthority"`
		Consignor          string  `json:"consignor"`
		Consignee          string  `json:"consignee"`
		DestinationCountry string  `json:"destinationCountry"`
		IssueDate          *string `json:"issueDate"`
		ValidUntil         *string `json:"validUntil"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.IssuingAuthority != "" {
		hc.IssuingAuthority = req.IssuingAuthority
	}
	if req.Consignor != "" {
		hc.Consignor = req.Consignor
	}
	if req.Consignee != "" {
		hc.Consignee = req.Consignee
	}
	if req.DestinationCountry != "" {
		hc.DestinationCountry = req.DestinationCountry
	}
	if req.IssueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.IssueDate); parseErr == nil {
			hc.IssueDate = &t
		}
	}
	if req.ValidUntil != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil {
			hc.ValidUntil = &t
		}
	}
	if err := h.services.TradeDocDetail.UpdateHealthCertificate(c.Request.Context(), hc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "health_cert_update_failed")
		return
	}
	c.JSON(http.StatusOK, hc)
}

// --- SettlementRecord ---

// AdminGetSettlements returns all settlement milestones for a transaction.
func (h *Handler) AdminGetSettlements(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	records, err := h.services.TradeDocDetail.ListSettlements(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settlement_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, records)
}

// AdminCreateSettlement creates a payment milestone for a transaction.
func (h *Handler) AdminCreateSettlement(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	var req struct {
		PaymentMethod string  `json:"paymentMethod" binding:"required"`
		AmountDue     float64 `json:"amountDue" binding:"required"`
		Currency      string  `json:"currency"`
		DueDate       *string `json:"dueDate"`
		LCReference   string  `json:"lcReference"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	record := &modelsTrade.SettlementRecord{
		TransactionID: id,
		PaymentMethod: req.PaymentMethod,
		AmountDue:     req.AmountDue,
		Currency:      req.Currency,
		LCReference:   req.LCReference,
	}
	if req.DueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.DueDate); parseErr == nil {
			record.DueDate = &t
		}
	}
	if err := h.services.TradeDocDetail.CreateSettlement(c.Request.Context(), record); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settlement_create_failed")
		return
	}
	c.JSON(http.StatusCreated, record)
}

// AdminUpdateSettlement updates a settlement record.
func (h *Handler) AdminUpdateSettlement(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	settlementID, err := parseUintParam(c, "settlementId")
	if err != nil {
		utils.InvalidResp(c, "invalid_settlement_id")
		return
	}
	record, err := h.services.TradeDocDetail.GetSettlement(c.Request.Context(), settlementID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	var req struct {
		Status      string  `json:"status"`
		AmountPaid  float64 `json:"amountPaid"`
		PaymentDate *string `json:"paymentDate"`
		LCReference string  `json:"lcReference"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Status != "" {
		record.Status = req.Status
	}
	if req.AmountPaid >= 0 {
		record.AmountPaid = req.AmountPaid
	}
	if req.PaymentDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.PaymentDate); parseErr == nil {
			record.PaymentDate = &t
		}
	}
	if req.LCReference != "" {
		record.LCReference = req.LCReference
	}
	if err := h.services.TradeDocDetail.UpdateSettlement(c.Request.Context(), record); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settlement_update_failed")
		return
	}
	c.JSON(http.StatusOK, record)
}

// AdminDeleteSettlement removes a settlement record.
func (h *Handler) AdminDeleteSettlement(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	settlementID, err := parseUintParam(c, "settlementId")
	if err != nil {
		utils.InvalidResp(c, "invalid_settlement_id")
		return
	}
	if err := h.services.TradeDocDetail.DeleteSettlement(c.Request.Context(), settlementID); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settlement_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Settlement deleted", "id": settlementID})
}

// --- Compliance ---

// AdminGetCompliance returns all compliance requirements for a trade transaction.
func (h *Handler) AdminGetCompliance(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}
	comps, err := h.services.Trade.ListCompliance(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_compliance_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, comps)
}

// AdminUpdateCompliance updates a compliance requirement status.
func (h *Handler) AdminUpdateCompliance(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	compID, err := parseUintParam(c, "compId")
	if err != nil {
		utils.InvalidResp(c, "invalid_compliance_id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}
	if err := h.services.Trade.UpdateComplianceStatus(c.Request.Context(), compID, req.Status); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_compliance_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Compliance status updated", "id": compID, "status": req.Status})
}

// docTypePrefix maps a document type constant to a short prefix for number generation.
func docTypePrefix(docType string) string {
	switch strings.ToUpper(docType) {
	case modelsTrade.DocTypeProformaInvoice:
		return "PI"
	case modelsTrade.DocTypeCommercialInvoice:
		return "CI"
	case modelsTrade.DocTypeSalesContract:
		return "SC"
	case modelsTrade.DocTypePackingList:
		return "PL"
	case modelsTrade.DocTypeBillOfLading:
		return "BL"
	case modelsTrade.DocTypeHealthCertificate:
		return "HC"
	case modelsTrade.DocTypeOriginCertificate:
		return "COO"
	default:
		return "DOC"
	}
}

// generateDocNo generates a sequential document number with a given prefix.
func generateDocNo(prefix string) string {
	return fmt.Sprintf("%s-%s-%06d", prefix, time.Now().UTC().Format("200601"), time.Now().UnixNano()%1000000)
}

// --- ProformaInvoice ---

func (h *Handler) AdminGetProformaInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	pi, err := h.services.TradeDocDetail.GetProformaInvoice(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "proforma_invoice_not_found"); return }
	c.JSON(http.StatusOK, pi)
}

func (h *Handler) AdminCreateProformaInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	var req modelsTrade.ProformaInvoice
	if !utils.BindJSONOrInvalid(c, &req) { return }
	req.TransactionID = id
	if err := h.services.TradeDocDetail.CreateProformaInvoice(c.Request.Context(), &req); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "proforma_invoice_create_failed"); return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *Handler) AdminUpdateProformaInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	pi, err := h.services.TradeDocDetail.GetProformaInvoice(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "proforma_invoice_not_found"); return }
	var req struct {
		BuyerName      string   `json:"buyerName"`
		SellerName     string   `json:"sellerName"`
		Incoterms      string   `json:"incoterms"`
		TermsOfPayment string   `json:"termsOfPayment"`
		TotalAmount    float64  `json:"totalAmount"`
		Currency       string   `json:"currency"`
		BankDetails    string   `json:"bankDetails"`
		Notes          string   `json:"notes"`
		Status         string   `json:"status"`
		ValidUntil     *string  `json:"validUntil"`
	}
	if !utils.BindJSONOrInvalid(c, &req) { return }
	if req.BuyerName != "" { pi.BuyerName = req.BuyerName }
	if req.SellerName != "" { pi.SellerName = req.SellerName }
	if req.Incoterms != "" { pi.Incoterms = req.Incoterms }
	if req.TermsOfPayment != "" { pi.TermsOfPayment = req.TermsOfPayment }
	if req.TotalAmount > 0 { pi.TotalAmount = req.TotalAmount }
	if req.Currency != "" { pi.Currency = req.Currency }
	if req.BankDetails != "" { pi.BankDetails = req.BankDetails }
	if req.Notes != "" { pi.Notes = req.Notes }
	if req.Status != "" { pi.Status = req.Status }
	if req.ValidUntil != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil { pi.ValidUntil = &t }
	}
	if err := h.services.TradeDocDetail.UpdateProformaInvoice(c.Request.Context(), pi); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "proforma_invoice_update_failed"); return
	}
	c.JSON(http.StatusOK, pi)
}

// --- CommercialInvoice ---

func (h *Handler) AdminGetCommercialInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	ci, err := h.services.TradeDocDetail.GetCommercialInvoice(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "commercial_invoice_not_found"); return }
	c.JSON(http.StatusOK, ci)
}

func (h *Handler) AdminCreateCommercialInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	var req modelsTrade.CommercialInvoice
	if !utils.BindJSONOrInvalid(c, &req) { return }
	req.TransactionID = id
	if err := h.services.TradeDocDetail.CreateCommercialInvoice(c.Request.Context(), &req); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "commercial_invoice_create_failed"); return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *Handler) AdminUpdateCommercialInvoice(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	ci, err := h.services.TradeDocDetail.GetCommercialInvoice(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "commercial_invoice_not_found"); return }
	var req struct {
		BuyerName      string  `json:"buyerName"`
		SellerName     string  `json:"sellerName"`
		Incoterms      string  `json:"incoterms"`
		TermsOfPayment string  `json:"termsOfPayment"`
		TotalAmount    float64 `json:"totalAmount"`
		Currency       string  `json:"currency"`
		BankDetails    string  `json:"bankDetails"`
		Notes          string  `json:"notes"`
		Status         string  `json:"status"`
		PINumber       string  `json:"piNumber"`
		ShipDate       *string `json:"shipDate"`
	}
	if !utils.BindJSONOrInvalid(c, &req) { return }
	if req.BuyerName != "" { ci.BuyerName = req.BuyerName }
	if req.SellerName != "" { ci.SellerName = req.SellerName }
	if req.Incoterms != "" { ci.Incoterms = req.Incoterms }
	if req.TermsOfPayment != "" { ci.TermsOfPayment = req.TermsOfPayment }
	if req.TotalAmount > 0 { ci.TotalAmount = req.TotalAmount }
	if req.Currency != "" { ci.Currency = req.Currency }
	if req.BankDetails != "" { ci.BankDetails = req.BankDetails }
	if req.Notes != "" { ci.Notes = req.Notes }
	if req.Status != "" { ci.Status = req.Status }
	if req.PINumber != "" { ci.PINumber = req.PINumber }
	if req.ShipDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ShipDate); parseErr == nil { ci.ShipDate = &t }
	}
	if err := h.services.TradeDocDetail.UpdateCommercialInvoice(c.Request.Context(), ci); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "commercial_invoice_update_failed"); return
	}
	c.JSON(http.StatusOK, ci)
}

// --- BillOfLading ---

func (h *Handler) AdminGetBillOfLading(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	bl, err := h.services.TradeDocDetail.GetBillOfLading(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "bill_of_lading_not_found"); return }
	c.JSON(http.StatusOK, bl)
}

func (h *Handler) AdminCreateBillOfLading(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	var req modelsTrade.BillOfLading
	if !utils.BindJSONOrInvalid(c, &req) { return }
	req.TransactionID = id
	if err := h.services.TradeDocDetail.CreateBillOfLading(c.Request.Context(), &req); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "bill_of_lading_create_failed"); return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *Handler) AdminUpdateBillOfLading(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil { utils.InvalidResp(c, "invalid_transaction_id"); return }
	bl, err := h.services.TradeDocDetail.GetBillOfLading(c.Request.Context(), id)
	if err != nil { utils.ErrorResp(c, http.StatusNotFound, "bill_of_lading_not_found"); return }
	var req struct {
		Shipper          string  `json:"shipper"`
		Consignee        string  `json:"consignee"`
		NotifyParty      string  `json:"notifyParty"`
		CarrierName      string  `json:"carrierName"`
		VesselVoyage     string  `json:"vesselVoyage"`
		PortOfLoading    string  `json:"portOfLoading"`
		PortOfDischarge  string  `json:"portOfDischarge"`
		PlaceOfDelivery  string  `json:"placeOfDelivery"`
		FreightTerms     string  `json:"freightTerms"`
		GoodsDescription string  `json:"goodsDescription"`
		GrossWeight      float64 `json:"grossWeight"`
		Measurement      float64 `json:"measurement"`
		NumberOfPackages int     `json:"numberOfPackages"`
		Status           string  `json:"status"`
		OnBoardDate      *string `json:"onBoardDate"`
	}
	if !utils.BindJSONOrInvalid(c, &req) { return }
	if req.Shipper != "" { bl.Shipper = req.Shipper }
	if req.Consignee != "" { bl.Consignee = req.Consignee }
	if req.NotifyParty != "" { bl.NotifyParty = req.NotifyParty }
	if req.CarrierName != "" { bl.CarrierName = req.CarrierName }
	if req.VesselVoyage != "" { bl.VesselVoyage = req.VesselVoyage }
	if req.PortOfLoading != "" { bl.PortOfLoading = req.PortOfLoading }
	if req.PortOfDischarge != "" { bl.PortOfDischarge = req.PortOfDischarge }
	if req.PlaceOfDelivery != "" { bl.PlaceOfDelivery = req.PlaceOfDelivery }
	if req.FreightTerms != "" { bl.FreightTerms = req.FreightTerms }
	if req.GoodsDescription != "" { bl.GoodsDescription = req.GoodsDescription }
	if req.GrossWeight > 0 { bl.GrossWeight = req.GrossWeight }
	if req.Measurement > 0 { bl.Measurement = req.Measurement }
	if req.NumberOfPackages > 0 { bl.NumberOfPackages = req.NumberOfPackages }
	if req.Status != "" { bl.Status = req.Status }
	if req.OnBoardDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.OnBoardDate); parseErr == nil { bl.OnBoardDate = &t }
	}
	if err := h.services.TradeDocDetail.UpdateBillOfLading(c.Request.Context(), bl); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "bill_of_lading_update_failed"); return
	}
	c.JSON(http.StatusOK, bl)
}
