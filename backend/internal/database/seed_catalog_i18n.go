package database

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"strings"

	"gorm.io/gorm"
)

// catalogSeedLocales 种子数据覆盖的全部语言。
var catalogSeedLocales = []string{"en", "zh", "ko", "ja", "ar", "th", "vi", "id", "ms"}

// categoryI18nSeed 分类 9 语言翻译种子。
var categoryI18nSeed = map[string]map[string]map[string]string{
	"gummy-candy": {
		"en": {"name": "Gummy Candy", "alias": "Gummies", "description": "Soft, chewy candies in various fruit flavors and fun shapes."},
		"zh": {"name": "软糖", "alias": "QQ糖", "description": "多种水果口味与趣味造型的柔软耐嚼糖果。"},
		"ko": {"name": "구미 캔디", "alias": "젤리", "description": "다양한 과일 맛과 재미있는 모양의 부드럽고 쫄깃한 캔디."},
		"ja": {"name": "グミキャンディ", "alias": "グミ", "description": "さまざまなフルーツ味と楽しい形のやわらかく chewy なキャンディ。"},
		"ar": {"name": "حلوى الجummies", "alias": "جيلي", "description": "حلوى طرية وممضغة بنكهات فاكهية متنوعة وأشكال ممتعة."},
		"th": {"name": "ขนมเยลลี่", "alias": "กัมมี่", "description": "ขนมเคี้ยวนุ่ม หลากรสผลไม้ และรูปทรงสนุกสนาน"},
		"vi": {"name": "Kẹo dẻo", "alias": "Gummy", "description": "Kẹo mềm dai với nhiều hương vị trái cây và hình dáng thú vị."},
		"id": {"name": "Permen Gummy", "alias": "Jelly", "description": "Permen kenyal dengan berbagai rasa buah dan bentuk menarik."},
		"ms": {"name": "Gula-gula Gummy", "alias": "Jeli", "description": "Gula-gula kenyal dengan pelbagai perisa buah dan bentuk menarik."},
	},
	"hard-candy": {
		"en": {"name": "Hard Candy", "alias": "Boiled Candy", "description": "Classic boiled candies with fruit fillings and refreshing flavors."},
		"zh": {"name": "硬糖", "alias": "水果硬糖", "description": "经典煮制硬糖，内含水果夹心与清爽风味。"},
		"ko": {"name": "하드 캔디", "alias": "사탕", "description": "과일 필링과 상쾌한 맛의 클래식 끓인 사탕."},
		"ja": {"name": "ハードキャンディ", "alias": "飴", "description": "フルーツフィリングと爽やかな味わいのクラシックな Boiled キャンディ。"},
		"ar": {"name": "حلوى صلبة", "alias": "سكاكر", "description": "حلوى كلاسيكية مطبوخة بحشوات فاكهية ونكهات منعشة."},
		"th": {"name": "ลูกอมแข็ง", "alias": "แคนดี้", "description": "ลูกอมต้มแบบคลาสสิก ไส้ผลไม้ รสสดชื่น"},
		"vi": {"name": "Kẹo cứng", "alias": "Kẹo viên", "description": "Kẹo nấu cổ điển với nhân trái cây và hương vị tươi mát."},
		"id": {"name": "Permen Keras", "alias": "Bonbon", "description": "Permen rebus klasik dengan isian buah dan rasa segar."},
		"ms": {"name": "Gula-gula Keras", "alias": "Gula-gula rebus", "description": "Gula-gula rebus klasik dengan inti buah dan rasa menyegarkan."},
	},
	"aerated-candy": {
		"en": {"name": "Aerated Candy", "alias": "Sponge Candy", "description": "Light and airy candies with unique textures and mouthfeel."},
		"zh": {"name": "充气糖果", "alias": "海绵糖", "description": "质地轻盈、口感独特的充气糖果。"},
		"ko": {"name": "발포 캔디", "alias": "스펀지 캔디", "description": "가볍고 독특한 식감의 발포 캔디."},
		"ja": {"name": "発泡キャンディ", "alias": "スポンジキャンディ", "description": "軽やかで独特な食感の発泡キャンディ。"},
		"ar": {"name": "حلوى مهوية", "alias": "حلوى إسفنجية", "description": "حلوى خفيفة وهشة بقوام فريد."},
		"th": {"name": "ขนมฟอง", "alias": "ขนมโฟม", "description": "ขนมเบาและมีเอกลักษณ์ด้านเนื้อสัมผัส"},
		"vi": {"name": "Kẹo bọt khí", "alias": "Kẹo xốp", "description": "Kẹo nhẹ và xốp với kết cấu độc đáo."},
		"id": {"name": "Permen Aerated", "alias": "Permen spons", "description": "Permen ringan dengan tekstur unik."},
		"ms": {"name": "Gula-gula Berudara", "alias": "Gula-gula span", "description": "Gula-gula ringan dengan tekstur unik."},
	},
	"toffee-candy": {
		"en": {"name": "Toffee Candy", "alias": "Toffee", "description": "Rich, buttery toffees with smooth caramel flavors."},
		"zh": {"name": "太妃糖", "alias": "焦糖太妃", "description": "浓郁黄油太妃，顺滑焦糖风味。"},
		"ko": {"name": "토피 캔디", "alias": "토피", "description": "버터 풍미가 풍부하고 부드러운 캐러멜 맛의 토피."},
		"ja": {"name": "トフィーキャンディ", "alias": "トフィー", "description": "バターのコクと滑らかなキャラメル風味のトフィー。"},
		"ar": {"name": "حلوى التoffee", "alias": "توفي", "description": "توفي غني بالزبدة بنكهة كراميل ناعمة."},
		"th": {"name": "ทoffee", "alias": "คาราเมลทoffee", "description": "ทoffee เนยเข้ม รสคาราเมลลื่นไหล"},
		"vi": {"name": "Kẹo toffee", "alias": "Toffee", "description": "Toffee bơ đậm với hương caramel mịn."},
		"id": {"name": "Permen Toffee", "alias": "Toffee", "description": "Toffee kaya mentega dengan rasa karamel lembut."},
		"ms": {"name": "Gula-gula Toffee", "alias": "Toffee", "description": "Toffee mentega kaya dengan rasa karamel licin."},
	},
	"compound-chocolate": {
		"en": {"name": "Compound Chocolate", "alias": "Compound Coating", "description": "Cost-effective chocolate alternatives for enrobing and moulding."},
		"zh": {"name": "代可可脂巧克力", "alias": "复合巧克力", "description": "性价比高的巧克力替代方案，适用于包衣与模制。"},
		"ko": {"name": "컴파운드 초콜릿", "alias": "대체 초콜릿", "description": "코팅 및 몰딩에 적합한 경제적인 초콜릿 대체재."},
		"ja": {"name": "コンパウンドチョコレート", "alias": "代替チョコ", "description": "コーティングやモルディング向けのコスト効率の良いチョコレート代替品。"},
		"ar": {"name": "شوكولاتة مركبة", "alias": "بديل الشوكولاتة", "description": "بدائل شوكولاتة اقتصادية للتغليف والقولبة."},
		"th": {"name": "ช็อกโกแลตคอมพาวด์", "alias": "เคลือบช็อกโกแลต", "description": "ทางเลือกช็อกโกแลตคุ้มค่าสำหรับเคลือบและขึ้นรูป"},
		"vi": {"name": "Socola compound", "alias": "Socola thay thế", "description": "Giải pháp socola thay thế hiệu quả cho phủ và đúc khuôn."},
		"id": {"name": "Cokelat Compound", "alias": "Coating cokelat", "description": "Alternatif cokelat ekonomis untuk enrobing dan moulding."},
		"ms": {"name": "Coklat Kompaun", "alias": "Salutan coklat", "description": "Alternatif coklat kos efektif untuk salutan dan acuan."},
	},
	"licorice": {
		"en": {"name": "Licorice Candy", "alias": "Licorice", "description": "Traditional and modern licorice varieties with chewy texture."},
		"zh": {"name": "甘草糖", "alias": "甘草软糖", "description": "传统与现代甘草品类，风味浓郁、口感耐嚼。"},
		"ko": {"name": "감초 캔디", "alias": "리코리스", "description": "전통 및 현대 감초 품종, 쫄깃한 식감."},
		"ja": {"name": "リコリスキャンディ", "alias": "甘草", "description": "伝統的・現代的なリコリス、 chewy な食感。"},
		"ar": {"name": "حلوى العرقسوس", "alias": "عرقسوس", "description": "أصناف عرقسوس تقليدية وحديثة بقوام ممضغ."},
		"th": {"name": "ขนมชะเอมเทศ", "alias": "ลิคควิด", "description": "ลิคควิดแบบดั้งเดิมและสมัยใหม่ เนื้อเคี้ยว"},
		"vi": {"name": "Kẹo cam thảo", "alias": "Licorice", "description": "Kẹo cam thảo truyền thống và hiện đại, dai."},
		"id": {"name": "Permen Licorice", "alias": "Akar manis", "description": "Varietas licorice tradisional dan modern, kenyal."},
		"ms": {"name": "Gula-gula Licorice", "alias": "Akar manis", "description": "Pelbagai licorice tradisional dan moden, kenyal."},
	},
	"sour-candies": {
		"en": {"name": "Sour Candies", "alias": "Sour Gummies", "description": "Tangy sour coated candies with intense fruit flavors."},
		"zh": {"name": "酸糖", "alias": "酸味糖果", "description": "外覆酸砂、果味浓郁的酸糖。"},
		"ko": {"name": "새콤 캔디", "alias": "새콤 젤리", "description": "강렬한 과일 맛의 새콤한 산미 코팅 캔디."},
		"ja": {"name": "サワーキャンディ", "alias": "酸っぱいグミ", "description": "強いフルーツ風味の酸っぱいコーティングキャンディ。"},
		"ar": {"name": "حلوى حامضة", "alias": "جيلي حامض", "description": "حلوى مغطاة بنكهات فاكهية حامضة مكثفة."},
		"th": {"name": "ขนมเปรี้ยว", "alias": "กัมมี่เปรี้ยว", "description": "ขนมเคลือบรสเปรี้ยว กลิ่นผลไม้เข้ม"},
		"vi": {"name": "Kẹo chua", "alias": "Gummy chua", "description": "Kẹo phủ vị chua với hương trái cây đậm."},
		"id": {"name": "Permen Asam", "alias": "Gummy asam", "description": "Permen lapisan asam dengan rasa buah intens."},
		"ms": {"name": "Gula-gula Masam", "alias": "Gummy masam", "description": "Gula-gula salutan masam dengan perisa buah kuat."},
	},
}

