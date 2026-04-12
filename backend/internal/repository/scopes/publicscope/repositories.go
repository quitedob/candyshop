package publicscope

import (
	content "candypro/api/internal/repository/content"
	inquiry "candypro/api/internal/repository/inquiry"
	oem "candypro/api/internal/repository/oem"
	product "candypro/api/internal/repository/product"

	"gorm.io/gorm"
)

type Repositories struct {
	Product  *product.ProductRepository
	Category *product.CategoryRepository
	OEM      *oem.OEMRepository
	Factory  *oem.FactoryRepository
	Content  *content.ContentRepository
	Inquiry  *inquiry.InquiryRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Product:  product.NewProductRepository(db),
		Category: product.NewCategoryRepository(db),
		OEM:      oem.NewOEMRepository(db),
		Factory:  oem.NewFactoryRepository(db),
		Content:  content.NewContentRepository(db),
		Inquiry:  inquiry.NewInquiryRepository(db),
	}
}
