// Package admin — One-click XLSX translation using AI.
// Upload a spreadsheet, select target language, download translated copy.
// Formatting and numeric cells are preserved. Only text cells are translated.
package admin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// AdminTranslateXLSX accepts an uploaded XLSX file, translates all text cells
// to the target language via AI, and returns the translated file.
//
// POST /admin/xlsx/translate
//
// Form fields:
//   - file: the XLSX file (multipart)
//   - targetLang: ISO language code (zh, en, ar, ja, es, fr, de, etc.)
//   - sourceLang: optional source language (auto-detect if empty)
func (h *Handler) AdminTranslateXLSX(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	targetLang := strings.TrimSpace(c.PostForm("targetLang"))
	if targetLang == "" {
		response.InvalidResp(c, "target_lang_required")
		return
	}
	sourceLang := strings.TrimSpace(c.PostForm("sourceLang"))

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidResp(c, "xlsx_file_required")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_read_failed")
		return
	}

	// Open the uploaded spreadsheet
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "xlsx_parse_failed")
		return
	}

	// Iterate all sheets and translate text cells
	sheets := f.GetSheetList()
	totalCells := 0
	translatedCells := 0

	for _, sheet := range sheets {
		rows, rowsErr := f.GetRows(sheet)
		if rowsErr != nil {
			continue
		}

		for rowIdx, row := range rows {
			for colIdx, cellValue := range row {
				totalCells++
				trimmed := strings.TrimSpace(cellValue)
				if trimmed == "" || isNumericOrDate(trimmed) {
					continue
				}
				if len(trimmed) < 2 {
					continue // skip single chars
				}

				translated, transErr := translateXLSXCell(c, h, trimmed, sourceLang, targetLang)
				if transErr != nil || translated == "" || translated == trimmed {
					continue
				}

				cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
				if setErr := f.SetCellValue(sheet, cellRef, translated); setErr != nil {
					continue
				}
				translatedCells++
			}
		}
	}

	// Write translated file
	var outBuf bytes.Buffer
	if writeErr := f.Write(&outBuf); writeErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_write_failed")
		return
	}

	filename := fmt.Sprintf("translated_%s_%s.xlsx", targetLang, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", outBuf.Bytes())
}

// AdminBatchTranslateXLSX translates all cells in ALL sheets to ALL requested languages
// and returns a single XLSX with a sheet per language.
//
// POST /admin/xlsx/translate-batch
func (h *Handler) AdminBatchTranslateXLSX(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	targetLangsStr := strings.TrimSpace(c.PostForm("targetLangs")) // comma-separated: en,zh,ar,ja
	if targetLangsStr == "" {
		response.InvalidResp(c, "target_langs_required")
		return
	}
	targetLangs := parseCommaList(targetLangsStr)
	if len(targetLangs) == 0 {
		response.InvalidResp(c, "target_langs_empty")
		return
	}

	sourceLang := strings.TrimSpace(c.PostForm("sourceLang"))

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidResp(c, "xlsx_file_required")
		return
	}
	defer file.Close()

	fileBytes, _ := io.ReadAll(file)

	// Open source once, extract text cells
	srcFile, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "xlsx_parse_failed")
		return
	}

	// Collect all text cells
	type cellPos struct {
		sheet string
		ref   string
		value string
	}
	var cells []cellPos
	for _, sheet := range srcFile.GetSheetList() {
		rows, _ := srcFile.GetRows(sheet)
		for r, row := range rows {
			for c, val := range row {
				v := strings.TrimSpace(val)
				if v == "" || isNumericOrDate(v) || len(v) < 2 {
					continue
				}
				ref, _ := excelize.CoordinatesToCellName(c+1, r+1)
				cells = append(cells, cellPos{sheet: sheet, ref: ref, value: v})
			}
		}
	}

	// For each target language, create a translated sheet
	outFile := excelize.NewFile()
	outFile.SetSheetName("Sheet1", targetLangs[0])

	type langResult struct {
		lang  string
		cells map[string]string // ref → translated text
	}
	var mu sync.Mutex
	var results []langResult

	for _, lang := range targetLangs {
		result := langResult{lang: lang, cells: make(map[string]string, len(cells))}
		for _, cell := range cells {
			translated, tErr := translateXLSXCell(c, h, cell.value, sourceLang, lang)
			if tErr == nil && translated != "" && translated != cell.value {
				result.cells[cell.ref] = translated
			}
		}
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	// Build sheets: one per language
	for i, res := range results {
		sheetName := res.lang
		if i > 0 {
			outFile.NewSheet(sheetName)
		}
		// Copy source structure, replace translated cells
		for _, sheet := range srcFile.GetSheetList() {
			rows, _ := srcFile.GetRows(sheet)
			for r, row := range rows {
				for c, val := range row {
					ref, _ := excelize.CoordinatesToCellName(c+1, r+1)
					if translated, ok := res.cells[ref]; ok {
						outFile.SetCellValue(sheetName, ref, translated)
					} else {
						outFile.SetCellValue(sheetName, ref, val)
					}
				}
			}
		}
	}

	var outBuf bytes.Buffer
	outFile.Write(&outBuf)

	filename := fmt.Sprintf("translated_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", outBuf.Bytes())
}

// ── helpers ──

func translateXLSXCell(c *gin.Context, h *Handler, text, sourceLang, targetLang string) (string, error) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		return "", fmt.Errorf("ai not configured")
	}

	translationPrompt := fmt.Sprintf(
		"Translate the following text from %s to %s. Return ONLY the translated text, no explanations, no quotation marks, no additional formatting:\n\n%s",
		sourceLangOrDefault(sourceLang), targetLang, text,
	)

	return h.aiService.Generate(c.Request.Context(), translationPrompt)
}

func sourceLangOrDefault(lang string) string {
	if lang == "" {
		return "auto"
	}
	return lang
}

func isNumericOrDate(s string) bool {
	s = strings.TrimSpace(s)
	// Common numeric patterns
	var num float64
	if _, err := fmt.Sscanf(s, "%f", &num); err == nil {
		return true
	}
	// Date patterns
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return true
	}
	if _, err := time.Parse("01/02/2006", s); err == nil {
		return true
	}
	if _, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return true
	}
	// Currency patterns
	if strings.HasPrefix(s, "$") || strings.HasPrefix(s, "¥") || strings.HasPrefix(s, "€") {
		return true
	}
	return false
}

func parseCommaList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