// SeedCategoryTranslations merges category i18n seeds into existing categories
// (idempotent, fill-only — never reverts admin edits, M3 / G27-c).
func SeedCategoryTranslations(db *gorm.DB) error {
	return seedAllCategoryTranslations(db)
}

// seedAllCategoryTranslations 写入分类 9 语言翻译（fill-only，保留已有翻译与管理员标量编辑）。
// M3 / G27-c: every category is routed through applyCategoryI18nSeedGuarded
// (seed.go), so the unconditional every-boot call from cmd/api/main.go:85 is
// idempotent — a translation key is only written when empty/whitespace, the en
// fallback backfills empty scalar Name/Alias/Description, and the zh name is only
// mapped onto the scalar Name on the initial seed run (never on an already
// localized category). Admin-authored zh/scalar edits survive restarts.
func seedAllCategoryTranslations(db *gorm.DB) error {
	for slug, localeMap := range categoryI18nSeed {
		var cat modelsProduct.Category
		if err := db.Where("slug = ?", slug).First(&cat).Error; err != nil {
			continue
		}
		applyCategoryI18nSeedGuarded(&cat, localeMap)
		if err := db.Model(&cat).Updates(map[string]interface{}{
			"translations": cat.Translations,
			"name":         cat.Name,
			"alias":        cat.Alias,
			"description":  cat.Description,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// repairProductZhCopiedFromEN 删除被误从 en 复制到 zh 的翻译（zh 主语言不应含英文副本）。
func repairProductZhCopiedFromEN(p *modelsProduct.Product) {
	if p.Translations == nil {
		return
	}
	en, hasEN := p.Translations["en"]
	zh, hasZH := p.Translations["zh"]
	if !hasEN || !hasZH || len(en) == 0 || len(zh) == 0 {
		return
	}
	enName := strings.TrimSpace(en["name"])
	zhName := strings.TrimSpace(zh["name"])
	if enName == "" || zhName == "" || enName != zhName {
		return
	}
	delete(p.Translations, "zh")
}

// syncProductScalarsFromZh 将 zh 翻译同步到产品标量列（主语言为中文）。
func syncProductScalarsFromZh(p *modelsProduct.Product) {
	if p.Translations == nil {
		return
	}
	zh, ok := p.Translations["zh"]
	if !ok || len(zh) == 0 {
		return
	}
	if v := strings.TrimSpace(zh["name"]); v != "" {
		p.Name = v
	}
	if v := strings.TrimSpace(zh["summary"]); v != "" {
		p.Summary = v
	}
	if v := strings.TrimSpace(zh["description"]); v != "" {
		p.Description = v
	}
	if v := strings.TrimSpace(zh["category"]); v != "" {
		p.Category = v
	}
	if v := strings.TrimSpace(zh["ingredients"]); v != "" {
		p.Ingredients = v
	}
	if v := strings.TrimSpace(zh["allergens"]); v != "" {
		p.Allergens = v
	}
	if v := strings.TrimSpace(zh["storage"]); v != "" {
		p.Storage = v
	}
	if v := strings.TrimSpace(zh["shelfLife"]); v != "" {
		p.ShelfLife = v
	}
	if v := strings.TrimSpace(zh["leadTime"]); v != "" {
		p.LeadTime = v
	}
}

// seedProductMissingLocales 为产品补全缺失 locale（优先保留已有，否则从 en 复制）。
func seedProductMissingLocales(db *gorm.DB) error {
	var products []modelsProduct.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}
	copyFields := []string{"name", "summary", "description", "category", "ingredients", "allergens", "storage", "shelfLife", "leadTime", "flavors", "shapes"}
	for i := range products {
		p := &products[i]
		if p.Translations == nil {
			p.Translations = make(modelsCommon.JSONMap)
		}
		seedEnglishProductTranslation(p)
		en, hasEN := p.Translations["en"]
		if !hasEN || len(en) == 0 {
			continue
		}
		repairProductZhCopiedFromEN(p)
		syncProductScalarsFromZh(p)
		for _, loc := range catalogSeedLocales {
			if loc == "en" || loc == "zh" {
				continue
			}
			if p.Translations[loc] != nil && strings.TrimSpace(p.Translations[loc]["name"]) != "" {
				continue
			}
			if p.Translations[loc] == nil {
				p.Translations[loc] = make(map[string]string)
			}
			for _, field := range copyFields {
				if strings.TrimSpace(p.Translations[loc][field]) == "" && strings.TrimSpace(en[field]) != "" {
					p.Translations[loc][field] = en[field]
				}
			}
		}
		syncProductScalarsFromZh(p)
		if err := db.Model(p).Updates(map[string]interface{}{
			"translations": p.Translations,
			"name":         p.Name,
			"summary":      p.Summary,
			"description":  p.Description,
			"category":     p.Category,
			"ingredients":  p.Ingredients,
			"allergens":    p.Allergens,
			"storage":      p.Storage,
			"shelf_life":   p.ShelfLife,
			"lead_time":    p.LeadTime,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
