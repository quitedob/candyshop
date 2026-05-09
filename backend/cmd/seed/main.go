package main

import (
	"flag"
	"log"
	"os"

	"candypro/api/internal/config"
	"candypro/api/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	all := flag.Bool("all", false, "Seed all data (business + demo)")
	business := flag.Bool("business", false, "Seed business data (products, categories, blog posts, etc.)")
	demo := flag.Bool("demo", false, "Seed demo workspace (demo users, orders, trades)")
	essential := flag.Bool("essential", false, "Seed essential system data (roles, superadmin)")
	flag.Parse()

	if !*all && !*business && !*demo && !*essential {
		*all = true
	}

	if *all || *essential {
		log.Println("Seeding essential system data...")
		if err := database.SeedEssential(db); err != nil {
			log.Fatalf("Essential seed failed: %v", err)
		}
	}

	if *all || *business {
		log.Println("Seeding business data...")
		if err := database.SeedDatabase(db); err != nil {
			log.Fatalf("Business seed failed: %v", err)
		}
		database.SeedCountryPaymentPolicies(db)
	}

	if *all || *demo {
		log.Println("Seeding demo workspace...")
		os.Setenv("SEED_DEMO_DATA", "true")
		if err := database.SeedDemoWorkspace(db); err != nil {
			log.Fatalf("Demo seed failed: %v", err)
		}
	}

	log.Println("Seeding completed successfully")
}
