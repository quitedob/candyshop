# CandyPro OEM Website

This repository contains a Go + Nuxt B2B candy OEM platform.
This README is aligned with the current implementation state.

## 1. Current Status (2026-03-09)

- API prefix is unified to `/api/v1`
- Auth has independent entry points:
  - Backend: `/api/v1/auth/*`
  - Frontend: `/auth/login`, `/auth/register`, `/auth/forgot-password`, `/auth/reset-password`
- customer/admin routes now cover the defined API set in router
- Trade ownership risk fixed:
  - trade `user_id` is string-based end to end
  - trade detail enforces ownership check
- Backend verification passed:
  - `cd backend && GOCACHE=./.gocache go test ./...`

## 2. Tech Stack

- Backend: Go 1.21+, Gin, GORM, PostgreSQL
- Frontend: Nuxt 3, Vue 3, TypeScript, Tailwind CSS
- Infrastructure: Docker / docker-compose

## 3. Structure

```text
backend/                     Go API service
  cmd/api/main.go            service entry
  http/router.go             HTTP route registration
  internal/                  handlers/services/repository/models
frontend/                    Nuxt frontend
  pages/                     page routes
  layouts/                   layouts (auth/customer/admin)
docker-compose.yml           local compose config
PROJECT_SPECIFICATION.md     current-state project specification
```

## 4. Quick Start

### 4.1 Local

```bash
# backend
cd backend
go mod download
go run cmd/api/main.go

# frontend
cd frontend
npm install
npm run dev
```

### 4.2 Docker

```bash
docker-compose up --build
```

Frontend API env is now unified to:
`API_BASE_URL=http://api:8080/api/v1`

## 5. API Coverage (router.go source of truth)

### 5.1 Public

- `GET /health`
- `GET /api/v1/products`
- `GET /api/v1/products/featured`
- `GET /api/v1/products/:slug`
- `GET /api/v1/products/:slug/related`
- `GET /api/v1/categories`
- `GET /api/v1/categories/:slug`
- `GET /api/v1/oem/flows`
- `GET /api/v1/oem/flows/:id`
- `GET /api/v1/oem/solutions`
- `GET /api/v1/oem/solutions/:slug`
- `GET /api/v1/factory`
- `GET /api/v1/factory/quality-controls`
- `GET /api/v1/factory/timeline`
- `GET /api/v1/certifications`
- `GET /api/v1/certifications/:id`
- `GET /api/v1/posts`
- `GET /api/v1/posts/:slug`
- `GET /api/v1/posts/:slug/related`
- `GET /api/v1/cases`
- `GET /api/v1/cases/:slug`
- `POST /api/v1/inquiry`
- `GET /api/v1/search`

### 5.2 Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/logout` (auth required)
- `GET /api/v1/auth/me` (auth required)
- `PUT /api/v1/auth/profile` (auth required)
- `PUT /api/v1/auth/change-password` (auth required)

### 5.3 Customer (auth + role: customer/admin/superadmin)

- `GET /api/v1/customer/dashboard`
- `GET /api/v1/customer/orders`
- `POST /api/v1/customer/orders`
- `POST /api/v1/customer/orders/ai-assist`
- `GET /api/v1/customer/orders/:id`
- `POST /api/v1/customer/orders/:id/confirm`
- `GET /api/v1/customer/orders/:id/progress`
- `PUT /api/v1/customer/profile`
- `POST /api/v1/customer/change-password`
- `GET /api/v1/customer/inquiries`
- `POST /api/v1/customer/inquiries`
- `GET /api/v1/customer/inquiries/:id`
- `PUT /api/v1/customer/inquiries/:id`
- `GET /api/v1/customer/quotes`
- `GET /api/v1/customer/notifications`
- `GET /api/v1/customer/trades`
- `POST /api/v1/customer/trades`
- `GET /api/v1/customer/trades/:id`
- `GET /api/v1/customer/ai/stream`

