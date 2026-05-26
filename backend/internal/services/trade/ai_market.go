package trade

import "strings"

// CountryContext holds market-specific information for AI-assisted trade analysis.
type CountryContext struct {
	Regulations    string
	MarketPrefs    string
	ShippingNotes  string
	Certifications string
}

// shippingNote 组合通用运输时间说明与各国/地区特定备注（不含具体天数）。
func shippingNote(notes string) string {
	const prefix = "Estimated transit times are provided by our sales team in the order confirmation — not as fixed day counts on this website"
	if notes == "" {
		return prefix
	}
	return prefix + "; " + notes
}

// getCountryContext returns market-specific context for AI analysis.
func getCountryContext(country string) CountryContext {
	normalized := strings.TrimSpace(strings.ToLower(country))

	aliases := map[string]string{
		"us": "usa", "united states": "usa", "united states of america": "usa", "america": "usa",
		"can": "canada", "mx": "mexico", "br": "brazil", "ar": "argentina",
		"ksa": "saudi arabia", "saudi": "saudi arabia", "united arab emirates": "uae", "emirates": "uae",
		"kw": "kuwait", "qa": "qatar", "om": "oman", "bh": "bahrain", "eg": "egypt", "tr": "turkey", "türkiye": "turkey",
		"de": "germany", "deutschland": "germany", "fr": "france", "gb": "uk", "united kingdom": "uk", "britain": "uk",
		"nl": "netherlands", "holland": "netherlands", "be": "belgium", "it": "italy", "es": "spain", "pl": "poland", "ru": "russia",
		"jp": "japan", "nippon": "japan", "cn": "china", "prc": "china", "kr": "south korea", "korea": "south korea", "tw": "taiwan", "hk": "hong kong",
		"id": "indonesia", "my": "malaysia", "th": "thailand", "vn": "vietnam", "ph": "philippines", "sg": "singapore",
		"in": "india", "pk": "pakistan", "bd": "bangladesh",
		"za": "south africa", "ng": "nigeria", "ke": "kenya", "ma": "morocco",
		"au": "australia", "nz": "new zealand",
	}

	if alias, exists := aliases[normalized]; exists {
		normalized = alias
	}

	contexts := map[string]CountryContext{
		"usa": {
			Regulations:    "FDA registration required, FSMA compliance, nutrition labeling (English)",
			MarketPrefs:    "Lower sweetness, natural ingredients preferred, portion-controlled packaging",
			ShippingNotes:  shippingNote("import duty ~6%"),
			Certifications: "FDA, HACCP, GMP, OU Kosher (optional)",
		},
		"canada": {
			Regulations:    "CFIA registration, bilingual labeling (English/French)",
			MarketPrefs:    "Similar to US preferences, natural/organic trend, maple flavors popular",
			ShippingNotes:  shippingNote("NAFTA/CUSMA benefits, import duty ~0-3%"),
			Certifications: "CFIA, HACCP, SQF, Organic (optional)",
		},
		"mexico": {
			Regulations:    "COFEPRIS registration, Spanish labeling, NOM standards compliance",
			MarketPrefs:    "Spicy/chili flavors, tamarind combinations, value-oriented packaging",
			ShippingNotes:  shippingNote("USMCA benefits, import duty ~0-10%"),
			Certifications: "COFEPRIS, HACCP, Halal (optional)",
		},
		"brazil": {
			Regulations:    "ANVISA registration, Portuguese labeling, Mercosur standards",
			MarketPrefs:    "Tropical fruit flavors, brigadeiro-style, family-size packaging",
			ShippingNotes:  shippingNote("import duty ~14%"),
			Certifications: "ANVISA, HACCP, Halal (optional)",
		},
		"argentina": {
			Regulations:    "ANMAT registration, Spanish labeling, Mercosur standards",
			MarketPrefs:    "Dulce de leche flavors, chocolate coatings, premium positioning",
			ShippingNotes:  shippingNote("import restrictions possible, duty ~12%"),
			Certifications: "ANMAT, HACCP",
		},
		"saudi arabia": {
			Regulations:    "SFDA registration, mandatory Halal certification, Arabic labeling required",
			MarketPrefs:    "Higher sweetness accepted, date/flavor combinations popular, family-size packaging",
			ShippingNotes:  shippingNote("temperature-sensitive during summer, import duty ~5%"),
			Certifications: "Halal (mandatory), SFDA, HACCP",
		},
		"uae": {
			Regulations:    "ESMA registration, Halal required, bilingual labeling (Arabic/English)",
			MarketPrefs:    "Premium positioning, luxury packaging, gift sets popular",
			ShippingNotes:  shippingNote("free zones available, import duty ~5%"),
			Certifications: "Halal (mandatory), ESMA, HACCP",
		},
		"kuwait": {
			Regulations:    "KFDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Premium products, date combinations, gift packaging",
			ShippingNotes:  shippingNote("summer heat considerations, duty ~5%"),
			Certifications: "Halal (mandatory), KFDA, HACCP",
		},
		"qatar": {
			Regulations:    "Qatar Ministry registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Ultra-premium segment, luxury gift sets, high-quality packaging",
			ShippingNotes:  shippingNote("free zone options, duty ~5%"),
			Certifications: "Halal (mandatory), HACCP",
		},
		"oman": {
			Regulations:    "Oman FDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Mid-range pricing, family packaging, traditional flavors",
			ShippingNotes:  shippingNote("GCC standards apply, duty ~5%"),
			Certifications: "Halal (mandatory), HACCP",
		},
		"bahrain": {
			Regulations:    "NHRA registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Premium segment, diverse expat market, gift packaging",
			ShippingNotes:  shippingNote("regional hub via Dubai, duty ~5%"),
			Certifications: "Halal (mandatory), HACCP",
		},
		"egypt": {
			Regulations:    "EOS registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Value-oriented, local flavors, family-size packaging",
			ShippingNotes:  shippingNote("currency considerations, duty ~5-30%"),
			Certifications: "Halal (mandatory), EOS, HACCP",
		},
		"turkey": {
			Regulations:    "Turkish Ministry registration, Halal preferred, Turkish labeling",
			MarketPrefs:    "Moderate sweetness, lokum-style textures, gift packaging",
			ShippingNotes:  shippingNote("customs union with EU, duty ~0-10%"),
			Certifications: "Halal (preferred), HACCP, TSE",
		},
		"germany": {
			Regulations:    "EU food safety standards, EFSA compliance, German labeling required",
			MarketPrefs:    "Low sugar trend, organic preferred, sustainable packaging important",
			ShippingNotes:  shippingNote("strict customs documentation, import duty ~7%"),
			Certifications: "IFS, BRC, Organic (optional), Halal (optional)",
		},
		"france": {
			Regulations:    "EU food safety standards, DGCCRF compliance, French labeling required",
			MarketPrefs:    "Premium positioning, artisanal image, high quality chocolate coatings",
			ShippingNotes:  shippingNote("standard EU customs"),
			Certifications: "IFS, BRC, Organic (optional)",
		},
		"uk": {
			Regulations:    "FSA regulations, post-Brexit UK labeling requirements",
			MarketPrefs:    "Premium packaging, traditional flavors, vegetarian/vegan options popular",
			ShippingNotes:  shippingNote("customs documentation required"),
			Certifications: "BRCGS, HACCP",
		},
		"netherlands": {
			Regulations:    "EU food safety standards, NVWA compliance, Dutch/English labeling",
			MarketPrefs:    "Natural ingredients, sustainable packaging, licorice popular",
			ShippingNotes:  shippingNote("major port access (Rotterdam), duty ~7%"),
			Certifications: "IFS, BRC, HACCP",
		},
		"belgium": {
			Regulations:    "EU food safety standards, FAVV compliance, Dutch/French labeling",
			MarketPrefs:    "Premium chocolate coatings, praline-style, high quality focus",
			ShippingNotes:  shippingNote("EU distribution hub (Rotterdam/Antwerp)"),
			Certifications: "IFS, BRC, HACCP",
		},
		"italy": {
			Regulations:    "EU food safety standards, Ministry of Health compliance, Italian labeling",
			MarketPrefs:    "Premium positioning, hazelnut combinations, artisanal appeal",
			ShippingNotes:  shippingNote("standard EU customs"),
			Certifications: "IFS, BRC, HACCP, Organic (optional)",
		},
		"spain": {
			Regulations:    "EU food safety standards, AESAN compliance, Spanish labeling",
			MarketPrefs:    "Fruit flavors, turron-style, value to mid-range segments",
			ShippingNotes:  shippingNote("standard EU customs"),
			Certifications: "IFS, BRC, HACCP, Halal (optional)",
		},
		"poland": {
			Regulations:    "EU food safety standards, GIS compliance, Polish labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, growing premium segment",
			ShippingNotes:  shippingNote("EU distribution, competitive logistics costs"),
			Certifications: "IFS, BRC, HACCP",
		},
		"russia": {
			Regulations:    "EAEU standards, Rospotrebnadzor registration, Russian labeling",
			MarketPrefs:    "Moderate sweetness, chocolate coatings, gift packaging",
			ShippingNotes:  shippingNote("sanctions considerations, duty ~10%"),
			Certifications: "EAC, HACCP, Halal (optional)",
		},
		"japan": {
			Regulations:    "MHLW standards, Japanese labeling, strict quality requirements",
			MarketPrefs:    "Refined sweetness, seasonal flavors, gift packaging important",
			ShippingNotes:  shippingNote("strict import inspection, import duty ~3%"),
			Certifications: "JAS, Organic JAS (optional), Halal (optional)",
		},
		"china": {
			Regulations:    "GACC registration, Chinese labeling, GB standards compliance",
			MarketPrefs:    "Regional flavor preferences, gift packaging for holidays, e-commerce ready",
			ShippingNotes:  shippingNote("CIQ inspection required, import duty ~10%"),
			Certifications: "GACC registration, HACCP, ISO22000",
		},
		"south korea": {
			Regulations:    "MFDS registration, Korean labeling, strict quality standards",
			MarketPrefs:    "Low sugar trend, innovative flavors, premium packaging",
			ShippingNotes:  shippingNote("strict customs, duty ~8%"),
			Certifications: "MFDS, HACCP, Halal (optional)",
		},
		"taiwan": {
			Regulations:    "TFDA registration, Chinese labeling, food safety standards",
			MarketPrefs:    "Similar to Japan, bubble tea flavors, gift packaging",
			ShippingNotes:  shippingNote("standard customs, duty ~5%"),
			Certifications: "TFDA, HACCP, Halal (optional)",
		},
		"hong kong": {
			Regulations:    "CFS registration, English/Chinese labeling, food safety standards",
			MarketPrefs:    "Premium positioning, gift packaging, diverse flavors",
			ShippingNotes:  shippingNote("free port benefits, duty ~0%"),
			Certifications: "HACCP, ISO22000, Halal (optional)",
		},
		"indonesia": {
			Regulations:    "BPOM registration, MUI Halal mandatory, Indonesian labeling",
			MarketPrefs:    "Tropical fruit flavors, moderate sweetness, value packaging",
			ShippingNotes:  shippingNote("duty ~5-15%"),
			Certifications: "Halal MUI (mandatory), BPOM, HACCP",
		},
		"malaysia": {
			Regulations:    "NPRA registration, JAKIM Halal mandatory, Malay/English labeling",
			MarketPrefs:    "Durian flavors, pandan combinations, diverse ethnic market",
			ShippingNotes:  shippingNote("ASEAN benefits, duty ~0-5%"),
			Certifications: "Halal JAKIM (mandatory), NPRA, HACCP",
		},
		"thailand": {
			Regulations:    "FDA Thailand registration, Thai labeling, food safety standards",
			MarketPrefs:    "Tropical flavors, mango/mango sticky rice, gift packaging",
			ShippingNotes:  shippingNote("ASEAN benefits, duty ~0-5%"),
			Certifications: "Thai FDA, HACCP, Halal (optional)",
		},
		"vietnam": {
			Regulations:    "VFA registration, Vietnamese labeling, food safety standards",
			MarketPrefs:    "Coffee flavors, tropical fruits, growing premium segment",
			ShippingNotes:  shippingNote("ASEAN benefits, duty ~0-10%"),
			Certifications: "VFA, HACCP, Halal (optional)",
		},
		"philippines": {
			Regulations:    "FDA Philippines registration, English/Filipino labeling",
			MarketPrefs:    "Ube flavors, tropical fruits, value to mid-range",
			ShippingNotes:  shippingNote("ASEAN benefits, duty ~0-10%"),
			Certifications: "FDA Philippines, HACCP, Halal (optional)",
		},
		"singapore": {
			Regulations:    "SFA registration, English labeling, strict food safety",
			MarketPrefs:    "Premium positioning, innovative flavors, health-conscious",
			ShippingNotes:  shippingNote("free port, duty ~0%"),
			Certifications: "SFA, HACCP, Halal MUIS (optional)",
		},
		"india": {
			Regulations:    "FSSAI registration, Hindi/English labeling, strict standards",
			MarketPrefs:    "Mango flavors, spice combinations, value packaging, vegetarian",
			ShippingNotes:  shippingNote("duty ~30%"),
			Certifications: "FSSAI, HACCP, Halal (optional), Vegetarian mark",
		},
		"pakistan": {
			Regulations:    "PSQCA registration, mandatory Halal, Urdu labeling",
			MarketPrefs:    "Moderate sweetness, value-oriented, family packaging",
			ShippingNotes:  shippingNote("import restrictions possible, duty ~10-20%"),
			Certifications: "Halal (mandatory), PSQCA, HACCP",
		},
		"bangladesh": {
			Regulations:    "BFSA registration, mandatory Halal, Bengali labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, family packaging",
			ShippingNotes:  shippingNote("SAFTA benefits, duty ~10-25%"),
			Certifications: "Halal (mandatory), BFSA, HACCP",
		},
		"south africa": {
			Regulations:    "DTI registration, English labeling, SABS standards",
			MarketPrefs:    "Value to mid-range, diverse market, gift packaging",
			ShippingNotes:  shippingNote("AGOA benefits, duty ~5-20%"),
			Certifications: "SABS, HACCP, Halal (optional)",
		},
		"nigeria": {
			Regulations:    "NAFDAC registration, English labeling, food safety standards",
			MarketPrefs:    "Value-oriented, growing middle class, family packaging",
			ShippingNotes:  shippingNote("port congestion possible, duty ~5-20%"),
			Certifications: "NAFDAC, HACCP, Halal (optional)",
		},
		"kenya": {
			Regulations:    "KEBS registration, English labeling, EAC standards",
			MarketPrefs:    "Value to mid-range, tropical flavors, growing market",
			ShippingNotes:  shippingNote("EAC benefits, duty ~0-25%"),
			Certifications: "KEBS, HACCP, Halal (optional)",
		},
		"morocco": {
			Regulations:    "ONSSA registration, Arabic/French labeling, food safety standards",
			MarketPrefs:    "Almond/honey combinations, value to mid-range",
			ShippingNotes:  shippingNote("EU trade agreement, duty ~0-10%"),
			Certifications: "ONSSA, HACCP, Halal (optional)",
		},
		"australia": {
			Regulations:    "FSANZ standards, strict biosecurity, mandatory Australian labeling",
			MarketPrefs:    "Natural colors/flavors, bite-sized portions, high quality ingredients",
			ShippingNotes:  shippingNote("strict border inspection"),
			Certifications: "HACCP, SQF",
		},
		"new zealand": {
			Regulations:    "FSANZ standards, MPI requirements, English labeling",
			MarketPrefs:    "Natural/organic trend, premium positioning, sustainable packaging",
			ShippingNotes:  shippingNote("strict biosecurity, duty ~0-5%"),
			Certifications: "HACCP, MPI approved, Organic (optional)",
		},
	}

	if ctx, ok := contexts[normalized]; ok {
		return ctx
	}

	return CountryContext{
		Regulations:    "Check local food safety authority requirements",
		MarketPrefs:    "Research local taste preferences and packaging norms",
		ShippingNotes:  shippingNote("verify import regulations"),
		Certifications: "HACCP, ISO22000 recommended, check specific requirements",
	}
}
