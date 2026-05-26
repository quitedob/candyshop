package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/sanitize"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type adminContentAuthorInput struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Title  string `json:"title"`
	Bio    string `json:"bio"`
}

type adminContentAuthorPatch struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
	Title  *string `json:"title"`
	Bio    *string `json:"bio"`
}

type adminContentCreateRequest struct {
	Type        string                   `json:"type"`
	ID          string                   `json:"id"`
	Slug        string                   `json:"slug"`
	Title       string                   `json:"title"`
	Excerpt     string                   `json:"excerpt"`
	Content     string                   `json:"content"`
	Category    string                   `json:"category"`
	Author      *adminContentAuthorInput `json:"author"`
	PublishedAt *time.Time               `json:"publishedAt"`
	Thumbnail   string                   `json:"thumbnail"`
	OgImage     string                   `json:"ogImage"`
	ReadTime    int                      `json:"readTime"`
	Tags        []string                 `json:"tags"`
	Translations *modelsCommon.JSONMap   `json:"translations"`

	Client    string   `json:"client"`
	Industry  string   `json:"industry"`
	Location  string   `json:"location"`
	Images    []string `json:"images"`
	Challenge string   `json:"challenge"`
	Solution  string   `json:"solution"`
	Result    string   `json:"result"`
	Timeline  string   `json:"timeline"`
	Services  []string `json:"services"`
}

type adminContentUpdateRequest struct {
	Type        string                   `json:"type"`
	Slug        *string                  `json:"slug"`
	Title       *string                  `json:"title"`
	Excerpt     *string                  `json:"excerpt"`
	Content     *string                  `json:"content"`
	Category    *string                  `json:"category"`
	Author      *adminContentAuthorPatch `json:"author"`
	PublishedAt *time.Time               `json:"publishedAt"`
	Thumbnail   *string                  `json:"thumbnail"`
	OgImage     *string                  `json:"ogImage"`
	ReadTime     *int                    `json:"readTime"`
	Tags         *[]string               `json:"tags"`
	Translations *modelsCommon.JSONMap   `json:"translations"`

	Client    *string   `json:"client"`
	Industry  *string   `json:"industry"`
	Location  *string   `json:"location"`
	Images    *[]string `json:"images"`
	Challenge *string   `json:"challenge"`
	Solution  *string   `json:"solution"`
	Result    *string   `json:"result"`
	Timeline  *string   `json:"timeline"`
	Services  *[]string `json:"services"`
}

// AdminGetContent lists CMS content for admin by type.
// @Summary Admin list content
// @Tags admin-content
// @Produce json
// @Router /admin/content [get]
func (h *Handler) AdminGetContent(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	contentType := normalizeContentType(c.Query("type"))

	if contentType == "" {
		contentType = "post"
	}

	switch contentType {
	case "post":
		category := strings.TrimSpace(c.Query("category"))
		posts, err := h.services.Content.GetPosts(c.Request.Context(), page, limit, category)
		if err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "posts_fetch_failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type":       "post",
			"data":       posts.Data,
			"pagination": posts.Pagination,
		})
	case "case":
		industry := strings.TrimSpace(c.Query("industry"))
		cases, err := h.services.Content.GetCases(c.Request.Context(), page, limit, industry)
		if err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "cases_fetch_failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type":       "case",
			"data":       cases.Data,
			"pagination": cases.Pagination,
		})
	}
}

// AdminGetContentByID gets CMS content detail by id.
// @Summary Admin get content detail
// @Tags admin-content
// @Produce json
// @Param id path string true "Content ID"
// @Router /admin/content/{id} [get]
func (h *Handler) AdminGetContentByID(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	contentType := normalizeContentType(c.Query("type"))

	if contentType == "post" {
		post, err := h.services.Content.GetPostByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "content_not_found")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"type":    "post",
			"content": post,
		})
		return
	}

	if contentType == "case" {
		caseStudy, err := h.services.Content.GetCaseByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "content_not_found")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"type":    "case",
			"content": caseStudy,
		})
		return
	}

	post, postErr := h.services.Content.GetPostByID(c.Request.Context(), id)
	if postErr == nil {
		c.JSON(http.StatusOK, gin.H{
			"type":    "post",
			"content": post,
		})
		return
	}

	caseStudy, caseErr := h.services.Content.GetCaseByID(c.Request.Context(), id)
	if caseErr == nil {
		c.JSON(http.StatusOK, gin.H{
			"type":    "case",
			"content": caseStudy,
		})
		return
	}

	response.ErrorResp(c, http.StatusNotFound, "content_not_found")
}

