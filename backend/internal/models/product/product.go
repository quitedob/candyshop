package product

import (
	"fmt"
	"time"

	common "candypro/api/internal/models/common"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Product represents a candy product
type Product struct {
	ID             string             `json:"id" gorm:"primaryKey"`
	Slug           string             `json:"slug" gorm:"uniqueIndex"`
	Name           string             `json:"name"`
	Summary        string             `json:"summary"`
	Description    string             `json:"description" gorm:"type:text"`
	Category       string             `json:"category"`
	CategorySlug   string             `json:"categorySlug" gorm:"index"`
	Thumbnail      string             `json:"thumbnail"`
	Images         common.StringArray `json:"images" gorm:"type:jsonb"`
	OEMAvailable   bool               `json:"oemAvailable"`
	HalalCertified bool               `json:"halalCertified"`
	Certifications common.StringArray `json:"certifications" gorm:"type:jsonb"`
	MOQ            int                `json:"moq"`
	StockQuantity  int                `json:"stockQuantity" gorm:"default:0"`
	LeadTime       string             `json:"leadTime"`
	Featured       bool               `json:"featured" gorm:"index"`
	Flavors        common.StringArray `json:"flavors" gorm:"type:jsonb"`
	Shapes         common.StringArray `json:"shapes" gorm:"type:jsonb"`
	Ingredients    string             `json:"ingredients"`
	Allergens      string             `json:"allergens"`
	ShelfLife      string             `json:"shelfLife"`
	Storage        string             `json:"storage"`
	HSCode         string             `json:"hsCode" gorm:"type:varchar(20)"` // Harmonized System code for customs
	CreatedBy      *string            `json:"createdBy" gorm:"index"`         // Audit field: user ID that created the record
	UpdatedBy      *string            `json:"updatedBy" gorm:"index"`         // Audit field: user ID that last updated the record
	Status         string             `json:"status" gorm:"default:'active'"` // draft, active, inactive
	BasePrice      float64            `json:"basePrice" gorm:"default:0"`     // Base unit price for server-side price resolution
	ViewCount      int                `json:"viewCount" gorm:"default:0"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt     `json:"-" gorm:"index"`
}

// Product status constants
const (
	ProductStatusDraft    = "draft"
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"
)

// IsValidProductStatus checks if a status value is valid.
func IsValidProductStatus(s string) bool {
	switch s {
	case ProductStatusDraft, ProductStatusActive, ProductStatusInactive:
		return true
	}
	return false
}

var validProductStatusTransitions = map[string][]string{
	ProductStatusDraft:    {ProductStatusActive},
	ProductStatusActive:   {ProductStatusInactive},
	ProductStatusInactive: {ProductStatusActive},
}

// ValidateProductStatusTransition returns an error if the transition is not allowed.
func ValidateProductStatusTransition(from, to string) error {
	if from == to {
		return nil
	}
	allowed, ok := validProductStatusTransitions[from]
	if !ok {
		return fmt.Errorf("product: unknown status %q", from)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	return fmt.Errorf("product: invalid status transition from %q to %q", from, to)
}

// ProductEmbedding stores semantic-search vectors separately so baseline product migration works on plain Postgres.
type ProductEmbedding struct {
	ProductID string           `json:"productId" gorm:"primaryKey;type:varchar(255)"`
	Embedding *pgvector.Vector `json:"-" gorm:"type:vector(1536)"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Product   Product          `json:"-" gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:CASCADE"`
}

func (ProductEmbedding) TableName() string {
	return "product_embeddings"
}

// Category represents a product category
type Category struct {
	Slug         string    `json:"slug" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Thumbnail    string    `json:"thumbnail"`
	Icon         string    `json:"icon"`
	ProductCount int       `json:"productCount" gorm:"-"`
	Products     []Product `json:"products,omitempty" gorm:"foreignKey:CategorySlug"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Inquiry represents an inquiry submission
type Inquiry struct {
	ID                    string             `json:"id" gorm:"primaryKey"`
	UserID                *string            `json:"userId" gorm:"index"`
	CompanyName           string             `json:"companyName"`
	ContactPerson         string             `json:"contactPerson"`
	Email                 string             `json:"email"`
	WhatsApp              string             `json:"whatsapp"`
	TargetCountry         string             `json:"targetCountry"`
	EstimatedQuantity     string             `json:"estimatedQuantity"`
	InterestedProducts    common.StringArray `json:"interestedProducts" gorm:"type:jsonb"`
	PackagingRequirements string             `json:"packagingRequirements"`
	FlavorRequirements    string             `json:"flavorRequirements"`
	OEMNeeded             bool               `json:"oemNeeded"`
	ExpectedDelivery      string             `json:"expectedDelivery"`
	Message               string             `json:"message" gorm:"type:text"`
	Files                 common.StringArray `json:"files" gorm:"type:jsonb"`
	Status                string             `json:"status" gorm:"index;default:'pending'"` // pending, quoted, negotiating, won, lost
	AssignedTo            *string            `json:"assignedTo" gorm:"index"`
	Priority              string             `json:"priority" gorm:"default:'normal'"` // low, normal, high, urgent
	Products              common.StringArray `json:"products" gorm:"type:jsonb"`       // Store selected products/specs
	QuotedAmount          float64            `json:"quotedAmount"`
	QuotedAt              *time.Time         `json:"quotedAt"`
	ValidUntil            *time.Time         `json:"validUntil"`
	InternalNotes         string             `json:"internalNotes" gorm:"type:text"`
	CustomerNotes         string             `json:"customerNotes" gorm:"type:text"`
	// Incoterms 询盘阶段议定的贸易术语（如 CIF New York），映射至 Trade.Terms
	Incoterms string `json:"incoterms" gorm:"type:varchar(50)"`
	// NegotiatedPaymentTerms 议定付款方式（如 30% T/T 预付），写入 Trade.CommercialNotes
	NegotiatedPaymentTerms string    `json:"negotiatedPaymentTerms" gorm:"type:varchar(255)"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// Use shared common model types to avoid cross-package duplication and reduce cycle risk.
type OEMStep = common.OEMStep
type OEMStepArray = common.OEMStepArray
type Milestone = common.Milestone
type MilestoneArray = common.MilestoneArray
type PaginatedResponse = common.PaginatedResponse
type Pagination = common.Pagination
type ErrorResponse = common.ErrorResponse

// OEMFlow represents an OEM service flow
type OEMFlow struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Type        string       `json:"type" gorm:"index"` // quick_odm, full_oem
	Steps       OEMStepArray `json:"steps" gorm:"type:jsonb"`
	Timeline    string       `json:"timeline"`
	MOQ         int          `json:"moq"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// OEMSolution represents an OEM solution/service
type OEMSolution struct {
	ID           string             `json:"id" gorm:"primaryKey"`
	Slug         string             `json:"slug" gorm:"uniqueIndex"`
	Title        string             `json:"title"`
	Description  string             `json:"description" gorm:"type:text"`
	Thumbnail    string             `json:"thumbnail"`
	Images       common.StringArray `json:"images" gorm:"type:jsonb"`
	Category     string             `json:"category" gorm:"index"`
	MOQ          int                `json:"moq"`
	Applications common.StringArray `json:"applications" gorm:"type:jsonb"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// FactoryInfo represents factory information
type FactoryInfo struct {
	ID          string             `json:"id" gorm:"primaryKey"`
	Name        string             `json:"name"`
	Founded     int                `json:"founded"`
	Description string             `json:"description" gorm:"type:text"`
	Address     string             `json:"address"`
	Capacity    FactoryCapacity    `json:"capacity" gorm:"embedded"`
	Area        string             `json:"area"`
	Employees   int                `json:"employees"`
	Markets     common.StringArray `json:"markets" gorm:"type:jsonb"`
	Images      common.StringArray `json:"images" gorm:"type:jsonb"`
	Milestones  MilestoneArray     `json:"milestones" gorm:"type:jsonb"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// FactoryCapacity represents factory capacity info
type FactoryCapacity struct {
	Daily   string `json:"daily"`
	Monthly string `json:"monthly"`
}

// Certification represents a certification
type Certification struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name"`
	Abbreviation   string    `json:"abbreviation"`
	Description    string    `json:"description"`
	Issuer         string    `json:"issuer"`
	ValidUntil     string    `json:"validUntil"`
	CertificateURL string    `json:"certificateUrl"`
	BadgeURL       string    `json:"badgeUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ProcessControl represents a quality control process
type ProcessControl struct {
	ID          string             `json:"id" gorm:"primaryKey"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Stage       string             `json:"stage" gorm:"index"` // raw_material, production, finished_product
	Standards   common.StringArray `json:"standards" gorm:"type:jsonb"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// BlogPost represents a blog post
type BlogPost struct {
	ID           string             `json:"id" gorm:"primaryKey"`
	Slug         string             `json:"slug" gorm:"uniqueIndex"`
	Title        string             `json:"title"`
	Excerpt      string             `json:"excerpt"`
	Content      string             `json:"content" gorm:"type:text"`
	Category     string             `json:"category" gorm:"index"`
	AuthorName   string             `json:"-" gorm:"column:author_name"`
	AuthorAvatar string             `json:"-" gorm:"column:author_avatar"`
	AuthorTitle  string             `json:"-" gorm:"column:author_title"`
	AuthorBio    string             `json:"-" gorm:"column:author_bio;type:text"`
	Author       BlogAuthor         `json:"author" gorm:"-"`
	PublishedAt  time.Time          `json:"publishedAt"`
	Thumbnail    string             `json:"thumbnail"`
	ReadTime     int                `json:"readTime"` // in minutes
	Tags         common.StringArray `json:"tags" gorm:"type:jsonb"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// BlogAuthor represents blog post author for JSON response
type BlogAuthor struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
	Title  string `json:"title,omitempty"`
	Bio    string `json:"bio,omitempty"`
}

// AfterFind GORM hook to populate Author from flat fields
func (b *BlogPost) AfterFind(tx *gorm.DB) error {
	b.Author = BlogAuthor{
		Name:   b.AuthorName,
		Avatar: b.AuthorAvatar,
		Title:  b.AuthorTitle,
		Bio:    b.AuthorBio,
	}
	return nil
}

// Author represents blog post author (kept for backward compatibility)
type Author struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// GetAuthor returns the author as an Author struct
func (b *BlogPost) GetAuthor() Author {
	return Author{
		Name:   b.AuthorName,
		Avatar: b.AuthorAvatar,
	}
}

// CaseStudy represents a case study
type CaseStudy struct {
	ID        string             `json:"id" gorm:"primaryKey"`
	Slug      string             `json:"slug" gorm:"uniqueIndex"`
	Title     string             `json:"title"`
	Client    string             `json:"client"`
	Industry  string             `json:"industry" gorm:"index"`
	Location  string             `json:"location"`
	Thumbnail string             `json:"thumbnail"`
	Images    common.StringArray `json:"images" gorm:"type:jsonb"`
	Challenge string             `json:"challenge" gorm:"type:text"`
	Solution  string             `json:"solution" gorm:"type:text"`
	Result    string             `json:"result" gorm:"type:text"`
	Timeline  string             `json:"timeline"`
	Services  common.StringArray `json:"services" gorm:"type:jsonb"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Products []Product   `json:"products"`
	Posts    []BlogPost  `json:"posts"`
	Cases    []CaseStudy `json:"cases"`
}

// ProductVariant represents a specific SKU variant of a product (e.g., flavor + shape + weight).
type ProductVariant struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ProductID     string    `json:"productId" gorm:"not null;index"`
	SKU           string    `json:"sku" gorm:"uniqueIndex;not null"`
	Flavor        string    `json:"flavor" gorm:"type:varchar(100)"`
	Shape         string    `json:"shape" gorm:"type:varchar(100)"`
	Weight        string    `json:"weight" gorm:"type:varchar(50)"` // e.g. "500g", "1kg"
	Packaging     string    `json:"packaging" gorm:"type:varchar(100)"`
	StockQuantity int       `json:"stockQuantity" gorm:"default:0"`
	PriceModifier float64   `json:"priceModifier" gorm:"default:0"` // +/- adjustment from base price
	IsActive      bool      `json:"isActive" gorm:"default:true"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
