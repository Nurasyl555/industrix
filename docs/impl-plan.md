# Industrial Equipment Marketplace — Implementation Plan

> Modular Monolith · 7 phases · single backend + frontend · KZ-compliant

---

## Current State (Phase 1 — Complete ✅)

### Implemented Backend Modules

| Module          | Routes                                                                                                    | Contracts                        |
| --------------- | --------------------------------------------------------------------------------------------------------- | -------------------------------- |
| **Identity**    | `/auth/email/register`, `/auth/email/login`, `/auth/phone/login`, `/auth/phone/verify`, `/auth/oauth/google`, `/auth/refresh`, `/users/me` | `UserProvider`    |
| **Integrity**   | `/companies`, `/companies/:id`                                                                            | `CompanyProvider`                |
| **Marketplace** | `/reviews`, `/reviews/:entityID`, `/reviews/:entityID/reputation`                                         | —                                |
| **Catalog**     | `/catalog/categories`, `/catalog/equipment` (CRUD + filter)                                               | `EquipmentProvider`              |
| **Listing**     | `/listings` (browse), `/my-listings` (CRUD + publish/archive)                                             | `ListingProvider`                |
| **Deal**        | `/deals` (create/get/close), `/my-deals`                                                                  | —                                |

> **MVP scope note:** Catalog/Listing/Deal are implemented as a minimal
> end-to-end slice. Deliberately deferred vs. the full architecture.md vision:
> Catalog has no `/compare` endpoint and publishes no Kafka events yet; Listing
> has no `moderation` state (`draft → active → archived`); Deal is a simple
> inquiry (`inquiry → closed`), not the full negotiation state machine. Search
> is plain SQL filtering, not OpenSearch.

### Implemented Platform

| Component                | Status                                           |
| ------------------------ | ------------------------------------------------ |
| JWT auth middleware      | ✅ Direct in-process validation                  |
| Rate limiting middleware | ✅ Redis sliding window                          |
| Logging middleware       | ✅ Structured JSON with trace-id                 |
| Single Dockerfile        | ✅ Multi-stage, optimized layer caching          |
| Swagger docs             | ✅ Auto-generated at `/swagger/`                 |
| Migrations               | ✅ 7 migration files (users, auth_providers, companies, reviews, equipment, listings, deals) |

### Shared Packages (`pkg/`)

| Package                                         | Status                          |
| ----------------------------------------------- | ------------------------------- |
| `postgres` — connection pool + migration runner | ✅                              |
| `redis` — client with sentinel support          | ✅                              |
| `kafka` — producer + consumer wrapper           | ✅                              |
| `jwt` — claims, issue, parse                    | ✅                              |
| `logger` — zerolog structured logging           | ✅                              |
| `errors` — typed error codes                    | ✅                              |
| `mongo`, `minio`, `geo`                         | ✅ Available for future modules |

---

## Phase 2 — Core Marketplace (Weeks 4–8)

### New Module: Catalog

> `modules/catalog/`

| Task                                                      | Priority |
| --------------------------------------------------------- | -------- |
| Equipment CRUD (create, read, update, delete)             | MUST     |
| Category taxonomy with hierarchical structure             | MUST     |
| Dynamic technical attribute schemas per category type     | MUST     |
| Comparison endpoint (`GET /catalog/compare?ids=1,2,3`)    | SHOULD   |
| Kafka events: `equipment.created`, `.updated`, `.deleted` | MUST     |

### New Module: Listing

> `modules/listing/`

| Task                                                 | Priority |
| ---------------------------------------------------- | -------- |
| Ad lifecycle: draft → moderation → active → archived | MUST     |
| Pricing rules (fixed, negotiable, rental per-day)    | MUST     |
| View counters, listing stats                         | SHOULD   |
| Subscription plan limit enforcement                  | SHOULD   |

### New Module: Search

> `modules/search/`

| Task                                                | Priority |
| --------------------------------------------------- | -------- |
| OpenSearch integration — full-text + faceted search | MUST     |
| Geo-region filtering via 2GIS API                   | MUST     |
| Autocomplete suggestions                            | SHOULD   |
| Redis cache for hot queries (TTL 60s)               | SHOULD   |
| Kafka consumer for index sync                       | MUST     |

