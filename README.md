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

## Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 15
- Docker & docker-compose (optional, for containerized setup)

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
    handlers/                  # HTTP layer (70 files)
    services/                  # Business logic (48 files)
    repository/                # Data access (39 files)
    models/                    # Database models (32 files)
    middleware/                 # Auth, CORS, rate limit, locale, active user
frontend/
  pages/
    admin/                     # Admin portal (26 pages)
    customer/                  # Customer portal (25 pages)
    auth/                      # Authentication (6 pages)
    products/, blog/, ...      # Public marketing pages (16 pages)
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

**Total: ~268 API endpoints, 77 frontend pages**

## Roles & Permissions

| Role | Access |
|------|--------|
| `customer` | Customer portal (order, inquire, trade) |
| `admin` | Admin portal (manage orders, products, etc.) |
| `superadmin` | Admin portal + user management |

User lifecycle: `pending` (register) → `active` (admin approves / KYB verified)

## Configuration

Copy the example files and edit as needed:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

Key variables you must set:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | — | Min 32 characters, use a strong random value |
| `DB_USER` / `DB_PASSWORD` | Yes | candypro | PostgreSQL credentials |
| `FRONTEND_URL` | Yes | http://localhost:3000 | CORS origin (frontend URL) |
| `OPENAI_API_KEY` | AI features | — | OpenAI-compatible API key |

All available options are documented in the example files:
`backend/.env.example` and `frontend/.env.example`.

## Troubleshooting

| Problem | Likely Cause | Fix |
|---------|-------------|-----|
| App fails to start with JWT error | `JWT_SECRET` not set or too short | Set `JWT_SECRET` to 32+ characters in `.env` |
| Database connection refused | PostgreSQL not running or wrong credentials | Check `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` |
| Frontend can't reach backend | `API_BASE_URL` mismatch | Ensure `frontend/.env` has `API_BASE_URL=http://localhost:8080/api` |
| CORS errors in browser | `FRONTEND_URL` doesn't match the frontend origin | Set `FRONTEND_URL=http://localhost:3000` in `backend/.env` |
| AI features return errors | `OPENAI_API_KEY` missing or invalid | Set a valid OpenAI-compatible API key |

## Notes

- Swagger UI at `/swagger/*` (toggle via `ENABLE_SWAGGER=true`)
- Health check at `/health`
- AI endpoints require valid `OPENAI_API_KEY`
- Order creation enforces destination-country compliance checks
- Background worker auto-cancels expired draft orders and restores inventory
- Default language is Chinese (zh); English available via language switcher
- File uploads served via `/uploads/*filepath` (payment proofs and KYB paths blocked)
- Customer portal write operations require active status (KYB gate)
