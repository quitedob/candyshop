package trade

import "strings"

// CountryContext holds market-specific information for AI-assisted trade analysis.
type CountryContext struct {
	Regulations    string
	MarketPrefs    string
	ShippingNotes  string
	Certifications string
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
			ShippingNotes:  "5-7 days air freight, 20-30 days sea freight, import duty ~6%",
			Certifications: "FDA, HACCP, GMP, OU Kosher (optional)",
		},
		"canada": {
			Regulations:    "CFIA registration, bilingual labeling (English/French)",
			MarketPrefs:    "Similar to US preferences, natural/organic trend, maple flavors popular",
			ShippingNotes:  "3-5 days truck from US, NAFTA/CUSMA benefits, import duty ~0-3%",
			Certifications: "CFIA, HACCP, SQF, Organic (optional)",
		},
		"mexico": {
			Regulations:    "COFEPRIS registration, Spanish labeling, NOM standards compliance",
			MarketPrefs:    "Spicy/chili flavors, tamarind combinations, value-oriented packaging",
			ShippingNotes:  "3-5 days truck, USMCA benefits, import duty ~0-10%",
			Certifications: "COFEPRIS, HACCP, Halal (optional)",
		},
		"brazil": {
			Regulations:    "ANVISA registration, Portuguese labeling, Mercosur standards",
			MarketPrefs:    "Tropical fruit flavors, brigadeiro-style, family-size packaging",
			ShippingNotes:  "7-10 days air freight, 25-35 days sea freight, import duty ~14%",
			Certifications: "ANVISA, HACCP, Halal (optional)",
		},
		"argentina": {
			Regulations:    "ANMAT registration, Spanish labeling, Mercosur standards",
			MarketPrefs:    "Dulce de leche flavors, chocolate coatings, premium positioning",
			ShippingNotes:  "10-14 days air freight, import restrictions possible, duty ~12%",
			Certifications: "ANMAT, HACCP",
		},
		"saudi arabia": {
			Regulations:    "SFDA registration, mandatory Halal certification, Arabic labeling required",
			MarketPrefs:    "Higher sweetness accepted, date/flavor combinations popular, family-size packaging",
			ShippingNotes:  "7-10 days air freight, temperature-sensitive during summer, import duty ~5%",
			Certifications: "Halal (mandatory), SFDA, HACCP",
		},
		"uae": {
			Regulations:    "ESMA registration, Halal required, bilingual labeling (Arabic/English)",
			MarketPrefs:    "Premium positioning, luxury packaging, gift sets popular",
			ShippingNotes:  "5-7 days air freight, free zones available, import duty ~5%",
			Certifications: "Halal (mandatory), ESMA, HACCP",
		},
		"kuwait": {
			Regulations:    "KFDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Premium products, date combinations, gift packaging",
			ShippingNotes:  "7-10 days air freight, summer heat considerations, duty ~5%",
			Certifications: "Halal (mandatory), KFDA, HACCP",
		},
		"qatar": {
			Regulations:    "Qatar Ministry registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Ultra-premium segment, luxury gift sets, high-quality packaging",
			ShippingNotes:  "5-7 days air freight, free zone options, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"oman": {
			Regulations:    "Oman FDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Mid-range pricing, family packaging, traditional flavors",
			ShippingNotes:  "7-10 days air freight, GCC standards apply, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"bahrain": {
			Regulations:    "NHRA registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Premium segment, diverse expat market, gift packaging",
			ShippingNotes:  "5-7 days air freight via Dubai, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"egypt": {
			Regulations:    "EOS registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Value-oriented, local flavors, family-size packaging",
			ShippingNotes:  "7-10 days air freight, currency considerations, duty ~5-30%",
			Certifications: "Halal (mandatory), EOS, HACCP",
		},
		"turkey": {
			Regulations:    "Turkish Ministry registration, Halal preferred, Turkish labeling",
			MarketPrefs:    "Moderate sweetness, lokum-style textures, gift packaging",
			ShippingNotes:  "5-7 days truck/air, customs union with EU, duty ~0-10%",
			Certifications: "Halal (preferred), HACCP, TSE",
		},
		"germany": {
			Regulations:    "EU food safety standards, EFSA compliance, German labeling required",
			MarketPrefs:    "Low sugar trend, organic preferred, sustainable packaging important",
			ShippingNotes:  "3-5 days truck from port, strict customs documentation, import duty ~7%",
			Certifications: "IFS, BRC, Organic (optional), Halal (optional)",
		},
		"france": {
			Regulations:    "EU food safety standards, DGCCRF compliance, French labeling required",
			MarketPrefs:    "Premium positioning, artisanal image, high quality chocolate coatings",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, Organic (optional)",
		},
		"uk": {
			Regulations:    "FSA regulations, post-Brexit UK labeling requirements",
			MarketPrefs:    "Premium packaging, traditional flavors, vegetarian/vegan options popular",
			ShippingNotes:  "5-7 days air freight, customs documentation required",
			Certifications: "BRCGS, HACCP",
		},
		"netherlands": {
			Regulations:    "EU food safety standards, NVWA compliance, Dutch/English labeling",
			MarketPrefs:    "Natural ingredients, sustainable packaging, licorice popular",
			ShippingNotes:  "Major port access (Rotterdam), 1-3 days distribution, duty ~7%",
			Certifications: "IFS, BRC, HACCP",
		},
		"belgium": {
			Regulations:    "EU food safety standards, FAVV compliance, Dutch/French labeling",
			MarketPrefs:    "Premium chocolate coatings, praline-style, high quality focus",
			ShippingNotes:  "3-5 days from Rotterdam/Antwerp, EU distribution hub",
			Certifications: "IFS, BRC, HACCP",
		},
		"italy": {
			Regulations:    "EU food safety standards, Ministry of Health compliance, Italian labeling",
			MarketPrefs:    "Premium positioning, hazelnut combinations, artisanal appeal",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, HACCP, Organic (optional)",
		},
		"spain": {
			Regulations:    "EU food safety standards, AESAN compliance, Spanish labeling",
			MarketPrefs:    "Fruit flavors, turron-style, value to mid-range segments",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, HACCP, Halal (optional)",
		},
		"poland": {
			Regulations:    "EU food safety standards, GIS compliance, Polish labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, growing premium segment",
			ShippingNotes:  "3-5 days truck from EU port, lower distribution costs",
			Certifications: "IFS, BRC, HACCP",
		},
		"russia": {
			Regulations:    "EAEU standards, Rospotrebnadzor registration, Russian labeling",
			MarketPrefs:    "Moderate sweetness, chocolate coatings, gift packaging",
			ShippingNotes:  "10-14 days rail/truck, sanctions considerations, duty ~10%",
			Certifications: "EAC, HACCP, Halal (optional)",
		},
		"japan": {
			Regulations:    "MHLW standards, Japanese labeling, strict quality requirements",
			MarketPrefs:    "Refined sweetness, seasonal flavors, gift packaging important",
			ShippingNotes:  "3-5 days air freight, strict import inspection, import duty ~3%",
			Certifications: "JAS, Organic JAS (optional), Halal (optional)",
		},
		"china": {
			Regulations:    "GACC registration, Chinese labeling, GB standards compliance",
			MarketPrefs:    "Regional flavor preferences, gift packaging for holidays, e-commerce ready",
			ShippingNotes:  "5-7 days air freight, CIQ inspection required, import duty ~10%",
			Certifications: "GACC registration, HACCP, ISO22000",
		},
		"south korea": {
			Regulations:    "MFDS registration, Korean labeling, strict quality standards",
			MarketPrefs:    "Low sugar trend, innovative flavors, premium packaging",
			ShippingNotes:  "3-5 days air freight, strict customs, duty ~8%",
			Certifications: "MFDS, HACCP, Halal (optional)",
		},
		"taiwan": {
			Regulations:    "TFDA registration, Chinese labeling, food safety standards",
			MarketPrefs:    "Similar to Japan, bubble tea flavors, gift packaging",
			ShippingNotes:  "3-5 days air freight, standard customs, duty ~5%",
			Certifications: "TFDA, HACCP, Halal (optional)",
		},
		"hong kong": {
			Regulations:    "CFS registration, English/Chinese labeling, food safety standards",
			MarketPrefs:    "Premium positioning, gift packaging, diverse flavors",
			ShippingNotes:  "3-5 days air freight, free port benefits, duty ~0%",
			Certifications: "HACCP, ISO22000, Halal (optional)",
		},
		"indonesia": {
			Regulations:    "BPOM registration, MUI Halal mandatory, Indonesian labeling",
			MarketPrefs:    "Tropical fruit flavors, moderate sweetness, value packaging",
			ShippingNotes:  "5-7 days air freight, 15-20 days sea, duty ~5-15%",
			Certifications: "Halal MUI (mandatory), BPOM, HACCP",
		},
		"malaysia": {
			Regulations:    "NPRA registration, JAKIM Halal mandatory, Malay/English labeling",
			MarketPrefs:    "Durian flavors, pandan combinations, diverse ethnic market",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-5%",
			Certifications: "Halal JAKIM (mandatory), NPRA, HACCP",
		},
		"thailand": {
			Regulations:    "FDA Thailand registration, Thai labeling, food safety standards",
			MarketPrefs:    "Tropical flavors, mango/mango sticky rice, gift packaging",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-5%",
			Certifications: "Thai FDA, HACCP, Halal (optional)",
		},
		"vietnam": {
			Regulations:    "VFA registration, Vietnamese labeling, food safety standards",
			MarketPrefs:    "Coffee flavors, tropical fruits, growing premium segment",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-10%",
			Certifications: "VFA, HACCP, Halal (optional)",
		},
		"philippines": {
			Regulations:    "FDA Philippines registration, English/Filipino labeling",
			MarketPrefs:    "Ube flavors, tropical fruits, value to mid-range",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-10%",
			Certifications: "FDA Philippines, HACCP, Halal (optional)",
		},
		"singapore": {
			Regulations:    "SFA registration, English labeling, strict food safety",
			MarketPrefs:    "Premium positioning, innovative flavors, health-conscious",
			ShippingNotes:  "3-5 days air freight, free port, duty ~0%",
			Certifications: "SFA, HACCP, Halal MUIS (optional)",
		},
		"india": {
			Regulations:    "FSSAI registration, Hindi/English labeling, strict standards",
			MarketPrefs:    "Mango flavors, spice combinations, value packaging, vegetarian",
			ShippingNotes:  "5-7 days air freight, 20-30 days sea, duty ~30%",
			Certifications: "FSSAI, HACCP, Halal (optional), Vegetarian mark",
		},
		"pakistan": {
			Regulations:    "PSQCA registration, mandatory Halal, Urdu labeling",
			MarketPrefs:    "Moderate sweetness, value-oriented, family packaging",
			ShippingNotes:  "5-7 days air freight, import restrictions possible, duty ~10-20%",
			Certifications: "Halal (mandatory), PSQCA, HACCP",
		},
		"bangladesh": {
			Regulations:    "BFSA registration, mandatory Halal, Bengali labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, family packaging",
			ShippingNotes:  "7-10 days air freight, SAFTA benefits, duty ~10-25%",
			Certifications: "Halal (mandatory), BFSA, HACCP",
		},
		"south africa": {
			Regulations:    "DTI registration, English labeling, SABS standards",
			MarketPrefs:    "Value to mid-range, diverse market, gift packaging",
			ShippingNotes:  "7-10 days air freight, AGOA benefits, duty ~5-20%",
			Certifications: "SABS, HACCP, Halal (optional)",
		},
		"nigeria": {
			Regulations:    "NAFDAC registration, English labeling, food safety standards",
			MarketPrefs:    "Value-oriented, growing middle class, family packaging",
			ShippingNotes:  "7-10 days air freight, port congestion possible, duty ~5-20%",
			Certifications: "NAFDAC, HACCP, Halal (optional)",
		},
		"kenya": {
			Regulations:    "KEBS registration, English labeling, EAC standards",
			MarketPrefs:    "Value to mid-range, tropical flavors, growing market",
			ShippingNotes:  "7-10 days air freight, EAC benefits, duty ~0-25%",
			Certifications: "KEBS, HACCP, Halal (optional)",
		},
		"morocco": {
			Regulations:    "ONSSA registration, Arabic/French labeling, food safety standards",
			MarketPrefs:    "Almond/honey combinations, value to mid-range",
			ShippingNotes:  "5-7 days air freight, EU trade agreement, duty ~0-10%",
			Certifications: "ONSSA, HACCP, Halal (optional)",
		},
		"australia": {
			Regulations:    "FSANZ standards, strict biosecurity, mandatory Australian labeling",
			MarketPrefs:    "Natural colors/flavors, bite-sized portions, high quality ingredients",
			ShippingNotes:  "7-10 days air freight, strict border inspection",
			Certifications: "HACCP, SQF",
		},
		"new zealand": {
			Regulations:    "FSANZ standards, MPI requirements, English labeling",
			MarketPrefs:    "Natural/organic trend, premium positioning, sustainable packaging",
			ShippingNotes:  "7-10 days air freight, strict biosecurity, duty ~0-5%",
			Certifications: "HACCP, MPI approved, Organic (optional)",
		},
	}

	if ctx, ok := contexts[normalized]; ok {
		return ctx
	}

	return CountryContext{
		Regulations:    "Check local food safety authority requirements",
		MarketPrefs:    "Research local taste preferences and packaging norms",
		ShippingNotes:  "Verify shipping routes and import regulations",
		Certifications: "HACCP, ISO22000 recommended, check specific requirements",
	}
}
