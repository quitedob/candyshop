package content

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type contentRepository interface {
	FindAllPosts(ctx context.Context, page, limit int, category string) ([]modelsProduct.BlogPost, int64, error)
	FindPostBySlug(ctx context.Context, slug string) (*modelsProduct.BlogPost, error)
	FindPostByID(ctx context.Context, id string) (*modelsProduct.BlogPost, error)
	FindRelatedPosts(ctx context.Context, slug string, limit int) ([]modelsProduct.BlogPost, error)
	CreatePost(ctx context.Context, post *modelsProduct.BlogPost) error
	UpdatePost(ctx context.Context, post *modelsProduct.BlogPost) error
	DeletePost(ctx context.Context, id string) error
	FindAllCases(ctx context.Context, page, limit int, industry string) ([]modelsProduct.CaseStudy, int64, error)
	FindCaseBySlug(ctx context.Context, slug string) (*modelsProduct.CaseStudy, error)
	FindCaseByID(ctx context.Context, id string) (*modelsProduct.CaseStudy, error)
	FindRelatedCases(ctx context.Context, slug string, limit int) ([]modelsProduct.CaseStudy, error)
	CreateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error
	UpdateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error
	DeleteCase(ctx context.Context, id string) error
}

// ContentService handles blog and case study business logic.
type ContentService struct {
	repo contentRepository
}

// NewContentService creates a new ContentService.
func NewContentService(repo contentRepository) *ContentService {
	return &ContentService{repo: repo}
}

// GetPosts returns paginated blog posts.
func (s *ContentService) GetPosts(ctx context.Context, page, limit int, category string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	posts, total, err := s.repo.FindAllPosts(ctx, page, limit, category)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: posts,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetPostBySlug returns a blog post by slug.
func (s *ContentService) GetPostBySlug(ctx context.Context, slug string) (*modelsProduct.BlogPost, error) {
	return s.repo.FindPostBySlug(ctx, slug)
}

// GetPostByID returns a blog post by id.
func (s *ContentService) GetPostByID(ctx context.Context, id string) (*modelsProduct.BlogPost, error) {
	return s.repo.FindPostByID(ctx, id)
}

// GetRelatedPosts returns related posts.
func (s *ContentService) GetRelatedPosts(ctx context.Context, slug string, limit int) ([]modelsProduct.BlogPost, error) {
	if limit <= 0 {
		limit = 3
	}
	return s.repo.FindRelatedPosts(ctx, slug, limit)
}

// CreatePost creates a new blog post.
func (s *ContentService) CreatePost(ctx context.Context, post *modelsProduct.BlogPost) error {
	return s.repo.CreatePost(ctx, post)
}

// UpdatePost updates a blog post.
func (s *ContentService) UpdatePost(ctx context.Context, post *modelsProduct.BlogPost) error {
	return s.repo.UpdatePost(ctx, post)
}

// DeletePost deletes a blog post.
func (s *ContentService) DeletePost(ctx context.Context, id string) error {
	return s.repo.DeletePost(ctx, id)
}

// GetCases returns paginated case studies.
func (s *ContentService) GetCases(ctx context.Context, page, limit int, industry string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	cases, total, err := s.repo.FindAllCases(ctx, page, limit, industry)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: cases,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetCaseBySlug returns a case study by slug.
func (s *ContentService) GetCaseBySlug(ctx context.Context, slug string) (*modelsProduct.CaseStudy, error) {
	return s.repo.FindCaseBySlug(ctx, slug)
}

// GetCaseByID returns a case study by id.
func (s *ContentService) GetCaseByID(ctx context.Context, id string) (*modelsProduct.CaseStudy, error) {
	return s.repo.FindCaseByID(ctx, id)
}

// GetRelatedCases returns related case studies by industry.
func (s *ContentService) GetRelatedCases(ctx context.Context, slug string, limit int) ([]modelsProduct.CaseStudy, error) {
	if limit <= 0 {
		limit = 3
	}
	return s.repo.FindRelatedCases(ctx, slug, limit)
}

// CreateCase creates a new case study.
func (s *ContentService) CreateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error {
	return s.repo.CreateCase(ctx, caseStudy)
}

// UpdateCase updates a case study.
func (s *ContentService) UpdateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error {
	return s.repo.UpdateCase(ctx, caseStudy)
}

// DeleteCase deletes a case study.
func (s *ContentService) DeleteCase(ctx context.Context, id string) error {
	return s.repo.DeleteCase(ctx, id)
}
