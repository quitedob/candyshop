package database

import (
	modelsAuth "candypro/api/internal/models/auth"
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
)

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"candypro/api/internal/roles"
	"candypro/api/internal/utils"

	"gorm.io/gorm"
)

func ensureSystemRoles(db *gorm.DB) (map[string]string, error) {
	roleSeeds := []modelsAuth.Role{
		{
			ID:          "role_customer",
			Name:        roles.User,
			Description: "B2B user account",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "role_admin",
			Name:        roles.Admin,
			Description: "Operations and sales admin",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "role_superadmin",
			Name:        roles.SuperAdmin,
			Description: "Platform super administrator",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	roleIDs := make(map[string]string, len(roleSeeds))
	for _, role := range roleSeeds {
		var existing modelsAuth.Role
		err := db.Where("name = ?", role.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := db.Create(&role).Error; createErr != nil {
				return nil, createErr
			}
			roleIDs[role.Name] = role.ID
			continue
		}
		if err != nil {
			return nil, err
		}

		if existing.Description == "" || !existing.IsSystem {
			existing.Description = role.Description
			existing.IsSystem = true
			existing.UpdatedAt = time.Now()
			if saveErr := db.Save(&existing).Error; saveErr != nil {
				return nil, saveErr
			}
		}
		roleIDs[existing.Name] = existing.ID
	}

	return roleIDs, nil
}

func resolveSuperadminBootstrapCredentials() (string, string, error) {
	email := strings.TrimSpace(os.Getenv("SEED_SUPERADMIN_EMAIL"))
	if email == "" {
		email = "admin@candypro.com"
	}

	passwordHash := strings.TrimSpace(os.Getenv("SEED_SUPERADMIN_PASSWORD_HASH"))
	if passwordHash != "" {
		return email, passwordHash, nil
	}

	password := strings.TrimSpace(os.Getenv("SEED_SUPERADMIN_PASSWORD"))
	if password == "" {
		return "", "", fmt.Errorf("missing SEED_SUPERADMIN_PASSWORD or SEED_SUPERADMIN_PASSWORD_HASH for superadmin bootstrap")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return "", "", err
	}
	return email, hashedPassword, nil
}

// seedSuperadmin seeds an initial superadmin user
func seedSuperadmin(db *gorm.DB) error {
	superadminEmail, hashedPassword, err := resolveSuperadminBootstrapCredentials()
	if err != nil {
		return err
	}

	roleIDs, err := ensureSystemRoles(db)
	if err != nil {
		return err
	}
	superadminRoleID := roleIDs[roles.SuperAdmin]

	var existing modelsUser.User
	err = db.Where("email = ?", superadminEmail).First(&existing).Error
	if err == nil {
		needsUpdate := false
		if existing.RoleID != superadminRoleID {
			existing.RoleID = superadminRoleID
			needsUpdate = true
		}
		if existing.Status != "active" {
			existing.Status = "active"
			needsUpdate = true
		}
		if !existing.EmailVerified {
			existing.EmailVerified = true
			now := time.Now()
			existing.EmailVerifiedAt = &now
			needsUpdate = true
		}
		// H8: Never overwrite an existing password — only set on first creation
		// Password changes must go through the change-password endpoint
		if needsUpdate {
			existing.UpdatedAt = time.Now()
			return db.Save(&existing).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	superadmin := modelsUser.User{
		ID:              utils.GenerateID(),
		FirstName:       "Super",
		LastName:        "Admin",
		Email:           superadminEmail,
		PasswordHash:    hashedPassword,
		Status:          "active",
		EmailVerified:   true,
		EmailVerifiedAt: &now,
		RoleID:          superadminRoleID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return db.Create(&superadmin).Error
}

// seedCategories seeds product categories
func seedCategories(db *gorm.DB) error {
	categories := []modelsProduct.Category{
		{
			Slug:        "gummy-candy",
			Name:        "Gummy Candy",
			Description: "Soft, chewy candies in various fruit flavors and fun shapes. Perfect for all ages with natural fruit juice options available.",
			Thumbnail:   "/images/categories/gummy-candy.jpg",
			Icon:        "gummy",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "hard-candy",
			Name:        "Hard Candy",
			Description: "Classic boiled candies with fruit fillings and refreshing flavors. Long-lasting taste in elegant packaging options.",
			Thumbnail:   "/images/categories/hard-candy.jpg",
			Icon:        "hard",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "aerated-candy",
			Name:        "Aerated Candy",
			Description: "Light and airy candies with unique textures and mouthfeel. Innovative production techniques create a one-of-a-kind confectionery experience.",
			Thumbnail:   "/images/categories/aerated-candy.jpg",
			Icon:        "aerated",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "toffee-candy",
			Name:        "Toffee Candy",
			Description: "Rich, buttery toffees with smooth caramel flavors. Classic recipes refined for modern tastes with premium ingredients.",
			Thumbnail:   "/images/categories/toffee-candy.jpg",
			Icon:        "toffee",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "compound-chocolate",
			Name:        "Compound Chocolate",
			Description: "Cost-effective chocolate alternatives with excellent taste and texture. Perfect for enrobing, moulding, and private label chocolate products.",
			Thumbnail:   "/images/categories/compound-chocolate.jpg",
			Icon:        "chocolate",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "licorice",
			Name:        "Licorice Candy",
			Description: "Traditional and modern licorice varieties including classic black licorice twists and fruity red licorice ropes. Rich flavors with chewy texture.",
			Thumbnail:   "/images/categories/licorice.jpg",
			Icon:        "licorice",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Slug:        "sour-candies",
			Name:        "Sour Candies",
			Description: "Tangy and sour coated candies with intense fruit flavors. From mild to extreme sour, perfect for thrill-seeking candy lovers.",
			Thumbnail:   "/images/categories/sour-candies.jpg",
			Icon:        "sour",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, cat := range categories {
		if err := db.FirstOrCreate(&cat, modelsProduct.Category{Slug: cat.Slug}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedProducts seeds product data
func seedProducts(db *gorm.DB) error {
	products := []modelsProduct.Product{
		{
			ID:             utils.GenerateID(),
			Slug:           "4d-fruit-gummy",
			Name:           "4D Fruit Gummy",
			Summary:        "Real fruit juice gummy with natural flavors and fun 3D shapes",
			Description:    "Our signature 4D Fruit Gummy combines real fruit juice with innovative molding technology to create candies that look as good as they taste. Available in strawberry, orange, grape, and apple flavors with fun animal and fruit shapes.",
			Category:       "Gummy Candy",
			CategorySlug:   "gummy-candy",
			Thumbnail:      "/images/products/4d-fruit-gummy.jpg",
			Images:         modelsCommon.StringArray{"/images/products/4d-fruit-gummy.jpg", "/images/products/4d-fruit-gummy-2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"},
			MOQ:            5000,
			StockQuantity:  260000,
			LeadTime:       "15-20 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Orange", "Grape", "Apple", "Mango"},
			Shapes:         modelsCommon.StringArray{"Bear", "Worm", "Fruit", "Custom"},
			Ingredients:    "Sugar, Glucose Syrup, Fruit Juice Concentrate, Gelatin, Citric Acid, Natural Flavors",
			Allergens:      "May contain traces of dairy",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place below 25°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "crystal-hard-candy",
			Name:           "Crystal Hard Candy",
			Summary:        "Premium crystal-clear hard candies with natural fruit flavors",
			Description:    "Crystal Hard Candy offers a premium candy experience with our signature transparency and long-lasting flavor. Each piece is crafted with precision to achieve perfect clarity and taste.",
			Category:       "Hard Candy",
			CategorySlug:   "hard-candy",
			Thumbnail:      "/images/products/crystal-hard-candy.jpg",
			Images:         modelsCommon.StringArray{"/images/products/crystal-hard-candy.jpg", "/images/products/crystal-hard-candy-2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            10000,
			StockQuantity:  180000,
			LeadTime:       "20-25 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Lemon", "Mint", "Grape", "Melon"},
			Shapes:         modelsCommon.StringArray{"Round", "Oval", "Square", "Custom"},
			Ingredients:    "Sugar, Glucose Syrup, Natural Flavors, Citric Acid, Colors",
			Allergens:      "None",
			ShelfLife:      "24 months",
			Storage:        "Store in cool, dry place below 25°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "rainbow-lollipop",
			Name:           "Rainbow Swirl Lollipop",
			Summary:        "Colorful swirl lollipops with multi-layered flavors",
			Description:    "Our Rainbow Swirl Lollipop is a visual and taste sensation. Each lollipop features beautiful color swirls with complementary flavors that create a unique taste experience.",
			Category:       "Aerated Candy",
			CategorySlug:   "aerated-candy",
			Thumbnail:      "/images/products/rainbow-lollipop.jpg",
			Images:         modelsCommon.StringArray{"/images/products/rainbow-lollipop.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "Halal"},
			MOQ:            8000,
			StockQuantity:  140000,
			LeadTime:       "18-22 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Rainbow Mix", "Berry Blast", "Tropical", "Sour Mix"},
			Shapes:         modelsCommon.StringArray{"Round", "Heart", "Star", "Custom"},
			Ingredients:    "Sugar, Glucose Syrup, Natural Flavors, Citric Acid, Natural Colors",
			Allergens:      "None",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place below 25°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "sour-belt",
			Name:           "Sour Belt Candy",
			Summary:        "Tangy sour belt candies with intense fruit flavors",
			Description:    "Sour Belt Candy delivers the perfect balance of sweet and sour. These long, chewy belts are coated with sour sugar crystals for an extra zesty kick.",
			Category:       "Compound Chocolate",
			CategorySlug:   "compound-chocolate",
			Thumbnail:      "/images/products/sour-belt.jpg",
			Images:         modelsCommon.StringArray{"/images/products/sour-belt.jpg", "/images/products/sour-belt-2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  120000,
			LeadTime:       "15-18 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Apple", "Watermelon", "Blue Raspberry"},
			Shapes:         modelsCommon.StringArray{"Belt", "Strip"},
			Ingredients:    "Sugar, Glucose Syrup, Wheat Flour, Citric Acid, Malic Acid, Natural Flavors",
			Allergens:      "Contains wheat",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place below 25°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "jelly-fruits",
			Name:           "Fruit Jelly Cups",
			Summary:        "Soft jelly cups with real fruit pieces and natural juice",
			Description:    "Fruit Jelly Cups combine the smooth texture of premium jelly with real fruit pieces. Each cup delivers a burst of natural fruit flavor in a convenient single-serve format.",
			Category:       "Toffee Candy",
			CategorySlug:   "toffee-candy",
			Thumbnail:      "/images/products/jelly-fruits.jpg",
			Images:         modelsCommon.StringArray{"/images/products/jelly-fruits.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"},
			MOQ:            10000,
			StockQuantity:  100000,
			LeadTime:       "20-25 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Mango", "Lychee", "Strawberry", "Peach", "Grape"},
			Shapes:         modelsCommon.StringArray{"Cup", "Sachet"},
			Ingredients:    "Water, Sugar, Fruit Juice, Fructose, Konjac, Citric Acid, Natural Flavors",
			Allergens:      "None",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place below 25°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "chewy-toffee",
			Name:           "Creamy Toffee Chews",
			Summary:        "Rich and creamy toffee chews with buttery flavor",
			Description:    "Our Creamy Toffee Chews offer a luxurious candy experience with smooth, buttery texture and rich caramel flavor. Perfect for those who appreciate premium confectionery.",
			Category:       "Toffee Candy",
			CategorySlug:   "toffee-candy",
			Thumbnail:      "/images/products/chewy-toffee.jpg",
			Images:         modelsCommon.StringArray{"/images/products/chewy-toffee.jpg"},
			OEMAvailable:   true,
			HalalCertified: false,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  90000,
			LeadTime:       "18-22 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Classic Toffee", "Salted Caramel", "Coffee", "Chocolate"},
			Shapes:         modelsCommon.StringArray{"Square", "Rectangle", "Bar"},
			Ingredients:    "Sugar, Glucose Syrup, Condensed Milk, Butter, Salt, Natural Flavors",
			Allergens:      "Contains milk and dairy",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place below 20°C",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "gummy-bear-classic",
			Name:           "Classic Gummy Bears",
			Summary:        "Traditional gummy bears with authentic fruit flavors and perfect chewy texture",
			Description:    "Our Classic Gummy Bears are the gold standard of gummy candy. Made with real fruit juice and a time-tested recipe.",
			Category:       "Gummy Candy",
			CategorySlug:   "gummy-candy",
			Thumbnail:      "/images/products/gummy-bear-classic.jpg",
			Images:         modelsCommon.StringArray{"/images/products/gummy-bear-classic.jpg", "/images/products/gummy-variety_1.jpg", "/images/products/gummy-variety_2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"},
			MOQ:            5000,
			StockQuantity:  300000,
			LeadTime:       "15-20 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Orange", "Lemon", "Apple", "Grape"},
			Shapes:         modelsCommon.StringArray{"Bear"},
			Ingredients:    "Sugar, Glucose Syrup, Fruit Juice Concentrate, Gelatin, Citric Acid, Natural Flavors",
			Allergens:      "None",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "gummy-worms",
			Name:           "Sour Gummy Worms",
			Summary:        "Dual-flavored gummy worms with sweet and sour coating",
			Description:    "Our Sour Gummy Worms combine two complementary fruit flavors in every worm, topped with a sour sugar coating.",
			Category:       "Gummy Candy",
			CategorySlug:   "gummy-candy",
			Thumbnail:      "/images/products/gummy-worms.jpg",
			Images:         modelsCommon.StringArray{"/images/products/gummy-worms.jpg", "/images/products/gummy-shapes-variety_1.jpg", "/images/products/gummy-shapes-variety_2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  250000,
			LeadTime:       "15-20 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Cherry-Lemon", "Orange-Grape", "Strawberry-Green Apple", "Blue Raspberry-Watermelon"},
			Shapes:         modelsCommon.StringArray{"Worm"},
			Ingredients:    "Sugar, Glucose Syrup, Corn Syrup, Citric Acid, Malic Acid, Natural Flavors, Colors",
			Allergens:      "None",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "fruit-slices",
			Name:           "Citrus Fruit Slices",
			Summary:        "Sugar-dusted fruit slice candies with authentic citrus flavors",
			Description:    "Our Citrus Fruit Slices are coated in fine sugar with a soft chewy center.",
			Category:       "Gummy Candy",
			CategorySlug:   "gummy-candy",
			Thumbnail:      "/images/products/fruit-slices.jpg",
			Images:         modelsCommon.StringArray{"/images/products/fruit-slices.jpg", "/images/products/gummy-shapes-variety_3.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "Halal"},
			MOQ:            8000,
			StockQuantity:  160000,
			LeadTime:       "18-22 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Orange", "Lemon", "Lime", "Grapefruit"},
			Shapes:         modelsCommon.StringArray{"Slice", "Wedge"},
			Ingredients:    "Sugar, Corn Syrup, Citric Acid, Natural Fruit Flavors, Coconut Oil, Colors",
			Allergens:      "Contains coconut",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "hard-candy-assorted-mix",
			Name:           "Assorted Hard Candy Mix",
			Summary:        "Premium mixed fruit hard candies with vibrant colors",
			Description:    "Our Assorted Hard Candy Mix features a rainbow of fruit flavors, individually wrapped.",
			Category:       "Hard Candy",
			CategorySlug:   "hard-candy",
			Thumbnail:      "/images/products/hard-candy-mix.jpg",
			Images:         modelsCommon.StringArray{"/images/products/hard-candy-mix.jpg", "/images/products/hard-candy-variety_1.jpg", "/images/products/hard-candy-variety_2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"},
			MOQ:            10000,
			StockQuantity:  200000,
			LeadTime:       "20-25 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Orange", "Grape", "Apple", "Pineapple", "Mango"},
			Shapes:         modelsCommon.StringArray{"Round", "Oval", "Diamond"},
			Ingredients:    "Sugar, Glucose Syrup, Natural Flavors, Citric Acid, Natural Colors",
			Allergens:      "None",
			ShelfLife:      "24 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "hard-candy-fruit-bonbon",
			Name:           "Fruit Bonbon Hard Candy",
			Summary:        "Fruit-filled bonbon hard candies with soft center",
			Description:    "Fruit Bonbon Hard Candies feature a crisp outer shell with a delicious fruit-flavored soft center.",
			Category:       "Hard Candy",
			CategorySlug:   "hard-candy",
			Thumbnail:      "/images/products/hard-candy-fruits_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/hard-candy-fruits_1.jpg", "/images/products/hard-candy-fruits_2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            8000,
			StockQuantity:  150000,
			LeadTime:       "18-22 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Strawberry Cream", "Orange Cream", "Lemon Cream", "Blueberry Cream"},
			Shapes:         modelsCommon.StringArray{"Round", "Square"},
			Ingredients:    "Sugar, Glucose Syrup, Fruit Puree, Citric Acid, Natural Flavors",
			Allergens:      "None",
			ShelfLife:      "24 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "chocolate-premium-truffles",
			Name:           "Premium Chocolate Truffles",
			Summary:        "Luxury compound chocolate truffles with smooth centers",
			Description:    "Our Premium Chocolate Truffles are crafted using the finest compound chocolate.",
			Category:       "Compound Chocolate",
			CategorySlug:   "compound-chocolate",
			Thumbnail:      "/images/products/chocolate-premium_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/chocolate-premium_1.jpg", "/images/products/chocolate-premium_2.jpg", "/images/products/chocolate-premium_3.jpg"},
			OEMAvailable:   true,
			HalalCertified: false,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  100000,
			LeadTime:       "20-25 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Dark Chocolate", "Milk Chocolate", "White Chocolate", "Hazelnut"},
			Shapes:         modelsCommon.StringArray{"Round", "Heart", "Square"},
			Ingredients:    "Sugar, Cocoa Powder, Vegetable Fat, Milk Powder, Whey Powder, Soy Lecithin, Natural Vanilla Flavor",
			Allergens:      "Contains milk, soy. May contain traces of nuts",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "chocolate-variety-box",
			Name:           "Chocolate Variety Gift Box",
			Summary:        "Assorted compound chocolate gift box",
			Description:    "The Chocolate Variety Gift Box showcases our best compound chocolates in a premium presentation box.",
			Category:       "Compound Chocolate",
			CategorySlug:   "compound-chocolate",
			Thumbnail:      "/images/products/chocolate-variety_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/chocolate-variety_1.jpg", "/images/products/chocolate-variety_2.jpg", "/images/products/chocolate-variety_3.jpg"},
			OEMAvailable:   true,
			HalalCertified: false,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            3000,
			StockQuantity:  80000,
			LeadTime:       "22-28 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Dark Ganache", "Milk Caramel", "White Raspberry", "Hazelnut Praline"},
			Shapes:         modelsCommon.StringArray{"Assorted"},
			Ingredients:    "Sugar, Cocoa Powder, Vegetable Fat, Milk Powder, Soy Lecithin, Natural Flavors",
			Allergens:      "Contains milk, soy. May contain traces of nuts",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "licorice-black-twists",
			Name:           "Classic Black Licorice Twists",
			Summary:        "Traditional black licorice twists with anise flavor",
			Description:    "Our Classic Black Licorice Twists are made using traditional recipes with real licorice root extract.",
			Category:       "Licorice Candy",
			CategorySlug:   "licorice",
			Thumbnail:      "/images/products/licorice-black.jpg",
			Images:         modelsCommon.StringArray{"/images/products/licorice-black.jpg", "/images/products/licorice-variety_1.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"},
			MOQ:            5000,
			StockQuantity:  120000,
			LeadTime:       "18-22 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Classic Anise", "Salted", "Sweet"},
			Shapes:         modelsCommon.StringArray{"Twist", "Strip", "Custom"},
			Ingredients:    "Wheat Flour, Corn Syrup, Molasses, Licorice Extract, Anise Oil, Salt",
			Allergens:      "Contains wheat/gluten",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "licorice-fruit-ropes",
			Name:           "Fruity Licorice Ropes",
			Summary:        "Colorful fruity licorice ropes",
			Description:    "Fruity Licorice Ropes offer a softer, sweeter alternative to traditional licorice.",
			Category:       "Licorice Candy",
			CategorySlug:   "licorice",
			Thumbnail:      "/images/products/licorice-variety_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/licorice-variety_1.jpg", "/images/products/licorice-variety_2.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "Halal"},
			MOQ:            5000,
			StockQuantity:  100000,
			LeadTime:       "18-22 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Strawberry", "Cherry", "Raspberry", "Watermelon"},
			Shapes:         modelsCommon.StringArray{"Rope", "Twist"},
			Ingredients:    "Wheat Flour, Corn Syrup, Sugar, Citric Acid, Natural Flavors, Colors",
			Allergens:      "Contains wheat/gluten",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "sour-gummy-extreme",
			Name:           "Extreme Sour Gummy",
			Summary:        "Ultra-sour gummy candies",
			Description:    "Extreme Sour Gummy delivers the most intense sour experience with a special sour crystal blend.",
			Category:       "Sour Candies",
			CategorySlug:   "sour-candies",
			Thumbnail:      "/images/products/sour-extreme_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/sour-extreme_1.jpg", "/images/products/sour-extreme_2.jpg", "/images/products/sour-extreme_3.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  180000,
			LeadTime:       "15-20 days",
			Featured:       true,
			Flavors:        modelsCommon.StringArray{"Extreme Lemon", "Warhead Apple", "Sour Watermelon", "Blue Raspberry Blast"},
			Shapes:         modelsCommon.StringArray{"Bear", "Worm", "Belt"},
			Ingredients:    "Sugar, Glucose Syrup, Corn Syrup, Citric Acid, Malic Acid, Natural Flavors, Colors",
			Allergens:      "None",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "sour-belt-rainbow",
			Name:           "Rainbow Sour Belts",
			Summary:        "Multi-colored sour belts with tangy coating",
			Description:    "Rainbow Sour Belts feature vibrant multi-color strips with tangy fruit flavors.",
			Category:       "Sour Candies",
			CategorySlug:   "sour-candies",
			Thumbnail:      "/images/products/sour-variety_1.jpg",
			Images:         modelsCommon.StringArray{"/images/products/sour-variety_1.jpg", "/images/products/sour-variety_2.jpg", "/images/products/sour-variety_3.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "ISO 22000"},
			MOQ:            5000,
			StockQuantity:  140000,
			LeadTime:       "15-18 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Strawberry-Banana", "Blue Raspberry-Lemon", "Watermelon-Green Apple", "Grape-Orange"},
			Shapes:         modelsCommon.StringArray{"Belt", "Strip"},
			Ingredients:    "Sugar, Glucose Syrup, Wheat Flour, Citric Acid, Malic Acid, Natural Flavors",
			Allergens:      "Contains wheat",
			ShelfLife:      "12 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Slug:           "sour-gummy-bears-zing",
			Name:           "Sour Zing Gummy Bears",
			Summary:        "Sour-coated gummy bears",
			Description:    "Sour Zing Gummy Bears combine the classic gummy bear shape with an electrifying sour sugar coating.",
			Category:       "Sour Candies",
			CategorySlug:   "sour-candies",
			Thumbnail:      "/images/products/sour-gummy-bears.jpg",
			Images:         modelsCommon.StringArray{"/images/products/sour-gummy-bears.jpg"},
			OEMAvailable:   true,
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"HACCP", "Halal"},
			MOQ:            5000,
			StockQuantity:  170000,
			LeadTime:       "15-20 days",
			Featured:       false,
			Flavors:        modelsCommon.StringArray{"Sour Cherry", "Sour Lemon", "Sour Apple", "Sour Grape"},
			Shapes:         modelsCommon.StringArray{"Bear"},
			Ingredients:    "Sugar, Glucose Syrup, Fruit Juice, Gelatin, Citric Acid, Malic Acid, Natural Flavors",
			Allergens:      "None",
			ShelfLife:      "18 months",
			Storage:        "Store in cool, dry place",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, product := range products {
		if err := db.FirstOrCreate(&product, modelsProduct.Product{Slug: product.Slug}).Error; err != nil {
			return err
		}
	}
	// Ensure BasePrice is set even on pre-existing rows (FirstOrCreate skips field updates)
	basePrices := map[string]float64{
		"4d-fruit-gummy":             8.50,
		"crystal-hard-candy":         5.10,
		"rainbow-lollipop":           3.80,
		"sour-belt":                  4.20,
		"jelly-fruits":               6.00,
		"chewy-toffee":               7.50,
		"gummy-bear-classic":         4.95,
		"gummy-worms":                5.25,
		"fruit-slices":               4.50,
		"hard-candy-assorted-mix":    3.60,
		"hard-candy-fruit-bonbon":    4.80,
		"chocolate-premium-truffles": 12.00,
		"chocolate-variety-box":      15.00,
		"licorice-black-twists":      3.90,
		"licorice-fruit-ropes":       4.10,
		"sour-gummy-extreme":         5.50,
		"sour-belt-rainbow":          4.30,
		"sour-gummy-bears-zing":      5.00,
	}
	for slug, price := range basePrices {
		db.Model(&modelsProduct.Product{}).Where("slug = ? AND (base_price = 0 OR base_price IS NULL)", slug).Update("base_price", price)
	}

	return nil
}

