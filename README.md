# CandyPro OEM B2B Platform

Professional candy OEM (Original Equipment Manufacturing) B2B platform with admin portal, customer portal, public marketing site, and AI-powered trade assistance.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Nuxt 3.15, Vue 3.5, TypeScript 5.6, Tailwind CSS, @nuxtjs/i18n v9 |
| Backend | Go 1.24.7, Gin 1.10, GORM 1.30, PostgreSQL 15 |
| Auth | JWT (HS256), Cookie-based tokens |
| AI | Cloudwego Eino (OpenAI compatible), pgvector |
| i18n | Chinese (default) + English |
| Infra | Docker / docker-compose |

## Quick Start

```bash
make install          # Install all dependencies
make dev-all          # Start backend (:8080) + frontend (:3000)
```

## Project Structure

```
backend/
  cmd/api/main.go              # Entry point
  internal/
    api/                       # Router + routes (5 scopes)
    handlers/                  # HTTP layer (66 files)
    services/                  # Business logic (49 files)
    repository/                # Data access (38 files)
    models/                    # Database models (31 files)
    middleware/                 # Auth, CORS, rate limit, locale, active user
frontend/
  pages/
    admin/                     # Admin portal (26 pages)
    customer/                  # Customer portal (25 pages)
    auth/                      # Authentication (6 pages)
    products/, blog/, ...      # Public marketing pages (12 pages)
  components/                  # Auto-imported (pathPrefix: false)
  composables/                 # useApi, useAuth, useInquiry, etc.
  layouts/                     # admin, customer, auth, default
  i18n/                        # en/ + zh/ locale files
```

## Architecture

**Handler → Service → Repository** with 5 scoped route groups:

| Scope | Auth | Endpoints | Description |
|-------|------|-----------|-------------|
| `public` | None | 23 | Products, categories, OEM, factory, blog, search |
| `auth` | Mixed | 11 | Login, register, profile, password reset |
| `user` | JWT + customer | 54 | Customer portal (read open, write needs KYB) |
| `admin` | JWT + admin | 155 | Full admin management |
| `system` | Mixed | 7 | AI chatbot, search, recommendations |

**Total: ~253 API endpoints, 77 frontend pages**

## Roles & Permissions

| Role | Access |
|------|--------|
| `customer` | Customer portal (order, inquire, trade) |
| `admin` | Admin portal (manage orders, products, etc.) |
| `superadmin` | Admin portal + user management |

User lifecycle: `pending` (register) → `active` (admin approves / KYB verified)

## Configuration

### Backend (.env)
```
PORT=8080  DB_HOST=localhost  DB_PORT=5432
DB_USER=postgres  DB_PASSWORD=postgres  DB_NAME=candypro
JWT_SECRET=your-secret
OPENAI_API_KEY=sk-...
CORS_ORIGINS=http://localhost:3000
```

### Frontend (.env)
```
API_BASE_URL=http://localhost:8080/api/v1
```

## Notes

- Swagger UI at `/swagger/*` (toggle via `ENABLE_SWAGGER=true`)
- Health check at `/health`
- AI endpoints require valid `OPENAI_API_KEY`
- Order creation enforces destination-country compliance checks
- Background worker auto-cancels expired draft orders and restores inventory
- Default language is Chinese (zh); English available via language switcher
- File uploads served via `/uploads/*filepath` (payment proofs and KYB paths blocked)
- Customer portal write operations require active status (KYB gate)
