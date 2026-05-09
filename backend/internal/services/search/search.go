package search

import (
	"context"
	"log"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/database"
	modelsProduct "candypro/api/internal/models/product"
	repositoryCommon "candypro/api/internal/repository/common"
)

// SearchService handles search business logic.
type SearchService struct {
	repos *repositoryCommon.PublicRepositories
	cfg   *config.Config
}

// NewService creates a new SearchService.
func NewService(repos *repositoryCommon.PublicRepositories, cfg *config.Config) *SearchService {
	return &SearchService{repos: repos, cfg: cfg}
}

// Search searches across all content types.
func (s *SearchService) Search(ctx context.Context, query string, searchType string, limit int) (*modelsProduct.SearchResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	response := &modelsProduct.SearchResponse{}

	// Search products（pgvector 语义 + ILIKE 混合）
	if searchType == "" || searchType == "all" || searchType == "products" {
		limitKw := limit
		if limitKw < 1 {
			limitKw = 10
		}

		semantic := s.semanticProductSearch(ctx, query, limitKw)
		keyword, err := s.repos.Product.Search(ctx, query, limitKw)
		if err != nil {
			keyword = nil
		}
		response.Products = mergeProductResults(semantic, keyword, limitKw)
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

func (s *SearchService) semanticProductSearch(ctx context.Context, query string, limit int) []modelsProduct.Product {
	if !s.semanticSearchEnabled() || strings.TrimSpace(query) == "" || limit <= 0 {
		return nil
	}

	emb, err := openAIEmbed(ctx, s.cfg.AI.OpenAIAPIKey, s.cfg.AI.OpenAIEmbeddingModel, query)
	if err != nil {
		log.Printf("Note: product semantic embedding failed, falling back to keyword search: %v", err)
		return nil
	}
	if len(emb) == 0 {
		return nil
	}

	products, err := s.repos.Product.SearchVectorSimilar(ctx, emb, limit)
	if err != nil {
		log.Printf("Note: product semantic search failed, falling back to keyword search: %v", err)
		return nil
	}
	return products
}

func (s *SearchService) semanticSearchEnabled() bool {
	if s.cfg == nil {
		return false
	}
	if s.cfg.AI.IsSemanticSearchDisabled() {
		return false
	}
	if strings.TrimSpace(s.cfg.AI.OpenAIAPIKey) == "" {
		return false
	}
	if !database.PgvectorAvailable() {
		return false
	}
	return true
}