### Frontend Tasks

| Task                                                   | Priority |
| ------------------------------------------------------ | -------- |
| Equipment catalog page with grid/list view             | MUST     |
| Equipment detail page with media gallery               | MUST     |
| Advanced search with filters (category, price, region) | MUST     |
| Equipment comparison page                              | SHOULD   |
| Favorites/watchlist                                    | SHOULD   |

### DB Migrations

- `004_equipment.up.sql` — equipment, categories, attributes tables
- `005_listings.up.sql` — listings, pricing, stats tables

---

## Phase 3 — Transactions (Weeks 9–13)

### New Module: Deal

> `modules/deal/`

- Deal state machine: `inquiry → negotiation → confirmed → in_progress → completed → cancelled`
- Coordinates booking, payment, document modules
- Kafka events: `deal.status.changed`, `deal.completed`

### New Module: Payment

> `modules/payment/`

- Kaspi Pay, Halyk Bank, Uzcard/Humo integrations (CIS-local providers)
- Escrow hold/release pattern
- Invoice generation, transaction history
- Subscription billing for seller tariff plans

### New Module: Booking

> `modules/booking/`

- Calendar-based availability for rentals
- Hold pattern (Redis TTL) + confirm (PostgreSQL)
- Optimistic locking for conflict prevention

### DB Migrations

- `006_deals.up.sql`, `007_payments.up.sql`, `008_bookings.up.sql`

---

## Phase 4 — Communication (Weeks 14–17)

### New Module: Chat

> `modules/chat/`

- WebSocket connections via Fiber
- MongoDB message storage
- Read receipts, typing indicators
- File sharing via MinIO presigned URLs

### New Module: Notification

> `modules/notification/`

- Kafka consumer — fans out to multiple channels
- Push: FCM/APNs (transient tokens only)
- SMS: Beeline KZ / Kcell SMPP (KZ operators)
- Email: Self-hosted Postal SMTP
- In-app: MongoDB notification feed

---

## Phase 5 — Intelligence (Weeks 18–21)

### New Module: Media

> `modules/media/`

- Presigned URL generation for direct browser→MinIO uploads
- imgproxy integration for image transforms (self-hosted)
- Video thumbnail extraction

### New Module: Engagement

> `modules/engagement/`

- Favorites & watchlists with price drop alerts
- Price history tracking (Kafka consumer)
- Market benchmarks per category+region

---

## Phase 6 — Integrity & Monetization (Weeks 22–26)

### Extend: Integrity Module

- Dispute management (complaint filing, evidence upload, arbitration)
- Subscription & tariff management (free/basic/pro/enterprise)
- Audit log (immutable append-only trail, CIS legal compliance)

### New Module: Analytics

> `modules/analytics/`

- Event aggregation from all Kafka topics
- Seller dashboards: views, contact rate, deal conversion
- Admin dashboards: GMV, active listings, user growth

---

## Phase 7 — Hardening & Launch (Weeks 27–30)

| Task                                                        | Priority |
| ----------------------------------------------------------- | -------- |
| Load testing (k6) — target 10k concurrent users             | MUST     |
| Security audit — OWASP, JWT rotation, input sanitization    | MUST     |
| Kubernetes manifests with HPA per-module scaling            | MUST     |
| CI/CD pipeline: GitHub Actions → container registry → k8s   | MUST     |
| Monitoring dashboards: Grafana + Prometheus + Loki + Jaeger | SHOULD   |
| Documentation finalization: API docs, runbooks, ADRs        | SHOULD   |

---

## Module Addition Pattern

Every new module follows the same vertical slice structure:

```
modules/<name>/
├── module.go      # NewModule() — wires repo → svc → handler
├── types.go       # domain models + request/response DTOs
├── repository.go  # database access layer
├── service.go     # business logic + optional contract impl
└── handler.go     # HTTP handler + route registration
```

If a module needs data from another module, it receives a `contracts.*Provider` interface — never a direct package import.

---

_Total scope: ~339 tasks across 7 phases. Phase 1 complete. Each phase produces a deployable, testable increment._
