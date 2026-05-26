package admin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ── Export Selected Inventory ──

// AdminExportInventoryXLSX exports selected products as XLSX.
// POST /admin/inventory/export-xlsx
func (h *Handler) AdminExportInventoryXLSX(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	currency, convert := h.resolveCurrency(c)
	f := excelize.NewFile()
	sheet := "Inventory"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Product ID", "Name", "Category", "Category Slug",
		fmt.Sprintf("Base Price (%s)", currency), "MOQ", "Stock", "Lead Time",
		"Halal", "OEM", "HS Code", "Shelf Life", "Storage", "Status",
		"Thumbnail", "Images", "Flavors", "Shapes", "Ingredients", "Allergens",
		"Certifications", "Description", "Created At"}
	f.SetSheetRow(sheet, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	rowIdx := 2
	for _, id := range req.IDs {
		product, err := h.services.Product.GetProductByID(c.Request.Context(), id)
		if err != nil {
			continue
		}
		halal, oem := "No", "No"
		if product.HalalCertified {
			halal = "Yes"
		}
		if product.OEMAvailable {
			oem = "Yes"
		}
		price := formatMoney(convert(product.BasePrice), currency)
		rowData := []string{
			product.ID, product.Name, product.Category, product.CategorySlug,
			price, strconv.Itoa(product.MOQ), strconv.Itoa(product.StockQuantity),
			product.LeadTime, halal, oem, product.HSCode, product.ShelfLife,
			product.Storage, product.Status, product.Thumbnail,
			strings.Join(product.Images, ", "),
			strings.Join(product.Flavors, ", "),
			strings.Join(product.Shapes, ", "),
			product.Ingredients, product.Allergens,
			strings.Join(product.Certifications, ", "),
			product.Description, product.CreatedAt.Format("2006-01-02"),
		}
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", rowIdx), &rowData)
		f.SetRowStyle(sheet, rowIdx, rowIdx, dataStyle)
		rowIdx++
	}

	for col := 'A'; col <= 'V'; col++ {
		_ = f.SetColWidth(sheet, string(col), string(col), 18)
	}

	serveXLSXFile(c, f, fmt.Sprintf("candypro_inventory_%s.xlsx", time.Now().Format("20060102")))
}

// ── Import XLSX Preview ──

type xlsxImportPreview struct {
	Headers          []string                    `json:"headers"`
	Rows             [][]string                  `json:"rows"`
	TotalRows        int                         `json:"totalRows"`
	ColumnMapping    map[string]string           `json:"columnMapping"`
	ImageColumns     []int                       `json:"imageColumns"`
	AIMapping        map[string]string           `json:"aiMapping,omitempty"`
}

var knownProductFields = map[string]string{
	"product id":    "id",
	"id":            "id",
	"name":          "name",
	"product name":  "name",
	"category":      "category",
	"category slug": "categorySlug",
	"cat slug":      "categorySlug",
	"price":         "basePrice",
	"base price":    "basePrice",
	"unit price":    "basePrice",
	"单价":            "basePrice",
	"价格":            "basePrice",
	"moq":           "moq",
	"min order":     "moq",
	"起订量":           "moq",
	"stock":         "stockQuantity",
	"quantity":      "stockQuantity",
	"qty":           "stockQuantity",
	"库存":            "stockQuantity",
	"数量":            "stockQuantity",
	"lead time":     "leadTime",
	"leadtime":      "leadTime",
	"交货期":           "leadTime",
	"halal":         "halalCertified",
	"halal certified": "halalCertified",
	"清真":            "halalCertified",
	"oem":           "oemAvailable",
	"oem available": "oemAvailable",
	"可定制":           "oemAvailable",
	"hs code":       "hsCode",
	"hscode":        "hsCode",
	"海关编码":          "hsCode",
	"shelf life":    "shelfLife",
	"shelflife":     "shelfLife",
	"保质期":           "shelfLife",
	"storage":       "storage",
	"存储":            "storage",
	"status":        "status",
	"状态":            "status",
	"thumbnail":     "thumbnail",
	"image":         "thumbnail",
	"图片":            "thumbnail",
	"主图":            "thumbnail",
	"images":        "images",
	"图片列表":          "images",
	"flavors":       "flavors",
	"flavours":      "flavors",
	"口味":            "flavors",
	"shapes":        "shapes",
	"形状":            "shapes",
	"ingredients":   "ingredients",
	"成分":            "ingredients",
	"allergens":     "allergens",
	"过敏原":           "allergens",
	"certifications": "certifications",
	"认证":            "certifications",
	"description":   "description",
	"描述":            "description",
	"summary":       "summary",
	"简介":            "summary",
}