// AdminCreateContent creates CMS content (blog/case study)
// @Summary Admin create content
// @Tags admin-content
// @Produce json
// @Router /admin/content [post]
func (h *Handler) AdminCreateContent(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminContentCreateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	contentType := normalizeContentType(req.Type)
	if contentType == "" {
		response.InvalidResp(c, "content_type_invalid")
		return
	}

	switch contentType {
	case "post":
		post := buildPostFromCreateRequest(req)
		syncPostScalarsFromLocale(post, i18n.DefaultLocale())
		if strings.TrimSpace(post.Title) == "" {
			response.InvalidResp(c, "content_title_required")
			return
		}

		if err := h.services.Content.CreatePost(c.Request.Context(), post); err != nil {
			if dberror.IsDuplicateKeyError(err) {
				response.ErrorResp(c, http.StatusConflict, "content_slug_conflict")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "content_create_failed")
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Content created successfully",
			"type":    "post",
			"content": post,
		})
	case "case":
		caseStudy := buildCaseFromCreateRequest(req)
		syncCaseScalarsFromLocale(caseStudy, i18n.DefaultLocale())
		if strings.TrimSpace(caseStudy.Title) == "" {
			response.InvalidResp(c, "content_title_required")
			return
		}

		if err := h.services.Content.CreateCase(c.Request.Context(), caseStudy); err != nil {
			if dberror.IsDuplicateKeyError(err) {
				response.ErrorResp(c, http.StatusConflict, "content_slug_conflict")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "content_create_failed")
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Content created successfully",
			"type":    "case",
			"content": caseStudy,
		})
	}
}

// AdminUpdateContent updates CMS content
// @Summary Admin update content
// @Tags admin-content
// @Produce json
// @Param id path string true "Content ID"
// @Router /admin/content/{id} [put]
func (h *Handler) AdminUpdateContent(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	var req adminContentUpdateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	contentType := normalizeContentType(req.Type)
	if contentType == "" {
		response.InvalidResp(c, "content_type_invalid")
		return
	}

	switch contentType {
	case "post":
		post, err := h.services.Content.GetPostByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "content_not_found")
			return
		}

		applyPostPatch(post, req)
		syncPostScalarsFromLocale(post, i18n.DefaultLocale())
		if strings.TrimSpace(post.Title) == "" {
			response.InvalidResp(c, "content_title_empty")
			return
		}

		post.UpdatedAt = time.Now()
		if err := h.services.Content.UpdatePost(c.Request.Context(), post); err != nil {
			if dberror.IsDuplicateKeyError(err) {
				response.ErrorResp(c, http.StatusConflict, "content_slug_conflict")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "content_update_failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Content updated successfully",
			"type":    "post",
			"content": post,
		})
	case "case":
		caseStudy, err := h.services.Content.GetCaseByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "content_not_found")
			return
		}

		applyCasePatch(caseStudy, req)
		syncCaseScalarsFromLocale(caseStudy, i18n.DefaultLocale())
		if strings.TrimSpace(caseStudy.Title) == "" {
			response.InvalidResp(c, "content_title_empty")
			return
		}

		caseStudy.UpdatedAt = time.Now()
		if err := h.services.Content.UpdateCase(c.Request.Context(), caseStudy); err != nil {
			if dberror.IsDuplicateKeyError(err) {
				response.ErrorResp(c, http.StatusConflict, "content_slug_conflict")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "content_update_failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Content updated successfully",
			"type":    "case",
			"content": caseStudy,
		})
	}
}

// AdminDeleteContent deletes CMS content
// @Summary Admin delete content
// @Tags admin-content
// @Produce json
// @Param id path string true "Content ID"
// @Router /admin/content/{id} [delete]
func (h *Handler) AdminDeleteContent(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	contentType := normalizeContentType(c.Query("type"))

	switch contentType {
	case "post":
		if err := h.services.Content.DeletePost(c.Request.Context(), id); err != nil {
			handleDeleteContentError(c, err, "post")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Content deleted successfully",
			"type":    "post",
			"id":      id,
		})
		return
	case "case":
		if err := h.services.Content.DeleteCase(c.Request.Context(), id); err != nil {
			handleDeleteContentError(c, err, "case")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Content deleted successfully",
			"type":    "case",
			"id":      id,
		})
		return
	}

	postErr := h.services.Content.DeletePost(c.Request.Context(), id)
	if postErr == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Content deleted successfully",
			"type":    "post",
			"id":      id,
		})
		return
	}
	if !errors.Is(postErr, gorm.ErrRecordNotFound) {
		handleDeleteContentError(c, postErr, "post")
		return
	}

	caseErr := h.services.Content.DeleteCase(c.Request.Context(), id)
	if caseErr == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Content deleted successfully",
			"type":    "case",
			"id":      id,
		})
		return
	}
	if errors.Is(caseErr, gorm.ErrRecordNotFound) {
		response.ErrorResp(c, http.StatusNotFound, "content_not_found")
		return
	}
	handleDeleteContentError(c, caseErr, "case")
}

