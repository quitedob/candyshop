# CandyPro OEM B2B Platform

Professional candy OEM (Original Equipment Manufacturing) B2B platform with admin portal, customer portal, public marketing site, and AI-powered trade assistance.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Nuxt 3.21, Vue 3.5, TypeScript 5.6, Tailwind CSS, @nuxtjs/i18n v10.3 |
| Backend | Go 1.24.7, Gin 1.10, GORM 1.30, PostgreSQL 15 |
| Auth | JWT (HS256), Cookie-based tokens |
| AI | Cloudwego Eino (OpenAI-compatible; default model `deepseek-v4-flash`), pgvector |
| i18n | Chinese (default) + English, 15 DB-backed namespaces |
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
    handlers/                  # HTTP layer (116 files)
    services/                  # Business logic (72 files)
    repository/                # Data access (57 files)
    models/                    # Database models (52 files)
    middleware/                 # Auth, CORS, rate limit, locale, active user
frontend/
  pages/
    admin/                     # Admin portal (41 pages)
    customer/                  # Customer portal (28 pages)
    auth/                      # Authentication (6 pages)
    products/, blog/, ...      # Public marketing pages (25 pages)
  components/                  # Auto-imported (pathPrefix: false), 42 components
  composables/                 # useApi, useAuth, useInquiry, etc. (24)
  layouts/                     # admin, customer, auth, default
  i18n/                        # en/ + zh/ — 15 namespaces, DB-backed translations
```

## Architecture

**Handler → Service → Repository** with 5 scoped route groups:

| Scope | Auth | Endpoints | Description |
|-------|------|-----------|-------------|
| `public` | None | 25 | Products, categories, OEM, factory, blog, search |
| `auth` | Mixed | 12 | Login, register, profile, password reset |
| `user` | JWT + customer | 94 | Customer portal (read open, write needs KYB) |
| `admin` | JWT + admin | 269 | Full admin management |
| `system` | Mixed | 16 | AI chatbot, search, recommendations, upload scan hook |

**Default: 416 business API endpoints; 421 routes including uploads, health, readiness, metrics and Swagger.**

The 2026-09-11 source inventory counts distinct HTTP method/path pairs with normal database-backed service wiring, including all eight translation handlers, and `ENABLE_SUPPLIER_PORTAL=false` / `ENABLE_MULTI_WAREHOUSE=false`. The route registrars contain 431 declarations: subtract 16 disabled routes and add the separately registered system scan hook to obtain 416. Enabling both feature flags gives 436 business endpoints. Disabling Swagger changes the overall route total only. Frontend page counts in the directory overview are historical and were not reverified in this inventory.

## Documentation & Changelog

- `docs/changelog/README.md` — index of fix/devlog entries (reverse-chronological); each entry links its source audit report and documents the fix approach, subagent usage, and verification
- `docs/changelog/drafts/` — per-subagent devlogs from the 2026-08-11 audit fix pass (per-bug fixer + adversarial-arguer workflow)
- `docs/reports/` — audit reports
- `docs/planning/` — `todo.md`, `plan.md`, `completed.md`, `future-problems.md` (backend & frontend risk register)

> Note: `docs/` is gitignored; the tracked changelog entries are force-added to git (`git add -f`).

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