var imageURLPattern = regexp.MustCompile(`(?i)^https?://.*\.(jpg|jpeg|png|gif|webp|svg|bmp)(\?.*)?$`)

// detectImageColumns checks each column's sample values for image URLs.
func detectImageColumns(rows [][]string, numCols int) []int {
	if len(rows) == 0 {
		return nil
	}
	imageCols := make(map[int]int) // col index → match count
	sampleSize := len(rows)
	if sampleSize > 10 {
		sampleSize = 10
	}
	for _, row := range rows[:sampleSize] {
		for colIdx := 0; colIdx < numCols && colIdx < len(row); colIdx++ {
			if imageURLPattern.MatchString(strings.TrimSpace(row[colIdx])) {
				imageCols[colIdx]++
			}
		}
	}
	var result []int
	for colIdx, count := range imageCols {
		if count > 0 {
			result = append(result, colIdx)
		}
	}
	return result
}

// mapColumnsHeuristic builds a column→field mapping using header name matching.
func mapColumnsHeuristic(headers []string) map[string]string {
	mapping := make(map[string]string)
	for i, h := range headers {
		normalized := strings.TrimSpace(strings.ToLower(h))
		if field, ok := knownProductFields[normalized]; ok {
			mapping[field] = strconv.Itoa(i)
		}
	}
	return mapping
}

// AdminImportInventoryXLSX parses an uploaded XLSX and returns a preview.
// POST /admin/inventory/import-xlsx
func (h *Handler) AdminImportInventoryXLSX(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidResp(c, "xlsx_import_no_file")
		return
	}
	defer file.Close()

	// R2 C-4: cap upload size and validate file type before buffering the whole
	// payload into memory. Previously a multi-GB non-XLSX upload would be fully
	// read by io.ReadAll and only rejected at excelize.OpenReader, exhausting
	// server memory in the meantime.
	const maxXLSXBytes = 25 * 1024 * 1024 // 25 MiB
	if header.Size > 0 && header.Size > maxXLSXBytes {
		response.ErrorResp(c, http.StatusRequestEntityTooLarge, "xlsx_import_too_large")
		return
	}
	// Reject by extension first (cheap), then by Content-Type if the client set
	// it. Real validation happens when excelize.OpenReader sees the magic
	// bytes; these checks short-circuit obvious abuse before we buffer.
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		response.InvalidResp(c, "xlsx_import_bad_extension")
		return
	}
	if ct := header.Header.Get("Content-Type"); ct != "" {
		// XLSX is a ZIP container — common Content-Types include the official
		// OOXML mime, the generic ZIP types, and octet-stream from CLI uploads.
		switch ct {
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-excel",
			"application/zip",
			"application/x-zip-compressed",
			"application/octet-stream":
		default:
			response.InvalidResp(c, "xlsx_import_bad_content_type")
			return
		}
	}

	// LimitReader caps the buffer size even when the multipart Content-Length
	// claimed a smaller size (defence against truncated/malicious headers).
	fileBytes, err := io.ReadAll(io.LimitReader(file, maxXLSXBytes+1))
	if err != nil {
		response.InvalidResp(c, "xlsx_import_read_failed")
		return
	}
	if int64(len(fileBytes)) > maxXLSXBytes {
		response.ErrorResp(c, http.StatusRequestEntityTooLarge, "xlsx_import_too_large")
		return
	}

	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		response.InvalidResp(c, "xlsx_import_parse_failed")
		return
	}

	sheet := f.GetSheetName(0)
	allRows, err := f.GetRows(sheet)
	if err != nil || len(allRows) < 2 {
		response.InvalidResp(c, "xlsx_import_empty")
		return
	}

	headers := allRows[0]
	var dataRows [][]string
	for i := 1; i < len(allRows); i++ {
		row := allRows[i]
		// Pad row to match header length
		for len(row) < len(headers) {
			row = append(row, "")
		}
		// Trim to header length
		if len(row) > len(headers) {
			row = row[:len(headers)]
		}
		// Skip completely empty rows
		hasContent := false
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				hasContent = true
				break
			}
		}
		if hasContent {
			dataRows = append(dataRows, row)
		}
	}

	heuristicMapping := mapColumnsHeuristic(headers)
	imageColumns := detectImageColumns(dataRows, len(headers))

	preview := xlsxImportPreview{
		Headers:       headers,
		Rows:          dataRows,
		TotalRows:     len(dataRows),
		ColumnMapping: heuristicMapping,
		ImageColumns:  imageColumns,
	}

		// AI-assisted mapping
		if h.aiService != nil && h.aiService.IsEnabled() {
			aiMapping := h.aiService.BuildColumnMapping(c.Request.Context(), headers, dataRows)
			if aiMapping != nil {
				resolved := make(map[string]string)
				for field, colName := range aiMapping {
					for i, hdr := range headers {
						if strings.EqualFold(strings.TrimSpace(hdr), strings.TrimSpace(colName)) {
							resolved[field] = strconv.Itoa(i)
							break
						}
					}
				}
				preview.AIMapping = resolved
			}
		}
	c.JSON(http.StatusOK, preview)
}