func buildPostFromCreateRequest(req adminContentCreateRequest) *modelsProduct.BlogPost {
	now := time.Now()
	title := strings.TrimSpace(req.Title)
	slug := normalizeSlug(req.Slug)
	if slug == "" {
		slug = buildProductSlug(title)
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = crypto.GenerateID()
	}

	post := &modelsProduct.BlogPost{
		ID:          id,
		Slug:        slug,
		Title:       title,
		Excerpt:     strings.TrimSpace(req.Excerpt),
		Content:     sanitize.HTML(strings.TrimSpace(req.Content)),
		Category:    strings.TrimSpace(req.Category),
		Thumbnail:   strings.TrimSpace(req.Thumbnail),
		OgImage:     strings.TrimSpace(req.OgImage),
		ReadTime:    req.ReadTime,
		Tags:        modelsCommon.StringArray(req.Tags),
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.PublishedAt != nil && !req.PublishedAt.IsZero() {
		post.PublishedAt = *req.PublishedAt
	}
	if req.Author != nil {
		post.AuthorName = strings.TrimSpace(req.Author.Name)
		post.AuthorAvatar = strings.TrimSpace(req.Author.Avatar)
		post.AuthorTitle = strings.TrimSpace(req.Author.Title)
		post.AuthorBio = strings.TrimSpace(req.Author.Bio)
	}
	if req.Translations != nil {
		post.Translations = *req.Translations
	}

	return post
}

func buildCaseFromCreateRequest(req adminContentCreateRequest) *modelsProduct.CaseStudy {
	now := time.Now()
	title := strings.TrimSpace(req.Title)
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = crypto.GenerateID()
	}

	slug := normalizeSlug(req.Slug)
	if slug == "" {
		base := title
		if base == "" {
			base = req.Client
		}
		slug = buildProductSlug(base)
	}

	cs := &modelsProduct.CaseStudy{
		ID:        id,
		Slug:      slug,
		Title:     title,
		Client:    strings.TrimSpace(req.Client),
		Industry:  strings.TrimSpace(req.Industry),
		Location:  strings.TrimSpace(req.Location),
		Thumbnail: strings.TrimSpace(req.Thumbnail),
		OgImage:   strings.TrimSpace(req.OgImage),
		Images:    modelsCommon.StringArray(req.Images),
		Challenge: sanitize.HTML(strings.TrimSpace(req.Challenge)),
		Solution:  sanitize.HTML(strings.TrimSpace(req.Solution)),
		Result:    sanitize.HTML(strings.TrimSpace(req.Result)),
		Timeline:  strings.TrimSpace(req.Timeline),
		Services:  modelsCommon.StringArray(req.Services),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Translations != nil {
		cs.Translations = *req.Translations
	}
	return cs
}

func applyPostPatch(post *modelsProduct.BlogPost, req adminContentUpdateRequest) {
	if req.Slug != nil {
		post.Slug = normalizeSlug(*req.Slug)
	}
	if req.Title != nil {
		post.Title = strings.TrimSpace(*req.Title)
	}
	if req.Excerpt != nil {
		post.Excerpt = strings.TrimSpace(*req.Excerpt)
	}
	if req.Content != nil {
		post.Content = sanitize.HTML(strings.TrimSpace(*req.Content))
	}
	if req.Category != nil {
		post.Category = strings.TrimSpace(*req.Category)
	}
	if req.Thumbnail != nil {
		post.Thumbnail = strings.TrimSpace(*req.Thumbnail)
	}
	if req.OgImage != nil {
		post.OgImage = strings.TrimSpace(*req.OgImage)
	}
	if req.ReadTime != nil {
		post.ReadTime = *req.ReadTime
	}
	if req.Tags != nil {
		post.Tags = modelsCommon.StringArray(*req.Tags)
	}
	if req.Translations != nil {
		post.Translations = *req.Translations
	}
	if req.PublishedAt != nil && !req.PublishedAt.IsZero() {
		post.PublishedAt = *req.PublishedAt
	}
	if req.Author != nil {
		if req.Author.Name != nil {
			post.AuthorName = strings.TrimSpace(*req.Author.Name)
		}
		if req.Author.Avatar != nil {
			post.AuthorAvatar = strings.TrimSpace(*req.Author.Avatar)
		}
		if req.Author.Title != nil {
			post.AuthorTitle = strings.TrimSpace(*req.Author.Title)
		}
		if req.Author.Bio != nil {
			post.AuthorBio = strings.TrimSpace(*req.Author.Bio)
		}
	}
	if post.Slug == "" && post.Title != "" {
		post.Slug = buildProductSlug(post.Title)
	}
}

func applyCasePatch(caseStudy *modelsProduct.CaseStudy, req adminContentUpdateRequest) {
	if req.Slug != nil {
		caseStudy.Slug = normalizeSlug(*req.Slug)
	}
	if req.Title != nil {
		caseStudy.Title = strings.TrimSpace(*req.Title)
	}
	if req.Client != nil {
		caseStudy.Client = strings.TrimSpace(*req.Client)
	}
	if req.Industry != nil {
		caseStudy.Industry = strings.TrimSpace(*req.Industry)
	}
	if req.Location != nil {
		caseStudy.Location = strings.TrimSpace(*req.Location)
	}
	if req.Thumbnail != nil {
		caseStudy.Thumbnail = strings.TrimSpace(*req.Thumbnail)
	}
	if req.OgImage != nil {
		caseStudy.OgImage = strings.TrimSpace(*req.OgImage)
	}
	if req.Images != nil {
		caseStudy.Images = modelsCommon.StringArray(*req.Images)
	}
	if req.Challenge != nil {
		caseStudy.Challenge = sanitize.HTML(strings.TrimSpace(*req.Challenge))
	}
	if req.Solution != nil {
		caseStudy.Solution = sanitize.HTML(strings.TrimSpace(*req.Solution))
	}
	if req.Result != nil {
		caseStudy.Result = sanitize.HTML(strings.TrimSpace(*req.Result))
	}
	if req.Timeline != nil {
		caseStudy.Timeline = strings.TrimSpace(*req.Timeline)
	}
	if req.Services != nil {
		caseStudy.Services = modelsCommon.StringArray(*req.Services)
	}
	if req.Translations != nil {
		caseStudy.Translations = *req.Translations
	}
	if caseStudy.Slug == "" {
		base := caseStudy.Title
		if base == "" {
			base = caseStudy.Client
		}
		caseStudy.Slug = buildProductSlug(base)
	}
}

func normalizeContentType(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "post", "blog", "blog_post", "blogpost":
		return "post"
	case "case", "case_study", "casestudy":
		return "case"
	default:
		return ""
	}
}

