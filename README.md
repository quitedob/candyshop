# CandyPro OEM B2B Platform

Professional candy OEM (Original Equipment Manufacturing) B2B platform with admin portal, customer portal, public marketing site, and AI-powered trade assistance.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Nuxt 3.15, Vue 3.5, TypeScript, Tailwind CSS, @nuxtjs/i18n v9 |
| Backend | Go 1.24, Gin 1.10, GORM 1.30, PostgreSQL 15 |
| Auth | JWT (HS256), Cookie-based tokens |
| AI | Cloudwego Eino (OpenAI compatible) |
| i18n | Chinese (default) + English |
| Infra | Docker / docker-compose |

## Quick Start

```bash
make install          # Install all dependencies
make dev-all          # Start backend (:8080) + frontend (:3000)
```

See [CLAUDE.md](CLAUDE.md) for full command reference.

## Project Structure

```
backend/
  cmd/api/main.go              # Entry point
  internal/
    api/                       # Router + routes (5 scopes)
    handlers/                  # HTTP layer (57 files)
    services/                  # Business logic (37 files)
    repository/                # Data access (GORM)
    models/                    # Database models (20 files)
    middleware/                 # Auth, CORS, rate limit, security headers
frontend/
  pages/
    admin/                     # Admin portal (25 pages)
    customer/                  # Customer portal (23 pages)
    auth/                      # Authentication (5 pages)
    products/, blog/, ...      # Public marketing pages
  components/                  # Auto-imported (pathPrefix: false)
  composables/                 # useApi, useAuth, useInquiry, etc.
  layouts/                     # admin, customer, auth, default
  i18n/                        # en/ + zh/ locale files
docs/                          # Business domain documentation
```

## Architecture

**Handler → Service → Repository** with 5 scoped route groups:

| Scope | Auth | Endpoints | Description |
|-------|------|-----------|-------------|
| `public` | None | 23 | Products, categories, OEM, factory, blog, search |
| `auth` | Mixed | 11 | Login, register, profile, password reset |
| `user` | JWT + customer | 49 | Customer portal (read open, write needs KYB) |
| `admin` | JWT + admin | 80 | Full admin management |
| `system` | Mixed | 8 | AI chatbot, search, recommendations |

**Total: ~160 API endpoints, 72 frontend pages**

## Roles & Permissions

| Role | Access |
|------|--------|
| `customer` | Customer portal (order, inquire, trade) |
| `admin` | Admin portal (manage orders, products, etc.) |
| `superadmin` | Admin portal + user management |

User lifecycle: `pending` (register) → `active` (admin approves / KYB verified)

## Business Documentation

See [docs/](docs/) for detailed per-domain documentation:

| Domain | Document | Description |
|--------|----------|-------------|
| Auth & Users | [docs/01-auth.md](docs/01-auth.md) | Registration, login, JWT, KYB, user management |
| Products | [docs/02-products.md](docs/02-products.md) | Product catalog, categories, featured products |
| Orders | [docs/03-orders.md](docs/03-orders.md) | Order lifecycle, cart, checkout, payments |
| Inquiries | [docs/04-inquiries.md](docs/04-inquiries.md) | B2B inquiries, AI analysis, quotation |
| Trade | [docs/05-trade.md](docs/05-trade.md) | International trade, documents, compliance |
| Inventory | [docs/06-inventory.md](docs/06-inventory.md) | Stock management, reservation, adjustments |
| Pricing | [docs/07-pricing.md](docs/07-pricing.md) | Price lists, product pricing, customer tiers |
| OEM Projects | [docs/08-oem.md](docs/08-oem.md) | Custom manufacturing projects |
| Invoices | [docs/09-invoices.md](docs/09-invoices.md) | Proforma, commercial invoices, credit notes |
| Shipments | [docs/10-shipments.md](docs/10-shipments.md) | Shipping tracking, logistics |
| AI Services | [docs/11-ai.md](docs/11-ai.md) | Chatbot, recommendations, semantic search |
| Public Site | [docs/12-public-site.md](docs/12-public-site.md) | Marketing pages, SEO, content |

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