### 5.4 Admin (auth + role: admin/superadmin)

- `GET /api/v1/admin/dashboard`
- `GET /api/v1/admin/dashboard/stats`
- `GET /api/v1/admin/reports/sales`
- `GET /api/v1/admin/reports/inquiries`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/:id`
- `POST /api/v1/admin/users`
- `PUT /api/v1/admin/users/:id`
- `DELETE /api/v1/admin/users/:id`
- `PUT /api/v1/admin/users/:id/status`
- `PUT /api/v1/admin/users/:id/role`
- `GET /api/v1/admin/orders`
- `POST /api/v1/admin/orders`
- `GET /api/v1/admin/orders/:id`
- `PUT /api/v1/admin/orders/:id`
- `DELETE /api/v1/admin/orders/:id`
- `PUT /api/v1/admin/orders/:id/status`
- `GET /api/v1/admin/inquiries`
- `POST /api/v1/admin/inquiries`
- `GET /api/v1/admin/inquiries/:id`
- `PUT /api/v1/admin/inquiries/:id`
- `DELETE /api/v1/admin/inquiries/:id`
- `PUT /api/v1/admin/inquiries/:id/status`
- `PUT /api/v1/admin/inquiries/:id/assign`
- `POST /api/v1/admin/inquiries/:id/analyze`
- `POST /api/v1/admin/inquiries/:id/quote`
- `GET /api/v1/admin/products`
- `POST /api/v1/admin/products`
- `GET /api/v1/admin/products/:id`
- `PUT /api/v1/admin/products/:id`
- `DELETE /api/v1/admin/products/:id`
- `PUT /api/v1/admin/products/:id/status`
- `GET /api/v1/admin/content`
- `GET /api/v1/admin/content/:id`
- `POST /api/v1/admin/content`
- `PUT /api/v1/admin/content/:id`
- `DELETE /api/v1/admin/content/:id`

### 5.5 AI

- `POST /api/v1/ai/chatbot`
- `POST /api/v1/ai/recommend-products`
- `POST /api/v1/ai/search`
- `POST /api/v1/ai/analyze-inquiry` (auth required)
- `POST /api/v1/ai/generate-quotation` (auth required)
- `POST /api/v1/ai/translate` (auth required)
- `GET /api/v1/ai/conversations/:id` (auth required)

## 6. Frontend Domain Split

- Public: `/`, `/products/*`, `/blog/*`, etc.
- Auth: `/auth/*`
- Customer: `/customer/*`
- Admin: `/admin/*`

`frontend/app.vue` now conditionally disables marketing global chrome
(Header/Footer/Floating) for auth/customer/admin domains.

## 7. Notes

- `backend/docs/openapi.yaml` server URLs are set to `/api/v1`
- `backend/cmd/api/main.go` Swagger `@BasePath` is `/api/v1`
- AI endpoints execute real request validation and service logic; runtime model generation requires valid AI configuration (`OPENAI_API_KEY` etc.)
- Customer order creation (`POST /customer/orders` and `POST /customer/orders/ai-assist`) enforces destination-country compliance checks and may return `422 compliance_violation` with detailed violations/warnings.
- AI draft orders (`pending_confirmation`) reserve inventory immediately; a background cleanup worker auto-cancels expired drafts and restores stock (`ORDER_DRAFT_CLEANUP_ENABLED`, `ORDER_DRAFT_CLEANUP_INTERVAL_MINUTES`, `ORDER_DRAFT_EXPIRE_MINUTES`, `ORDER_DRAFT_CLEANUP_BATCH_SIZE`).
- AI compliance retrieval now prioritizes official-domain evidence in local corpus and returns per-reference `official/url` flags for frontend audit display.
- Initial superadmin seeding requires explicit bootstrap credentials:
  - `SEED_SUPERADMIN_PASSWORD_HASH` (preferred, bcrypt hash) or
  - `SEED_SUPERADMIN_PASSWORD` (plain password, hashed at seed time)
