package admin

import (
	"testing"

	modelsProduct "candypro/api/internal/models/product"
)

func TestAdminUpdateOEMProjectRequest_AdminNotesSeparateFromCustomerNotes(t *testing.T) {
	customerNotes := "Customer wants sour belt with halal cert"
	adminNotes := "Internal: schedule sampling next week"
	project := &modelsProduct.OEMProject{
		ID:          "oem-1",
		ProductName: "Sour Belt Candy OEM",
		Status:      modelsProduct.OEMStatusInquiry,
		Notes:       customerNotes,
	}

	req := adminUpdateOEMProjectRequest{
		AdminNotes: strPtr(adminNotes),
	}
	if req.AdminNotes != nil {
		project.AdminNotes = *req.AdminNotes
	}

	if project.Notes != customerNotes {
		t.Fatalf("customer notes overwritten: got %q, want %q", project.Notes, customerNotes)
	}
	if project.AdminNotes != adminNotes {
		t.Fatalf("admin notes = %q, want %q", project.AdminNotes, adminNotes)
	}
}

func strPtr(s string) *string { return &s }
