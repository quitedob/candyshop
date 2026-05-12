// Package docxgen generates DOCX documents for B2B trade documents (contracts, invoices)
// using Go standard library only. A DOCX is a ZIP archive containing XML files per the OOXML standard.
package docxgen

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"time"
)

// xmlHeader is the standard XML declaration.
const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`

// contentTypes returns the [Content_Types].xml for a minimal DOCX.
func contentTypes() string {
	return xmlHeader + `
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
}

// rels returns the .rels file.
func rels() string {
	return xmlHeader + `
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
}

// wordRels returns the word/_rels/document.xml.rels file.
func wordRels() string {
	return xmlHeader + `
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
}

// styles returns a basic styles.xml with default body and table styles.
func styles() string {
	return xmlHeader + `
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault><w:rPr><w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr></w:rPrDefault>
  </w:docDefaults>
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/><w:pPr><w:spacing w:after="120" w:line="276" w:lineRule="auto"/></w:pPr></w:style>
  <w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:before="240" w:after="120"/></w:pPr><w:rPr><w:b/><w:sz w:val="36"/><w:szCs w:val="36"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:before="240" w:after="120"/></w:pPr><w:rPr><w:b/><w:sz w:val="28"/><w:szCs w:val="28"/></w:rPr></w:style>
  <w:style w:type="paragraph" w:styleId="Subtitle"><w:name w:val="Subtitle"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/></w:pPr><w:rPr><w:i/><w:sz w:val="20"/><w:szCs w:val="20"/><w:color w:val="666666"/></w:rPr></w:style>
  <w:style w:type="table" w:styleId="TableGrid"><w:name w:val="Table Grid"/><w:basedOn w:val="TableNormal"/><w:tblPr><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/><w:left w:val="single" w:sz="4" w:space="0" w:color="000000"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/><w:right w:val="single" w:sz="4" w:space="0" w:color="000000"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="000000"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="000000"/></w:tblBorders></w:tblPr></w:style>
</w:styles>`
}