// ── Apply Import ──

type xlsxApplyRequest struct {
	Rows          []map[string]interface{} `json:"rows"`
	ImageColumns  []string                 `json:"imageColumns"`
}

// AdminApplyInventoryImport creates/updates products from imported data.
// POST /admin/inventory/import-xlsx/apply
func (h *Handler) AdminApplyInventoryImport(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req xlsxApplyRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if len(req.Rows) == 0 {
		response.InvalidResp(c, "xlsx_import_no_rows")
		return
	}

	// Build image column set
	imageCols := make(map[string]bool)
	for _, col := range req.ImageColumns {
		imageCols[col] = true
	}

	operatorID := adminActorID(c)
	var created, updated, errors int
	var errorMessages []string

	for _, row := range req.Rows {
		// Handle image URLs: fetch and re-upload
		for colName := range imageCols {
			if val, ok := row[colName]; ok {
				if urlStr, isStr := val.(string); isStr && imageURLPattern.MatchString(urlStr) {
			uploadedURL, err := fetchAndUploadImage(c.Request.Context(), h.storage, urlStr)
			if err == nil {
				row[colName] = uploadedURL
			}
			}
			}
		}

		var product modelsProduct.Product
		productMap := make(map[string]interface{})

		// Extract known fields
		if v, ok := row["id"]; ok {
			if s, isStr := v.(string); isStr && s != "" {
				product.ID = s
			}
		}
		if v, ok := row["name"]; ok {
			if s, isStr := v.(string); isStr {
				product.Name = strings.TrimSpace(s)
			}
		}
		if v, ok := row["category"]; ok {
			if s, isStr := v.(string); isStr {
				product.Category = strings.TrimSpace(s)
			}
		}
		if v, ok := row["categorySlug"]; ok {
			if s, isStr := v.(string); isStr {
				product.CategorySlug = strings.TrimSpace(s)
			}
		}
		if v, ok := row["basePrice"]; ok {
			product.BasePrice = toFloat64(v)
		}
		if v, ok := row["moq"]; ok {
			product.MOQ = toInt(v)
		}
		if v, ok := row["stockQuantity"]; ok {
			product.StockQuantity = toInt(v)
		}
		if v, ok := row["leadTime"]; ok {
			if s, isStr := v.(string); isStr {
				product.LeadTime = strings.TrimSpace(s)
			}
		}
		if v, ok := row["halalCertified"]; ok {
			product.HalalCertified = toBool(v)
		}
		if v, ok := row["oemAvailable"]; ok {
			product.OEMAvailable = toBool(v)
		}
		if v, ok := row["hsCode"]; ok {
			if s, isStr := v.(string); isStr {
				product.HSCode = strings.TrimSpace(s)
			}
		}
		if v, ok := row["shelfLife"]; ok {
			if s, isStr := v.(string); isStr {
				product.ShelfLife = strings.TrimSpace(s)
			}
		}
		if v, ok := row["storage"]; ok {
			if s, isStr := v.(string); isStr {
				product.Storage = strings.TrimSpace(s)
			}
		}
		if v, ok := row["status"]; ok {
			if s, isStr := v.(string); isStr {
				s = strings.TrimSpace(s)
				if modelsProduct.IsValidProductStatus(s) {
			product.Status = s
			}
			}
		}
		if v, ok := row["thumbnail"]; ok {
			if s, isStr := v.(string); isStr {
				product.Thumbnail = strings.TrimSpace(s)
			}
		}
		if v, ok := row["images"]; ok {
			product.Images = toStringArray(v)
		}
		if v, ok := row["flavors"]; ok {
			product.Flavors = toStringArray(v)
		}
		if v, ok := row["shapes"]; ok {
			product.Shapes = toStringArray(v)
		}
		if v, ok := row["ingredients"]; ok {
			if s, isStr := v.(string); isStr {
				product.Ingredients = strings.TrimSpace(s)
			}
		}
		if v, ok := row["allergens"]; ok {
			if s, isStr := v.(string); isStr {
				product.Allergens = strings.TrimSpace(s)
			}
		}
		if v, ok := row["certifications"]; ok {
			product.Certifications = toStringArray(v)
		}
		if v, ok := row["description"]; ok {
			if s, isStr := v.(string); isStr {
				product.Description = strings.TrimSpace(s)
			}
		}
		if v, ok := row["summary"]; ok {
			if s, isStr := v.(string); isStr {
				product.Summary = strings.TrimSpace(s)
			}
		}

		// Validation
		if product.Name == "" {
			errors++
			errorMessages = append(errorMessages, "Row missing name")
			continue
		}

		// ID handling
		isUpdate := false
		if strings.TrimSpace(product.ID) != "" {
			existing, err := h.services.Product.GetProductByID(c.Request.Context(), product.ID)
			if err == nil && existing != nil {
				isUpdate = true
			}
		}

		if isUpdate {
			existing, _ := h.services.Product.GetProductByID(c.Request.Context(), product.ID)
			product.CreatedAt = existing.CreatedAt
			product.CreatedBy = existing.CreatedBy
			product.Slug = existing.Slug
			product.ViewCount = existing.ViewCount
			product.UpdatedAt = time.Now()
			product.UpdatedBy = &operatorID
			if err := h.services.Product.UpdateProduct(c.Request.Context(), &product); err != nil {
				errors++
				errorMessages = append(errorMessages, fmt.Sprintf("Update %s failed: %v", product.Name, err))
				continue
			}
			updated++
		} else {
			product.ID = crypto.GenerateID()
			product.Slug = buildProductSlug(product.Name)
			if strings.TrimSpace(product.CategorySlug) == "" {
				product.CategorySlug = normalizeSlug(product.Category)
			}
			if strings.TrimSpace(product.Status) == "" {
				product.Status = modelsProduct.ProductStatusActive
			}
			product.CreatedBy = &operatorID
			product.UpdatedBy = &operatorID
			product.CreatedAt = time.Now()
			product.UpdatedAt = time.Now()
			if err := h.services.Product.CreateProduct(c.Request.Context(), &product); err != nil {
				errors++
				errorMessages = append(errorMessages, fmt.Sprintf("Create %s failed: %v", product.Name, err))
				continue
			}
			created++
		}
		_ = productMap // Reserved for future extended field support
	}

	c.JSON(http.StatusOK, gin.H{
		"created":       created,
		"updated":       updated,
		"errors":        errors,
		"errorMessages": errorMessages,
	})
}

