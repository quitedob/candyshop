package rag

import "strings"

var countryAliases = map[string]string{
	// Americas
	"us":                       "usa",
	"united states":            "usa",
	"united states of america": "usa",
	"america":                  "usa",
	"can":                      "canada",
	"mx":                       "mexico",
	"br":                       "brazil",
	"ar":                       "argentina",
	// Middle East
	"ksa":                  "saudi arabia",
	"saudi":                "saudi arabia",
	"uae":                  "uae",
	"united arab emirates": "uae",
	"emirates":             "uae",
	"kw":                   "kuwait",
	"qa":                   "qatar",
	"om":                   "oman",
	"bh":                   "bahrain",
	"eg":                   "egypt",
	"tr":                   "turkey",
	"turkiye":              "turkey",
	// Europe
	"uk":             "uk",
	"gb":             "uk",
	"united kingdom": "uk",
	"great britain":  "uk",
	"britain":        "uk",
	"de":             "germany",
	"deutschland":    "germany",
	"fr":             "france",
	"nl":             "netherlands",
	"holland":        "netherlands",
	"be":             "belgium",
	"it":             "italy",
	"es":             "spain",
	"pl":             "poland",
	"ru":             "russia",
	"eu":             "europe",
	"european union": "europe",
	// Asia
	"prc":    "china",
	"cn":     "china",
	"jp":     "japan",
	"nippon": "japan",
	"kr":     "south korea",
	"korea":  "south korea",
	"tw":     "taiwan",
	"hk":     "hong kong",
	// Southeast Asia
	"id": "indonesia",
	"my": "malaysia",
	"th": "thailand",
	"vn": "vietnam",
	"ph": "philippines",
	"sg": "singapore",
	// South Asia
	"in": "india",
	"pk": "pakistan",
	"bd": "bangladesh",
	// Africa
	"za": "south africa",
	"ng": "nigeria",
	"ke": "kenya",
	"ma": "morocco",
	// Oceania
	"au": "australia",
	"nz": "new zealand",
}

func canonicalCountry(country string) string {
	normalized := strings.TrimSpace(strings.ToLower(country))
	if normalized == "" {
		return ""
	}
	if canonical, ok := countryAliases[normalized]; ok {
		return canonical
	}
	return normalized
}
