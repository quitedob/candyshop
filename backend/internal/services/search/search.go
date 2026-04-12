package search

import (
	"context"

	modelsProduct "candypro/api/internal/models/product"
	repositoryCommon "candypro/api/internal/repository/common"
)

// SearchService handles search business logic.
type SearchService struct {
	repos *repositoryCommon.PublicRepositories
}

// NewService creates a new SearchService.
func NewService(repos *repositoryCommon.PublicRepositories) *SearchService {
	return &SearchService{repos: repos}
}

// Search searches across all content types.
func (s *SearchService) Search(ctx context.Context, query string, searchType string, limit int) (*modelsProduct.SearchResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	response := &modelsProduct.SearchResponse{}

	// Search products
	if searchType == "" || searchType == "all" || searchType == "products" {
		products, err := s.repos.Product.Search(ctx, query, limit)
		if err == nil {
			response.Products = products
		}
	}

	// Search posts
	if searchType == "" || searchType == "all" || searchType == "posts" {
		posts, err := s.repos.Content.SearchPosts(ctx, query, limit)
		if err == nil {
			response.Posts = posts
		}
	}

	// Search cases
	if searchType == "" || searchType == "all" || searchType == "cases" {
		cases, err := s.repos.Content.SearchCases(ctx, query, limit)
		if err == nil {
			response.Cases = cases
		}
	}

	return response, nil
}