// ── Helpers ──

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f
		}
	}
	return 0
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return 0
		}
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		s := strings.TrimSpace(strings.ToLower(val))
		return s == "yes" || s == "true" || s == "y" || s == "1" || s == "是"
	case float64:
		return val != 0
	case int:
		return val != 0
	}
	return false
}

func toStringArray(v interface{}) []string {
	switch val := v.(type) {
	case string:
		if val == "" {
			return nil
		}
		parts := strings.Split(val, ",")
		var result []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	case []interface{}:
		var result []string
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(s))
			}
		}
		return result
	}
	return nil
}

func fetchAndUploadImage(ctx context.Context, st storage.StorageService, imageURL string) (string, error) {
	if st == nil {
		return imageURL, fmt.Errorf("storage not available")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return imageURL, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return imageURL, fmt.Errorf("fetch image returned %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return imageURL, fmt.Errorf("not an image: %s", contentType)
	}

	var buf bytes.Buffer
	limited := io.LimitReader(resp.Body, 5*1024*1024) // 5MB max
	if _, err := io.Copy(&buf, limited); err != nil {
		return imageURL, err
	}

	filename := "imported_image"
	if idx := strings.LastIndex(imageURL, "/"); idx >= 0 {
		parts := strings.SplitN(imageURL[idx+1:], "?", 2)
		filename = parts[0]
	}
	if !strings.Contains(filename, ".") {
		filename += ".jpg"
	}

	uploadedURL, err := st.Upload(ctx, io.NopCloser(&buf), storage.UploadOptions{
		Folder:      "products",
		FileName:    filename,
		MaxFileSize: 5 * 1024 * 1024,
	})
	if err != nil {
		return imageURL, err
	}

	return uploadedURL, nil
}
