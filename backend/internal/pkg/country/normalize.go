package country

import "strings"

// aliases 常见国家全名/别名 → ISO 3166-1 alpha-2
var aliases = map[string]string{
	"us": "US", "usa": "US", "u.s.": "US", "u.s.a.": "US",
	"united states": "US", "united states of america": "US", "america": "US",
	"uk": "UK", "gb": "UK", "gbr": "UK", "united kingdom": "UK", "great britain": "UK", "england": "UK",
	"de": "DE", "deu": "DE", "germany": "DE", "deutschland": "DE",
	"fr": "FR", "fra": "FR", "france": "FR",
	"ca": "CA", "can": "CA", "canada": "CA",
	"au": "AU", "aus": "AU", "australia": "AU",
	"jp": "JP", "jpn": "JP", "japan": "JP",
	"kr": "KR", "kor": "KR", "korea": "KR", "south korea": "KR", "republic of korea": "KR",
	"cn": "CN", "chn": "CN", "china": "CN", "prc": "CN", "people's republic of china": "CN",
	"sg": "SG", "sgp": "SG", "singapore": "SG",
	"my": "MY", "mys": "MY", "malaysia": "MY",
	"th": "TH", "tha": "TH", "thailand": "TH",
	"vn": "VN", "vnm": "VN", "vietnam": "VN", "viet nam": "VN",
	"id": "ID", "idn": "ID", "indonesia": "ID",
	"ph": "PH", "phl": "PH", "philippines": "PH",
	"in": "IN", "ind": "IN", "india": "IN",
	"pk": "PK", "pak": "PK", "pakistan": "PK",
	"ae": "AE", "are": "AE", "uae": "AE", "united arab emirates": "AE",
	"sa": "SA", "sau": "SA", "saudi arabia": "SA", "ksa": "SA",
	"mx": "MX", "mex": "MX", "mexico": "MX",
	"nl": "NL", "nld": "NL", "netherlands": "NL", "holland": "NL",
	"it": "IT", "ita": "IT", "italy": "IT",
	"es": "ES", "esp": "ES", "spain": "ES",
	"br": "BR", "bra": "BR", "brazil": "BR",
}

// NormalizeCountryCode 将自由文本国家名转为 ISO 码；已是 2–3 字母码则大写返回
func NormalizeCountryCode(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}
	key := strings.ToLower(raw)
	if code, ok := aliases[key]; ok {
		return code
	}
	if len(raw) <= 3 && !strings.Contains(raw, " ") {
		return strings.ToUpper(raw)
	}
	return strings.ToUpper(raw)
}
