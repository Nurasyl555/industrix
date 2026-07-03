# Industrix — Folder Structure

> Modular Monolith · Single Go Binary · v2.0

```
industrix/
├── .github/
│   └── workflows/
│       ├── ci.yml                          # lint + test on PR
│       ├── cd-staging.yml                  # deploy to staging on merge to main
│       └── cd-prod.yml                     # deploy to prod on release tag
│
├── docker-compose.yml                      # full local stack (backend + infra)
├── docker-compose.infra.yml                # infra only (PG, Redis, Kafka, etc.)
├── docker-compose.override.yml             # local dev overrides (ports, env)
├── .env.example                            # all env vars documented, no secrets
├── Makefile                                # top-level dev commands
├── go.work                                 # Go workspace (single entry: ./backend)
├── buf.yaml                                # Buf proto config
├── buf.gen.yaml                            # Buf codegen config
│
├── frontend/                               # ── Next.js web app ──────────────────────
│   ├── Dockerfile
│   ├── next.config.ts
│   ├── package.json
│   └── src/
│       ├── app/                            # Next.js App Router
│       │   ├── (auth)/login, register
│       │   ├── (marketplace)/catalog, search, compare
│       │   ├── (deals)/deals
│       │   ├── (account)/dashboard, settings
│       │   └── layout.tsx
│       ├── components/                     # UI components
│       ├── lib/api/                        # typed API client (single backend)
│       ├── store/                          # Zustand global state
│       └── types/                          # shared TS types
│
├── backend/                                # ── Modular Monolith ──────────────────────
│   ├── Dockerfile                          # single multi-stage build
│   ├── go.mod                              # single module: github.com/industrix/backend
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                     # entry point — wires modules + middleware
│   │
│   ├── contracts/                          # ── Cross-module interfaces ───────────────
│   │   └── contracts.go                    # UserProvider, CompanyProvider
│   │
│   ├── modules/                            # ── Domain modules (vertical slices) ──────
│   │   ├── identity/                       # Auth + Profile
│   │   │   ├── module.go                   # entry: NewModule() wires repo→svc→handler
│   │   │   ├── types.go                    # User, RegisterRequest, LoginRequest, etc.
│   │   │   ├── repository.go              # PostgreSQL + Redis queries
│   │   │   ├── service.go                 # business logic + contracts.UserProvider
│   │   │   └── handler.go                 # public + protected HTTP routes
│   │   │
│   │   ├── integrity/                      # Company management & verification
│   │   │   ├── module.go
│   │   │   ├── types.go                    # Company, CompanyStatus, CreateCompanyReq
│   │   │   ├── repository.go
│   │   │   ├── service.go                 # + contracts.CompanyProvider
│   │   │   └── handler.go
│   │   │
│   │   ├── marketplace/                    # Reviews & reputation
│   │   │   ├── module.go
│   │   │   ├── types.go                    # Review, ReputationScore, CreateReviewReq
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   │
│   │   ├── catalog/                         # Equipment + categories
│   │   │   ├── module.go
│   │   │   ├── types.go                    # Equipment, Category, filters
│   │   │   ├── repository.go
│   │   │   ├── service.go                 # + contracts.EquipmentProvider
│   │   │   └── handler.go
│   │   │
│   │   ├── listing/                         # Sale/rental listings
│   │   │   ├── module.go                   # NewModule(pg, EquipmentProvider)
│   │   │   ├── types.go                    # Listing, ListingView, filters
│   │   │   ├── repository.go
│   │   │   ├── service.go                 # + contracts.ListingProvider
│   │   │   └── handler.go
│   │   │
│   │   └── deal/                            # Buyer inquiries (MVP)
│   │       ├── module.go                   # NewModule(pg, ListingProvider)
│   │       ├── types.go                    # Deal, DealView, CreateDealReq
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── handler.go
│   │
│   ├── platform/                           # ── Platform middleware ────────────────────
│   │   └── middleware/
│   │       ├── auth.go                     # JWT validation (direct, no gRPC)
│   │       ├── ratelimit.go               # Redis sliding window
│   │       └── logging.go                 # request logging with trace-id
│   │
│   ├── pkg/                                # ── Shared infrastructure packages ────────
│   │   ├── postgres/client.go             # PG connection pool + migration runner
│   │   ├── redis/client.go                # Redis client
│   │   ├── mongo/client.go                # MongoDB client
│   │   ├── kafka/producer.go, consumer.go # Kafka wrapper
│   │   ├── minio/client.go                # MinIO + presigned URLs
│   │   ├── jwt/jwt.go                     # JWT claims, issue, parse
│   │   ├── logger/zerolog.go              # structured JSON logging
│   │   ├── errors/errors.go              # typed error codes
│   │   └── geo/client.go                  # 2GIS geocoding
│   │
│   ├── migrations/                         # ── Database migrations ───────────────────
│   │   ├── 001_users.up.sql
│   │   ├── 002_auth_providers.up.sql
│   │   ├── 003_companies.up.sql
│   │   ├── 004_reviews.up.sql
│   │   ├── 005_equipment.up.sql
│   │   ├── 006_listings.up.sql
│   │   └── 007_deals.up.sql
│   │
│   ├── docs/                               # ── Swagger generated docs ────────────────
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   │
│   ├── gen/go/                             # ── Proto-generated code ──────────────────
│   │   ├── go.mod                          # separate module (generated, not hand-edited)
│   │   └── backend/proto/*/v1/*.pb.go
│   │
│   └── proto/                              # ── Protobuf definitions ──────────────────
│       └── */v1/*.proto
│
├── infra/                                  # ── Infrastructure configs ────────────────
│   ├── nginx/                              # NGINX config + SSL
│   ├── postgres/                           # DB init scripts
│   ├── kafka/                              # topic creation
│   ├── opensearch/                         # index mappings
│   ├── minio/                              # bucket creation
│   ├── grafana/                            # monitoring dashboards
│   ├── prometheus/                         # metrics scraping
│   └── loki/                               # log aggregation
│
├── scripts/                                # ── Automation scripts ────────────────────
│   ├── proto-gen.sh                        # regenerate proto → Go
│   ├── migrate.sh                          # run DB migrations
│   ├── seed.sh                             # seed dev data
│   └── healthcheck.sh                      # check containers
│
├── docs/                                   # ── Documentation ─────────────────────────
│   ├── architecture.md                     # system architecture (this doc)
│   ├── folder-tree.md                      # this file
│   ├── impl-plan.md                        # implementation plan
│   ├── PRD.md                              # product requirements
│   └── tz.md                               # technical spec (ТЗ)
│
└── README.md
```

