package trade

// DocMetaData holds basic file information
type DocMetaData struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
}

// RawDocument represents the input to the classifier
type RawDocument struct {
	Metadata DocMetaData `json:"metadata"`
	Text     string      `json:"text"` // Extracted text or OCR result
	Content  []byte      `json:"-"`    // Raw file bytes
}

// ClassifierResult represents the output of the classification node
type ClassifierResult struct {
	DocumentType string  `json:"document_type" jsonschema:"description=The type of trade document (e.g. Commercial Invoice, Packing List, Health Certificate)"`
	Confidence   float64 `json:"confidence" jsonschema:"description=Confidence score between 0 and 1"`
}

// ExtractedData represents a generalized output from the extraction node
type ExtractedData struct {
	DocumentType string         `json:"document_type"`
	Data         map[string]any `json:"data"`
}

type VerificationReport struct {
	Passed           bool     `json:"passed" jsonschema:"description=True if all documents are consistent and complete"`
	Conflicts        []string `json:"conflicts" jsonschema:"description=List of data inconsistencies across documents"`
	MissingDocuments []string `json:"missing_documents" jsonschema:"description=List of missing mandatory documents like Health Certificates"`
	ManualReviewReq  bool     `json:"manual_review_required" jsonschema:"description=True if human intervention is needed"`
}

// HitlRequest is passed to the Human in the loop tool
type HitlRequest struct {
	Report *VerificationReport `json:"report"`
	Data   []*ExtractedData    `json:"data"`
}