// seedOEMFlows seeds OEM flow data
func seedOEMFlows(db *gorm.DB) error {
	flows := []modelsProduct.OEMFlow{
		{
			ID:          "quick-odm",
			Title:       "Quick ODM Service",
			Description: "Fast-track product development using our existing formulations with custom branding",
			Type:        "quick_odm",
			Steps: modelsProduct.OEMStepArray{
				modelsProduct.OEMStep{Order: 1, Title: "Product Selection", Description: "Choose from our catalog of proven formulations", Duration: "1-2 days", Deliverables: modelsCommon.StringArray{"Product samples", "Specification sheet"}},
				modelsProduct.OEMStep{Order: 2, Title: "Branding Design", Description: "Custom packaging and label design", Duration: "5-7 days", Deliverables: modelsCommon.StringArray{"Design mockups", "Artwork files"}},
				modelsProduct.OEMStep{Order: 3, Title: "Sample Approval", Description: "Review and approve final samples", Duration: "3-5 days", Deliverables: modelsCommon.StringArray{"Final samples", "Production specs"}},
				modelsProduct.OEMStep{Order: 4, Title: "Production", Description: "Manufacturing and quality control", Duration: "15-20 days", Deliverables: modelsCommon.StringArray{"Finished products", "QC reports"}},
			},
			Timeline:  "25-35 days",
			MOQ:       3000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "full-oem",
			Title:       "Full OEM Service",
			Description: "Complete custom product development from formulation to packaging",
			Type:        "full_oem",
			Steps: modelsProduct.OEMStepArray{
				modelsProduct.OEMStep{Order: 1, Title: "Requirement Analysis", Description: "Define product specifications and target market", Duration: "3-5 days", Deliverables: modelsCommon.StringArray{"Product brief", "Market analysis"}},
				modelsProduct.OEMStep{Order: 2, Title: "R&D Development", Description: "Formulation development and testing", Duration: "14-21 days", Deliverables: modelsCommon.StringArray{"Prototype samples", "Formulation specs"}},
				modelsProduct.OEMStep{Order: 3, Title: "Design & Packaging", Description: "Custom packaging design and engineering", Duration: "10-14 days", Deliverables: modelsCommon.StringArray{"Packaging design", "Mold specifications"}},
				modelsProduct.OEMStep{Order: 4, Title: "Compliance Testing", Description: "Quality and safety certification", Duration: "7-10 days", Deliverables: modelsCommon.StringArray{"Lab reports", "Certificates"}},
				modelsProduct.OEMStep{Order: 5, Title: "Production", Description: "Mass production and quality control", Duration: "20-30 days", Deliverables: modelsCommon.StringArray{"Finished products", "QC reports", "Documentation"}},
			},
			Timeline:  "60-90 days",
			MOQ:       10000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, flow := range flows {
		if err := db.FirstOrCreate(&flow, modelsProduct.OEMFlow{ID: flow.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedOEMSolutions seeds OEM solution data
func seedOEMSolutions(db *gorm.DB) error {
	solutions := []modelsProduct.OEMSolution{
		{
			ID:           utils.GenerateID(),
			Slug:         "private-label-gummy",
			Title:        "Private Label Gummy Manufacturing",
			Description:  "Complete private label solution for gummy candies including formulation, design, and production. Ideal for brands looking to enter the confectionery market quickly.",
			Thumbnail:    "/images/oem/private-label-gummy.jpg",
			Images:       modelsCommon.StringArray{"/images/oem/private-label-gummy-1.jpg", "/images/oem/private-label-gummy-2.jpg"},
			Category:     "Private Label",
			MOQ:          5000,
			Applications: modelsCommon.StringArray{"Retail Brands", "Promotional Items", "Gift Sets", "Seasonal Products"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID(),
			Slug:         "custom-shape-candy",
			Title:        "Custom Shape Candy Development",
			Description:  "Create unique candy shapes for your brand with our custom mold development service. Perfect for brand differentiation and special occasions.",
			Thumbnail:    "/images/oem/custom-shape-candy.jpg",
			Images:       modelsCommon.StringArray{"/images/oem/custom-shape-candy-1.jpg"},
			Category:     "Custom Development",
			MOQ:          10000,
			Applications: modelsCommon.StringArray{"Brand Mascots", "Logo Candies", "Event Souvenirs", "Holiday Specials"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID(),
			Slug:         "functional-candy",
			Title:        "Functional Candy Solutions",
			Description:  "Develop candies with added health benefits including vitamins, probiotics, and natural supplements. Meet the growing demand for healthier confectionery options.",
			Thumbnail:    "/images/oem/functional-candy.jpg",
			Images:       modelsCommon.StringArray{"/images/oem/functional-candy-1.jpg"},
			Category:     "Functional",
			MOQ:          8000,
			Applications: modelsCommon.StringArray{"Vitamin Gummies", "Immune Support", "Sleep Aid", "Energy Boost"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, solution := range solutions {
		if err := db.FirstOrCreate(&solution, modelsProduct.OEMSolution{Slug: solution.Slug}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedCertifications seeds certification data
func seedCertifications(db *gorm.DB) error {
	certifications := []modelsProduct.Certification{
		{
			ID:             utils.GenerateID(),
			Name:           "Hazard Analysis Critical Control Point",
			Abbreviation:   "HACCP",
			Description:    "International food safety management system ensuring product safety from raw materials to consumption",
			Issuer:         "SGS",
			ValidUntil:     "2027-12-31",
			CertificateURL: "/certificates/haccp.pdf",
			BadgeURL:       "/images/certifications/haccp-badge.png",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Name:           "ISO 22000 Food Safety Management",
			Abbreviation:   "ISO 22000",
			Description:    "International standard for food safety management systems covering all organizations in the food chain",
			Issuer:         "Bureau Veritas",
			ValidUntil:     "2027-12-31",
			CertificateURL: "/certificates/iso22000.pdf",
			BadgeURL:       "/images/certifications/iso22000-badge.png",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Name:           "Halal Certification",
			Abbreviation:   "Halal",
			Description:    "Certification confirming products comply with Islamic dietary laws and are suitable for Muslim consumers",
			Issuer:         "Halal Certification Authority",
			ValidUntil:     "2027-12-31",
			CertificateURL: "/certificates/halal.pdf",
			BadgeURL:       "/images/certifications/halal-badge.png",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Name:           "FDA Registration",
			Abbreviation:   "FDA",
			Description:    "Registration with the U.S. Food and Drug Administration for food facility compliance",
			Issuer:         "U.S. FDA",
			ValidUntil:     "2027-12-31",
			CertificateURL: "/certificates/fda.pdf",
			BadgeURL:       "/images/certifications/fda-badge.png",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             utils.GenerateID(),
			Name:           "BRC Food Safety",
			Abbreviation:   "BRC",
			Description:    "British Retail Consortium Global Standard for Food Safety certification",
			Issuer:         "BRCGS",
			ValidUntil:     "2027-12-31",
			CertificateURL: "/certificates/brc.pdf",
			BadgeURL:       "/images/certifications/brc-badge.png",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, cert := range certifications {
		if err := db.FirstOrCreate(&cert, modelsProduct.Certification{Abbreviation: cert.Abbreviation}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedFactoryInfo seeds factory information
func seedFactoryInfo(db *gorm.DB) error {
	info := modelsProduct.FactoryInfo{
		ID:          "1",
		Name:        "CandyPro Manufacturing",
		Founded:     2010,
		Description: "CandyPro Manufacturing is a leading OEM candy manufacturer with over 14 years of experience in creating premium confectionery products. Our state-of-the-art facility combines traditional candy-making craftsmanship with modern production technology to deliver exceptional products for brands worldwide.",
		Address:     "Industrial Zone, Guangdong Province, China",
		Capacity: modelsProduct.FactoryCapacity{
			Daily:   "50 tons",
			Monthly: "1,500 tons",
		},
		Area:      "25,000 sqm",
		Employees: 350,
		Markets:   modelsCommon.StringArray{"North America", "Europe", "Middle East", "Southeast Asia", "Australia"},
		Images:    modelsCommon.StringArray{"/images/factory/exterior.jpg", "/images/factory/production-line.jpg", "/images/factory/quality-lab.jpg"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return db.FirstOrCreate(&info, modelsProduct.FactoryInfo{ID: "1"}).Error
}

// seedBlogPosts seeds blog post data
func seedBlogPosts(db *gorm.DB) error {
	posts := []modelsProduct.BlogPost{
		{
			ID:           utils.GenerateID(),
			Slug:         "trends-candy-industry-2026",
			Title:        "Top 10 Candy Industry Trends for 2026",
			Excerpt:      "Discover the latest trends shaping the confectionery industry, from functional candies to sustainable packaging.",
			Content:      "The candy industry continues to evolve with changing consumer preferences and technological advancements. In 2026, we're seeing several key trends that are reshaping how manufacturers approach product development and marketing...",
			Category:     "Industry Trends",
			AuthorName:   "Sarah Chen",
			AuthorAvatar: "/images/team/sarah-chen.jpg",
			PublishedAt:  time.Now().AddDate(0, -1, 0),
			Thumbnail:    "/images/blog/trends-2024.jpg",
			ReadTime:     8,
			Tags:         modelsCommon.StringArray{"Trends", "Industry", "Innovation"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID(),
			Slug:         "guide-private-label-candy",
			Title:        "Complete Guide to Private Label Candy Manufacturing",
			Excerpt:      "Everything you need to know about launching your own candy brand through private label manufacturing.",
			Content:      "Private label candy manufacturing offers an excellent opportunity for businesses to enter the confectionery market without the complexity of developing their own formulations. This comprehensive guide covers everything from choosing the right manufacturer to understanding MOQs and lead times...",
			Category:     "Business Guide",
			AuthorName:   "Michael Wong",
			AuthorAvatar: "/images/team/michael-wong.jpg",
			PublishedAt:  time.Now().AddDate(0, -2, 0),
			Thumbnail:    "/images/blog/private-label-guide.jpg",
			ReadTime:     12,
			Tags:         modelsCommon.StringArray{"Private Label", "Business", "Manufacturing"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID(),
			Slug:         "halal-certification-candy",
			Title:        "Understanding Halal Certification for Candy Products",
			Excerpt:      "A comprehensive overview of halal certification requirements and processes for confectionery manufacturers.",
			Content:      "With the growing global Muslim population and increasing demand for halal-certified products, understanding halal certification has become essential for candy manufacturers targeting international markets...",
			Category:     "Certifications",
			AuthorName:   "Ahmad Hassan",
			AuthorAvatar: "/images/team/ahmad-hassan.jpg",
			PublishedAt:  time.Now().AddDate(0, -3, 0),
			Thumbnail:    "/images/blog/halal-certification.jpg",
			ReadTime:     6,
			Tags:         modelsCommon.StringArray{"Halal", "Certification", "Quality"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, post := range posts {
		if err := db.FirstOrCreate(&post, modelsProduct.BlogPost{Slug: post.Slug}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedCaseStudies seeds case study data
func seedCaseStudies(db *gorm.DB) error {
	cases := []modelsProduct.CaseStudy{
		{
			ID:        utils.GenerateID(),
			Slug:      "european-retailer-gummy-launch",
			Title:     "European Retailer Gummy Product Line Launch",
			Client:    "EuroCandy Mart",
			Industry:  "Retail",
			Location:  "Germany",
			Thumbnail: "/images/cases/euro-retailer.jpg",
			Images:    modelsCommon.StringArray{"/images/cases/euro-retailer-1.jpg", "/images/cases/euro-retailer-2.jpg"},
			Challenge: "EuroCandy Mart needed to develop a premium private label gummy line to compete with established brands while maintaining competitive pricing and meeting strict EU food safety standards.",
			Solution:  "We developed a custom gummy formulation using natural colors and flavors, created unique fruit shapes representing European fruits, and designed eco-friendly packaging that highlighted the natural ingredients.",
			Result:    "The product line launched in 500+ stores across Germany, France, and Benelux. First-year sales exceeded projections by 40%, and the line has since expanded to include seasonal varieties.",
			Timeline:  "4 months",
			Services:  modelsCommon.StringArray{"Product Development", "Packaging Design", "Compliance Support", "Production"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        utils.GenerateID(),
			Slug:      "functional-vitamin-gummies",
			Title:     "Functional Vitamin Gummy Brand Development",
			Client:    "NutriLife Supplements",
			Industry:  "Health & Wellness",
			Location:  "United States",
			Thumbnail: "/images/cases/vitamin-gummies.jpg",
			Images:    modelsCommon.StringArray{"/images/cases/vitamin-gummies-1.jpg"},
			Challenge: "NutriLife wanted to enter the growing functional gummy market with a line of vitamin-infused gummies that tasted great while delivering effective dosages of vitamins and minerals.",
			Solution:  "Our R&D team developed a sugar-free formulation using pectin instead of gelatin, incorporated targeted vitamin blends, and created a clean-label product with natural sweeteners.",
			Result:    "The product line achieved 300% growth in the first year, received positive reviews for taste, and helped establish NutriLife as a leader in the functional gummy segment.",
			Timeline:  "6 months",
			Services:  modelsCommon.StringArray{"R&D", "Formulation", "Regulatory Support", "Production"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        utils.GenerateID(),
			Slug:      "middle-east-halal-expansion",
			Title:     "Middle East Market Halal Candy Expansion",
			Client:    "Gulf Trading Company",
			Industry:  "Distribution",
			Location:  "UAE",
			Thumbnail: "/images/cases/middle-east-halal.jpg",
			Images:    modelsCommon.StringArray{"/images/cases/middle-east-halal-1.jpg"},
			Challenge: "Gulf Trading Company needed a diverse range of halal-certified candies to expand their product portfolio for the Middle East market, with specific requirements for halal compliance and Arabic packaging.",
			Solution:  "We developed a complete halal-certified candy range including gummies, hard candies, and lollipops with Arabic and English packaging, and obtained halal certification from recognized authorities.",
			Result:    "Successfully launched across 8 Middle Eastern countries with distribution in major retail chains. The halal certification and quality helped capture 15% market share in the first year.",
			Timeline:  "5 months",
			Services:  modelsCommon.StringArray{"Product Development", "Halal Certification", "Localization", "Production"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, caseStudy := range cases {
		if err := db.FirstOrCreate(&caseStudy, modelsProduct.CaseStudy{Slug: caseStudy.Slug}).Error; err != nil {
			return err
		}
	}
	return nil
}
