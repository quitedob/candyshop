// Package admin — One-click XLSX translation using AI.
package admin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type xlsxCell struct {
	key   string // Sheet!A1
	sheet string
	ref   string
	value string
}

// AdminTranslateXLSX accepts an uploaded XLSX file, translates all text cells
// to the target language via AI, and returns the translated file.
//
// POST /admin/xlsx/translate
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

	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "xlsx_parse_failed")
		return
	}
	defer func() { _ = f.Close() }()

	cells := collectTranslatableCells(f)
	translated, warnings, transErr := h.translateXlsxCells(c, cells, sourceLang, targetLang)
	if transErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	applied := applyXlsxTranslations(f, cells, translated)

	outBuf, writeErr := writeXlsxToBuffer(f)
	if writeErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_write_failed")
		return
	}

	filename := fmt.Sprintf("translated_%s_%s.xlsx", targetLang, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Translated-Cells", fmt.Sprintf("%d", applied))
	c.Header("X-Translation-Warnings", fmt.Sprintf("%d", len(warnings)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", outBuf)
}

// AdminBatchTranslateXLSX translates text cells to multiple languages;
// output workbook uses sheets named {originalSheet}_{lang}.
//
// POST /admin/xlsx/translate-batch
func (h *Handler) AdminBatchTranslateXLSX(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	targetLangsStr := strings.TrimSpace(c.PostForm("targetLangs"))
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_read_failed")
		return
	}

	srcFile, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "xlsx_parse_failed")
		return
	}
	defer func() { _ = srcFile.Close() }()

	cells := collectTranslatableCells(srcFile)
	outFile := excelize.NewFile()
	defaultSheet := outFile.GetSheetName(0)
	firstOut := true
	totalApplied := 0
	totalWarnings := 0

	for _, lang := range targetLangs {
		translated, warnings, transErr := h.translateXlsxCells(c, cells, sourceLang, lang)
		if transErr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
			return
		}
		totalWarnings += len(warnings)

		for _, srcSheet := range srcFile.GetSheetList() {
			outSheet := sanitizeSheetName(fmt.Sprintf("%s_%s", srcSheet, lang))
			if firstOut {
				outFile.SetSheetName(defaultSheet, outSheet)
				firstOut = false
			} else if idx, _ := outFile.GetSheetIndex(outSheet); idx == -1 {
				_, _ = outFile.NewSheet(outSheet)
			}

			rows, _ := srcFile.GetRows(srcSheet)
			for r, row := range rows {
				for c, val := range row {
					ref, _ := excelize.CoordinatesToCellName(c+1, r+1)
					key := xlsxCellKey(srcSheet, ref)
					if tr, ok := translated[key]; ok && tr != "" {
						_ = outFile.SetCellValue(outSheet, ref, tr)
						totalApplied++
					} else {
						_ = outFile.SetCellValue(outSheet, ref, val)
					}
				}
			}
		}
	}

	outBuf, writeErr := writeXlsxToBuffer(outFile)
	if writeErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_write_failed")
		return
	}

	filename := fmt.Sprintf("translated_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Translated-Cells", fmt.Sprintf("%d", totalApplied))
	c.Header("X-Translation-Warnings", fmt.Sprintf("%d", totalWarnings))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", outBuf)
}

func (h *Handler) translateXlsxCells(c *gin.Context, cells []xlsxCell, sourceLang, targetLang string) (map[string]string, []string, error) {
	items := make(map[string]string, len(cells))
	for _, cell := range cells {
		items[cell.key] = cell.value
	}
	return h.aiService.BatchTranslateTexts(c.Request.Context(), items, sourceLang, targetLang)
}

func collectTranslatableCells(f *excelize.File) []xlsxCell {
	var cells []xlsxCell
	for _, sheet := range f.GetSheetList() {
		rows, rowsErr := f.GetRows(sheet)
		if rowsErr != nil {
			continue
		}
		for rowIdx, row := range rows {
			for colIdx, cellValue := range row {
				trimmed := strings.TrimSpace(cellValue)
				if trimmed == "" || isNumericOrDate(trimmed) || len(trimmed) < 2 {
					continue
				}
				ref, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
				cells = append(cells, xlsxCell{
					key:   xlsxCellKey(sheet, ref),
					sheet: sheet,
					ref:   ref,
					value: trimmed,
				})
			}
		}
	}
	return cells
}

func applyXlsxTranslations(f *excelize.File, cells []xlsxCell, translated map[string]string) int {
	applied := 0
	for _, cell := range cells {
		tr, ok := translated[cell.key]
		if !ok || tr == "" || tr == cell.value {
			continue
		}
		if setErr := f.SetCellValue(cell.sheet, cell.ref, tr); setErr == nil {
			applied++
		}
	}
	return applied
}

func writeXlsxToBuffer(f *excelize.File) ([]byte, error) {
	var outBuf bytes.Buffer
	if err := f.Write(&outBuf); err != nil {
		return nil, err
	}
	return outBuf.Bytes(), nil
}

func xlsxCellKey(sheet, ref string) string {
	return sheet + "!" + ref
}

func sanitizeSheetName(name string) string {
	// Excel sheet name max 31 chars, no : \ / ? * [ ]
	replacer := strings.NewReplacer(":", "_", "\\", "_", "/", "_", "?", "_", "*", "_", "[", "_", "]", "_")
	name = replacer.Replace(name)
	if len(name) > 31 {
		name = name[:31]
	}
	if name == "" {
		return "Sheet1"
	}
	return name
}

func isNumericOrDate(s string) bool {
	s = strings.TrimSpace(s)
	var num float64
	if _, err := fmt.Sscanf(s, "%f", &num); err == nil {
		return true
	}
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return true
	}
	if _, err := time.Parse("01/02/2006", s); err == nil {
		return true
	}
	if _, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return true
	}
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