func handleDeleteContentError(c *gin.Context, err error, contentType string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.ErrorResp(c, http.StatusNotFound, "content_not_found")
		return
	}
	response.ErrorResp(c, http.StatusInternalServerError, "content_delete_failed")
}

// syncPostScalarsFromLocale 将博客翻译字段同步到标量列。
func syncPostScalarsFromLocale(post *modelsProduct.BlogPost, locale string) {
	if post.Translations == nil {
		return
	}
	fields, ok := post.Translations[locale]
	if !ok {
		return
	}
	if v := strings.TrimSpace(fields["title"]); v != "" {
		post.Title = v
	}
	if v := strings.TrimSpace(fields["excerpt"]); v != "" {
		post.Excerpt = v
	}
	if v := strings.TrimSpace(fields["content"]); v != "" {
		post.Content = sanitize.HTML(v)
	}
	if v := strings.TrimSpace(fields["authorName"]); v != "" {
		post.AuthorName = v
	}
	if v := strings.TrimSpace(fields["authorTitle"]); v != "" {
		post.AuthorTitle = v
	}
	if v := strings.TrimSpace(fields["authorBio"]); v != "" {
		post.AuthorBio = v
	}
	if post.Slug == "" && post.Title != "" {
		post.Slug = buildProductSlug(post.Title)
	}
}

// syncCaseScalarsFromLocale 将案例翻译字段同步到标量列。
func syncCaseScalarsFromLocale(cs *modelsProduct.CaseStudy, locale string) {
	if cs.Translations == nil {
		return
	}
	fields, ok := cs.Translations[locale]
	if !ok {
		return
	}
	if v := strings.TrimSpace(fields["title"]); v != "" {
		cs.Title = v
	}
	if v := strings.TrimSpace(fields["challenge"]); v != "" {
		cs.Challenge = sanitize.HTML(v)
	}
	if v := strings.TrimSpace(fields["solution"]); v != "" {
		cs.Solution = sanitize.HTML(v)
	}
	if v := strings.TrimSpace(fields["result"]); v != "" {
		cs.Result = sanitize.HTML(v)
	}
	if cs.Slug == "" {
		base := cs.Title
		if base == "" {
			base = cs.Client
		}
		if base != "" {
			cs.Slug = buildProductSlug(base)
		}
	}
}