## Key Structure Principles

### Vertical Module Isolation

Each module under `modules/` is a self-contained vertical slice:

- `types.go` — domain models and request DTOs
- `repository.go` — database access
- `service.go` — business logic + contract implementation
- `handler.go` — HTTP routes
- `module.go` — entry point that wires dependencies

### Cross-Module Communication

Modules never import each other. Communication happens through `contracts/`:

```
identity.Service ──implements──► contracts.UserProvider
integrity.Service ──implements──► contracts.CompanyProvider
catalog.Service  ──implements──► contracts.EquipmentProvider
listing.Service  ──implements──► contracts.ListingProvider

listing.Service ──consumes──► contracts.EquipmentProvider   (validate equipment ownership)
deal.Service    ──consumes──► contracts.ListingProvider     (validate listing is active)
```

### Single Binary Deployment

```yaml
# docker-compose.yml
backend:
  build: ./backend
  ports: ["8080:8080"]
  depends_on: [postgres, redis, kafka]
```

## Docker Compose Targets

| File                          | Purpose                                    |
| ----------------------------- | ------------------------------------------ |
| `docker-compose.yml`          | Full stack: backend + frontend + all infra |
| `docker-compose.infra.yml`    | Infra only: PG, Redis, Kafka, MinIO, etc.  |
| `docker-compose.override.yml` | Dev overrides: env vars, ports             |
