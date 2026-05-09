package publicscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	content "candypro/api/internal/services/content"
	inquiry "candypro/api/internal/services/inquiry"
	oem "candypro/api/internal/services/oem"
	product "candypro/api/internal/services/product"
	search "candypro/api/internal/services/search"
)

type Services struct {
	Product  *product.ProductService
	Category *product.CategoryService
	Inquiry  *inquiry.InquiryService
	OEM      *oem.OEMService
	Factory  *oem.FactoryService
	Content  *content.ContentService
	Search   *search.SearchService
}

func New(repos *repositoryCommon.PublicRepositories, cfg *config.Config) *Services {
	if repos == nil {
		return &Services{}
	}

	return &Services{
		Product:  product.NewProductService(repos.Product),
		Category: product.NewCategoryService(repos.Category),
		Inquiry:  inquiry.NewInquiryService(repos.Inquiry, cfg),
		OEM:      oem.NewOEMService(repos.OEM),
		Factory:  oem.NewFactoryService(repos.Factory),
		Content:  content.NewContentService(repos.Content),
		Search:   search.NewService(repos, cfg),
	}
}