// GenerateDocx creates a .docx file from the given document body XML content.
// bodyXML should contain the contents of <w:body>...</w:body> (without the body wrapper itself).
func GenerateDocx(bodyXML string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	files := map[string]string{
		"[Content_Types].xml":      contentTypes(),
		"_rels/.rels":              rels(),
		"word/_rels/document.xml.rels": wordRels(),
		"word/styles.xml":          styles(),
		"word/document.xml": documentBody(bodyXML),
		"docProps/app.xml":         appProps(),
		"docProps/core.xml":        coreProps(),
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, fmt.Errorf("create %s in zip: %w", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("write %s: %w", name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return buf.Bytes(), nil
}

func documentBody(bodyXML string) string {
	return xmlHeader + `
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>` + bodyXML + `
  </w:body>
</w:document>`
}

func appProps() string {
	return xmlHeader + `
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>CandyPro OEM</Application>
  <Company>CandyPro Manufacturing</Company>
</Properties>`
}

func coreProps() string {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	return xmlHeader + `
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/">
  <dc:creator>CandyPro OEM</dc:creator>
  <dcterms:created>` + now + `</dcterms:created>
  <dcterms:modified>` + now + `</dcterms:modified>
</cp:coreProperties>`
}

// --- XML builders ---

// Paragraph creates a <w:p> with a single run.
func Paragraph(text string, opts ...ParOption) string {
	p := &parProps{}
	for _, opt := range opts {
		opt(p)
	}

	run := runProps{fontSize: p.fontSize, bold: p.bold, italic: p.italic, color: p.color}
	rPr := runPropsXML(run)

	var sb strings.Builder
	sb.WriteString("<w:p>")
	if p.align != "" {
		sb.WriteString(fmt.Sprintf(`<w:pPr><w:jc w:val="%s"/></w:pPr>`, p.align))
	}
	sb.WriteString(fmt.Sprintf(`<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>`, rPr, escapeXML(text)))
	sb.WriteString("</w:p>")
	return sb.String()
}

type parProps struct {
	align    string
	fontSize string
	bold     bool
	italic   bool
	color    string
}

// ParOption configures paragraph formatting.
type ParOption func(*parProps)

// ParCenter centers the paragraph.
func ParCenter() ParOption { return func(p *parProps) { p.align = "center" } }

// ParRight right-aligns the paragraph.
func ParRight() ParOption { return func(p *parProps) { p.align = "right" } }

// ParBold makes the text bold.
func ParBold() ParOption { return func(p *parProps) { p.bold = true } }

// ParItalic makes the text italic.
func ParItalic() ParOption { return func(p *parProps) { p.italic = true } }

// ParFontSize sets the font size in half-points (e.g., 22 = 11pt).
func ParFontSize(halfPt int) ParOption { return func(p *parProps) { p.fontSize = fmt.Sprintf("%d", halfPt) } }

// ParColor sets the text color (hex like "FF0000").
// The hex value is validated for XML safety (alphanumeric only).
func ParColor(hex string) ParOption {
	return func(p *parProps) { p.color = sanitizeColor(hex) }
}

// sanitizeColor ensures a color value contains only valid hex characters.
// Returns "000000" if the value is empty or contains invalid characters.
func sanitizeColor(s string) string {
	if len(s) == 0 {
		return "000000"
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return "000000"
		}
	}
	return s
}

type runProps struct {
	fontSize string
	bold     bool
	italic   bool
	color    string
}

func runPropsXML(r runProps) string {
	var sb strings.Builder
	sb.WriteString("<w:rPr>")
	if r.fontSize != "" {
		sb.WriteString(fmt.Sprintf("<w:sz w:val=\"%s\"/><w:szCs w:val=\"%s\"/>", r.fontSize, r.fontSize))
	}
	if r.bold {
		sb.WriteString("<w:b/>")
	}
	if r.italic {
		sb.WriteString("<w:i/>")
	}
	if r.color != "" {
		sb.WriteString(fmt.Sprintf("<w:color w:val=\"%s\"/>", r.color))
	}
	sb.WriteString("</w:rPr>")
	return sb.String()
}

// Table creates a simple <w:tbl> with header and data rows.
// headers are the column headers; rows are the data rows (each row is a slice of cell texts).
func Table(headers []string, rows [][]string, opts ...TableOption) string {
	t := &tblProps{colWidths: make([]string, len(headers))}
	for _, opt := range opts {
		opt(t)
	}

	var sb strings.Builder
	sb.WriteString(`<w:tbl><w:tblPr>`)
	// Set table width to 100%
	sb.WriteString(`<w:tblW w:w="5000" w:type="pct"/>`)
	sb.WriteString(`<w:tblStyle w:val="TableGrid"/>`)
	sb.WriteString(`</w:tblPr>`)

	// Header row
	sb.WriteString("<w:tr>")
	for i, h := range headers {
		sb.WriteString("<w:tc>")
		if t.colWidths[i] != "" {
			sb.WriteString(fmt.Sprintf(`<w:tcPr><w:tcW w:w="%s" w:type="pct"/></w:tcPr>`, t.colWidths[i]))
		}
		sb.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(h)))
		sb.WriteString("</w:tc>")
	}
	sb.WriteString("</w:tr>")

	// Data rows
	for _, row := range rows {
		sb.WriteString("<w:tr>")
		for i, cell := range row {
			if i >= len(headers) {
				break
			}
			sb.WriteString("<w:tc>")
			sb.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(cell)))
			sb.WriteString("</w:tc>")
		}
		sb.WriteString("</w:tr>")
	}

	sb.WriteString("</w:tbl>")
	return sb.String()
}

type tblProps struct {
	colWidths []string
}

// TableOption configures table formatting.
type TableOption func(*tblProps)

// ColWidths sets relative column widths (percentage strings like "20", "30", "50").
// Each width is validated to contain only digits.
func ColWidths(widths ...string) TableOption {
	return func(t *tblProps) {
		t.colWidths = make([]string, len(widths))
		for i, w := range widths {
			t.colWidths[i] = sanitizeNumeric(w)
		}
	}
}

// sanitizeNumeric ensures a string contains only digit characters.
// Returns "0" if the value is empty or contains non-digits.
func sanitizeNumeric(s string) string {
	if len(s) == 0 {
		return "0"
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "0"
		}
	}
	return s
}

// Spacer creates an empty paragraph (vertical space).
func Spacer() string {
	return `<w:p><w:r><w:br/></w:r></w:p>`
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// FormatMoney formats a float64 as "USD 1,234.56".
func FormatMoney(amt float64, currency string) string {
	intPart := int(amt)
	decPart := int((amt - float64(intPart)) * 100)
	if decPart < 0 {
		decPart = -decPart
	}
	return fmt.Sprintf("%s %s.%02d", currency, commaFormat(intPart), decPart)
}

func commaFormat(n int) string {
	if n < 0 {
		return "-" + commaFormat(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	return commaFormat(n/1000) + "," + fmt.Sprintf("%03d", n%1000)
}

// DateStr formats a time as "Jan 2, 2006".
func DateStr(t time.Time) string {
	return t.Format("Jan 2, 2006")
}
