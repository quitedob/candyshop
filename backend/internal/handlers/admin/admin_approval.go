package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ── Buyer Organizations ──

// AdminGetOrganizations lists all buyer organizations.
func (h *Handler) AdminGetOrganizations(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orgs, err := h.services.Approval.GetOrganizations(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_list_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orgs})
}

// AdminCreateOrganization creates a new buyer organization.
func (h *Handler) AdminCreateOrganization(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Name              string  `json:"name" binding:"required"`
		CreditLimit       float64 `json:"creditLimit"`
		PaymentTerms      string  `json:"paymentTerms"`
		ApprovalThreshold float64 `json:"approvalThreshold"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	org := &modelsOrder.BuyerOrganization{
		Name:              req.Name,
		CreditLimit:       req.CreditLimit,
		PaymentTerms:      req.PaymentTerms,
		ApprovalThreshold: req.ApprovalThreshold,
	}
	if err := h.services.Approval.CreateOrganization(c.Request.Context(), org); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_create_failed")
		return
	}
	c.JSON(http.StatusCreated, org)
}

// AdminUpdateOrganization updates a buyer organization.
func (h *Handler) AdminUpdateOrganization(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_id")
		return
	}
	org, err := h.services.Approval.GetOrganization(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "org_not_found")
		return
	}
	var req struct {
		Name              *string  `json:"name"`
		CreditLimit       *float64 `json:"creditLimit"`
		PaymentTerms      *string  `json:"paymentTerms"`
		ApprovalThreshold *float64 `json:"approvalThreshold"`
		IsActive          *bool    `json:"isActive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.CreditLimit != nil {
		org.CreditLimit = *req.CreditLimit
	}
	if req.PaymentTerms != nil {
		org.PaymentTerms = *req.PaymentTerms
	}
	if req.ApprovalThreshold != nil {
		org.ApprovalThreshold = *req.ApprovalThreshold
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}
	if err := h.services.Approval.UpdateOrganization(c.Request.Context(), org); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_update_failed")
		return
	}
	c.JSON(http.StatusOK, org)
}

// ── Org Members ──

// AdminAddOrgMember adds a user to a buyer organization.
func (h *Handler) AdminAddOrgMember(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		OrganizationID uint   `json:"organizationId" binding:"required"`
		UserID         string `json:"userId" binding:"required"`
		Role           string `json:"role" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	m := &modelsOrder.OrgMember{
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
		Role:           modelsOrder.OrgMemberRole(req.Role),
	}
	if err := h.services.Approval.AddMember(c.Request.Context(), m); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_member_add_failed")
		return
	}
	c.JSON(http.StatusCreated, m)
}

// AdminRemoveOrgMember removes a user from a buyer organization.
func (h *Handler) AdminRemoveOrgMember(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_id")
		return
	}
	if err := h.services.Approval.RemoveMember(c.Request.Context(), uint(id)); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_member_remove_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// AdminGetOrgMembers lists members of a buyer organization.
func (h *Handler) AdminGetOrgMembers(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orgID, err := strconv.ParseUint(c.Query("orgId"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_org_id")
		return
	}
	// H-8: query members via indexed FindByOrgID instead of loading every member with
	// FindByUserID("") and filtering in Go. The previous implementation pulled the
	// entire org_members table on every admin page load.
	members, err := h.services.Approval.GetOrgMembers(c.Request.Context(), uint(orgID))
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "org_members_fetch_failed")
		return
	}
	if members == nil {
		members = []modelsOrder.OrgMember{}
	}
	c.JSON(http.StatusOK, gin.H{"data": members})
}

// ── Approval Actions ──

// AdminApproveOrder approves a pending-approval order.
func (h *Handler) AdminApproveOrder(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID := c.GetString("userID")

	var req struct {
		Comment string `json:"comment"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if err := h.services.Approval.ApproveOrderByAdmin(c.Request.Context(), orderID, userID, req.Comment); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_approve_failed")
		return
	}
	if order, oerr := h.services.Order.GetOrder(c.Request.Context(), orderID); oerr == nil {
		h.notifyOrderApprovalResult(c, order, true)
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

// AdminRejectOrder rejects a pending-approval order.
func (h *Handler) AdminRejectOrder(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID := c.GetString("userID")

	var req struct {
		Comment string `json:"comment"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if err := h.services.Approval.RejectOrderByAdmin(c.Request.Context(), orderID, userID, req.Comment); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_reject_failed")
		return
	}
	if order, oerr := h.services.Order.GetOrder(c.Request.Context(), orderID); oerr == nil {
		h.notifyOrderApprovalResult(c, order, false)
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

// AdminGetApprovalHistory returns approval actions for an order.
func (h *Handler) AdminGetApprovalHistory(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	actions, err := h.services.Approval.GetApprovalHistory(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "approval_history_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": actions})
}
