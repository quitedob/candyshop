.PHONY: all help dev build test clean docker-up docker-down docker-logs migrate seed

# Default target
all: help

# Help target
help:
	@echo "CandyPro OEM - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev           Start backend development server"
	@echo "  dev-frontend  Start frontend development server"
	@echo "  dev-all       Start both backend and frontend"
	@echo ""
	@echo "Building:"
	@echo "  build         Build backend binary"
	@echo "  build-frontend Build frontend for production"
	@echo "  build-all     Build both backend and frontend"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run backend tests"
	@echo "  test-coverage Run backend tests with coverage"
	@echo "  test-frontend Run frontend tests"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up     Start all services with Docker Compose"
	@echo "  docker-down   Stop all Docker services"
	@echo "  docker-logs   View Docker logs"
	@echo "  docker-build  Rebuild Docker images"
	@echo ""
	@echo "Database:"
	@echo "  migrate       Run database migrations"
	@echo "  seed          Seed database with sample data"
	@echo "  reset-db      Reset database (drop and recreate)"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint          Run linter"
	@echo "  fmt           Format code"
	@echo "  vet           Run go vet"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         Clean build artifacts"
	@echo "  swagger       Generate Swagger documentation"
	@echo "  install       Install dependencies"

# Development targets
dev:
	cd backend && go run cmd/api/main.go

dev-frontend:
	cd frontend && npm run dev

dev-all:
	@echo "Starting backend and frontend..."
	@make -j2 dev dev-frontend

# Build targets
build:
	cd backend && go build -o bin/api cmd/api/main.go

build-frontend:
	cd frontend && npm run build

build-all: build build-frontend

# Test targets
test:
	cd backend && go test -v ./...

test-coverage:
	cd backend && go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

test-frontend:
	cd frontend && npm run test

# Docker targets
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build --no-cache

docker-restart:
	docker-compose restart

# Database targets
migrate:
	@echo "Running database migrations..."
	cd backend && go run cmd/api/main.go -migrate

seed:
	@echo "Seeding database..."
	cd backend && go run cmd/api/main.go -seed

reset-db:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	@sleep 5
	@make migrate
	@make seed

# Code quality targets
lint:
	cd backend && golangci-lint run

fmt:
	cd backend && go fmt ./...

vet:
	cd backend && go vet ./...

# Utility targets
clean:
	rm -rf backend/bin
	rm -rf frontend/.nuxt
	rm -rf frontend/.output
	rm -rf frontend/node_modules/.cache

swagger:
	cd backend && swag init -g cmd/api/main.go -o docs

install:
	cd backend && go mod download
	cd frontend && npm install

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Production targets
prod-build: build-all
	@echo "Production build complete"

prod-start:
	docker-compose -f docker-compose.prod.yml up -d

prod-stop:
	docker-compose -f docker-compose.prod.yml down
