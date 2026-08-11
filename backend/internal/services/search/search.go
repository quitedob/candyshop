package search

import (
	"context"
	"log"
	"strings"
	"sync"

	"candypro/api/internal/config"
	"candypro/api/internal/database"
	modelsProduct "candypro/api/internal/models/product"
	repositoryCommon "candypro/api/internal/repository/common"
)

// embeddingBackfillBatchSize bounds how many active products are embedded per
// backfill pass so a cold start never issues an unbounded burst of embeddings
// API calls inside a search request.
const embeddingBackfillBatchSize = 20

// SearchService handles search business logic.
type SearchService struct {
	repos *repositoryCommon.PublicRepositories
	cfg   *config.Config
	// backfillMu serializes the product_embeddings backfill so concurrent search
	// requests cannot stampede the embeddings API (G24b).
	backfillMu sync.Mutex
	// dimWarnOnce logs the "embedding model dimension != vector(1536)" warning
	// exactly once per process so a misconfigured model is loud and diagnosable
	// instead of silently failing every write and leaving semantic search inert
	// (G24b).
	dimWarnOnce sync.Once
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

	// Ensure the vector table is populated before querying it (G24b).
	s.ensureProductEmbeddings(ctx)

	emb, err := openAIEmbed(ctx, s.cfg.AI.OpenAIAPIKey, s.cfg.AI.OpenAIEmbeddingModel, query)
	if err != nil {
		log.Printf("Note: product semantic embedding failed, falling back to keyword search: %v", err)
		return nil
	}
	if len(emb) != modelsProduct.EmbeddingDim {
		s.warnEmbeddingDimension(emb)
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

// ensureProductEmbeddings lazily backfills product_embeddings so the pgvector
// semantic path actually has vectors to query (G24b). product_embeddings was
// previously never written, so SearchVectorSimilar's JOIN returned nothing and
// semantic search silently fell back to keyword on every request. Only products
// without an embedding row are processed, so after the first pass this becomes a
// cheap NOT-IN scan that picks up newly created products. Embedding API failures
// are logged, never propagated — keyword search remains the fallback. TryLock
// lets concurrent search requests proceed (with keyword fallback) instead of
// serializing behind a slow cold-start backfill.
func (s *SearchService) ensureProductEmbeddings(ctx context.Context) {
	if !s.semanticSearchEnabled() {
		return
	}
	if !s.backfillMu.TryLock() {
		return
	}
	defer s.backfillMu.Unlock()

	missing, err := s.repos.Product.FindProductsMissingEmbeddings(ctx, embeddingBackfillBatchSize)
	if err != nil {
		log.Printf("Note: product embedding backfill lookup failed: %v", err)
		return
	}
	if len(missing) == 0 {
		return
	}
	for _, p := range missing {
		text := strings.TrimSpace(p.Name)
		if summary := strings.TrimSpace(p.Summary); summary != "" {
			text += "\n" + summary
		}
		if text == "" {
			continue
		}
		emb, err := openAIEmbed(ctx, s.cfg.AI.OpenAIAPIKey, s.cfg.AI.OpenAIEmbeddingModel, text)
		if err != nil {
			log.Printf("Note: product embedding generation failed for %q (%s): %v", p.Name, p.ID, err)
			continue
		}
		if len(emb) != modelsProduct.EmbeddingDim {
			// vector(1536) can never store it — warning once, then skip the write
			// rather than failing it per product against the DB (G24b).
			s.warnEmbeddingDimension(emb)
			continue
		}
		if err := s.repos.Product.UpsertEmbedding(ctx, p.ID, emb); err != nil {
			log.Printf("Note: product embedding write failed for %s: %v", p.ID, err)
		}
	}
}

// warnEmbeddingDimension logs a once-per-process warning when the configured
// embedding model's output dimension does not match the vector(1536) column, so
// the "semantic search is inert" symptom is diagnosable instead of silent (G24b).
func (s *SearchService) warnEmbeddingDimension(emb []float32) {
	s.dimWarnOnce.Do(func() {
		model := ""
		if s.cfg != nil {
			model = s.cfg.AI.OpenAIEmbeddingModel
		}
		log.Printf("Warning: embedding model %q returned %d dimensions, want %d (vector(%d)); semantic search is disabled and falling back to keyword. Configure an embedding model with %d-dim output (G24b).",
			model, len(emb), modelsProduct.EmbeddingDim, modelsProduct.EmbeddingDim, modelsProduct.EmbeddingDim)
	})
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
