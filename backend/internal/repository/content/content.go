package content

import (
	"context"
	"strings"

	modelsProduct "candypro/api/internal/models/product"

	"gorm.io/gorm"
)

// ContentRepository handles blog and case study data operations
type ContentRepository struct {
	db *gorm.DB
}

// NewContentRepository creates a new ContentRepository
func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

// ===== Blog Posts =====

// FindAllPosts returns paginated blog posts
func (r *ContentRepository) FindAllPosts(ctx context.Context, page, limit int, category string) ([]modelsProduct.BlogPost, int64, error) {
	var posts []modelsProduct.BlogPost
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.BlogPost{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Scopes(r.withCategoryFilter(category)).
		Offset(offset).Limit(limit).Order("published_at DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *ContentRepository) withCategoryFilter(category string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if category != "" {
			return db.Where("category = ?", category)
		}
		return db
	}
}

// FindPostBySlug returns a blog post by slug
func (r *ContentRepository) FindPostBySlug(ctx context.Context, slug string) (*modelsProduct.BlogPost, error) {
	var post modelsProduct.BlogPost
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// FindPostByID returns a blog post by ID
func (r *ContentRepository) FindPostByID(ctx context.Context, id string) (*modelsProduct.BlogPost, error) {
	var post modelsProduct.BlogPost
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// FindRelatedPosts returns related posts by category
func (r *ContentRepository) FindRelatedPosts(ctx context.Context, slug string, limit int) ([]modelsProduct.BlogPost, error) {
	var post modelsProduct.BlogPost
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&post).Error; err != nil {
		return nil, err
	}

	var posts []modelsProduct.BlogPost
	if err := r.db.WithContext(ctx).Where("category = ? AND slug != ?", post.Category, slug).
		Limit(limit).Order("published_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// SearchPosts searches blog posts by query
func (r *ContentRepository) SearchPosts(ctx context.Context, query string, limit int) ([]modelsProduct.BlogPost, error) {
	var posts []modelsProduct.BlogPost
	searchPattern := "%" + escapeLikePattern(query) + "%"
	if err := r.db.WithContext(ctx).Where("title ILIKE ? OR excerpt ILIKE ? OR content ILIKE ?",
		searchPattern, searchPattern, searchPattern).
		Limit(limit).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// CreatePost creates a new blog post
func (r *ContentRepository) CreatePost(ctx context.Context, post *modelsProduct.BlogPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// UpdatePost updates a blog post
func (r *ContentRepository) UpdatePost(ctx context.Context, post *modelsProduct.BlogPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// DeletePost deletes a blog post
func (r *ContentRepository) DeletePost(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&modelsProduct.BlogPost{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ===== Case Studies =====

// FindAllCases returns paginated case studies
func (r *ContentRepository) FindAllCases(ctx context.Context, page, limit int, industry string) ([]modelsProduct.CaseStudy, int64, error) {
	var cases []modelsProduct.CaseStudy
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.CaseStudy{})
	if industry != "" {
		query = query.Where("industry = ?", industry)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Scopes(r.withIndustryFilter(industry)).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

func (r *ContentRepository) withIndustryFilter(industry string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if industry != "" {
			return db.Where("industry = ?", industry)
		}
		return db
	}
}

// FindCaseBySlug returns a case study by slug
func (r *ContentRepository) FindCaseBySlug(ctx context.Context, slug string) (*modelsProduct.CaseStudy, error) {
	var caseStudy modelsProduct.CaseStudy
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&caseStudy).Error; err != nil {
		return nil, err
	}
	return &caseStudy, nil
}

// FindCaseByID returns a case study by ID
func (r *ContentRepository) FindCaseByID(ctx context.Context, id string) (*modelsProduct.CaseStudy, error) {
	var caseStudy modelsProduct.CaseStudy
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&caseStudy).Error; err != nil {
		return nil, err
	}
	return &caseStudy, nil
}

// SearchCases searches case studies by query
func (r *ContentRepository) SearchCases(ctx context.Context, query string, limit int) ([]modelsProduct.CaseStudy, error) {
	var cases []modelsProduct.CaseStudy
	searchPattern := "%" + escapeLikePattern(query) + "%"
	if err := r.db.WithContext(ctx).Where("title ILIKE ? OR client ILIKE ? OR challenge ILIKE ? OR solution ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(limit).Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}

// CreateCase creates a new case study
func (r *ContentRepository) CreateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error {
	return r.db.WithContext(ctx).Create(caseStudy).Error
}

// UpdateCase updates a case study
func (r *ContentRepository) UpdateCase(ctx context.Context, caseStudy *modelsProduct.CaseStudy) error {
	return r.db.WithContext(ctx).Save(caseStudy).Error
}

// DeleteCase deletes a case study
func (r *ContentRepository) DeleteCase(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&modelsProduct.CaseStudy{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// escapeLikePattern escapes special characters for SQL LIKE operator
func escapeLikePattern(pattern string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(pattern)
}
