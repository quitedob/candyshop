// Package sanitize provides HTML sanitization for user/AI-generated content.
// Uses bluemonday with an allowlist that mirrors the frontend DOMPurify config.
package sanitize

import "github.com/microcosm-cc/bluemonday"

var policy *bluemonday.Policy

func init() {
	policy = bluemonday.NewPolicy()

	// Inline text semantics
	policy.AllowElements(
		"strong", "b", "em", "i", "u", "s", "del", "ins",
		"sub", "sup", "small", "mark",
		"br", "hr",
	)

	// Headings
	policy.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")

	// Paragraph and block elements
	policy.AllowElements("p", "div", "span", "blockquote", "pre", "code")

	// Lists
	policy.AllowElements("ul", "ol", "li")

	// Links
	policy.AllowAttrs("href", "target", "rel").OnElements("a")
	policy.RequireNoFollowOnLinks(false)
	policy.AllowStandardURLs()
	policy.AllowURLSchemes("http", "https", "mailto", "tel")

	// Images
	policy.AllowAttrs("src", "alt", "width", "height", "loading").OnElements("img")

	// Tables
	policy.AllowAttrs("colspan", "rowspan").OnElements("td", "th")
	policy.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td")

	// Global attributes
	policy.AllowAttrs("class", "id").Globally()
	policy.AllowAttrs("start", "type").OnElements("ol")

	// No inline styles, no data-* attributes, no event handlers, no scripts
}

// HTML sanitizes raw HTML, removing scripts, event handlers, javascript: URLs,
// and any elements/attributes not in the allowlist.
func HTML(raw string) string {
	if raw == "" {
		return ""
	}
	return policy.Sanitize(raw)
}
