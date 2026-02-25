Root 
├── .github/
│   └── workflows/
│       ├── ci.yml                          # lint + test all services on PR
│       ├── cd-staging.yml                  # deploy to staging on merge to main
│       └── cd-prod.yml                     # deploy to prod on release tag
│
├── docker-compose.yml                      # full local stack (all services + infra)
├── docker-compose.infra.yml                # infra only (PG, Mongo, Redis, Kafka, etc.)
├── docker-compose.override.yml             # local dev overrides (hot reload, ports)
├── .env.example                            # all env vars documented, no secrets
├── Makefile                                # top-level dev commands (see below)
│
├── frontend/                               # ── Next.js web app ──────────────────────
│   ├── Dockerfile
│   ├── next.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   ├── tailwind.config.ts
│   ├── public/
│   └── src/
│       ├── app/                            # Next.js App Router
│       │   ├── (auth)/
│       │   │   ├── login/page.tsx
│       │   │   └── register/page.tsx
│       │   ├── (marketplace)/
│       │   │   ├── catalog/
│       │   │   │   ├── page.tsx            # equipment listing
│       │   │   │   └── [id]/page.tsx       # equipment detail
│       │   │   ├── search/page.tsx
│       │   │   ├── compare/page.tsx
│       │   │   └── favorites/page.tsx
│       │   ├── (deals)/
│       │   │   ├── deals/page.tsx
│       │   │   └── deals/[id]/page.tsx
│       │   ├── (account)/
│       │   │   ├── dashboard/page.tsx
│       │   │   ├── listings/page.tsx
│       │   │   ├── analytics/page.tsx
│       │   │   └── settings/page.tsx
│       │   ├── (chat)/
│       │   │   └── chat/[dealId]/page.tsx
│       │   ├── (admin)/
│       │   │   └── admin/                  # admin panel (separate layout)
│       │   └── layout.tsx
│       ├── components/
│       │   ├── ui/                         # base design system (Button, Input, etc.)
│       │   ├── equipment/                  # EquipmentCard, SpecTable, MediaGallery
│       │   ├── chat/                       # ChatWindow, MessageBubble, FileUpload
│       │   ├── deal/                       # DealCard, StatusBadge, ContractViewer
│       │   └── layout/                     # Header, Footer, Sidebar, Nav
│       ├── lib/
│       │   ├── api/                        # typed API client (fetch wrappers per service)
│       │   │   ├── identity.ts
│       │   │   ├── catalog.ts
│       │   │   ├── search.ts
│       │   │   ├── deal.ts
│       │   │   └── ...
│       │   ├── ws/                         # WebSocket client for chat
│       │   ├── hooks/                      # useSearch, useChat, useDeal, etc.
│       │   └── utils/
│       ├── store/                          # Zustand global state
│       └── types/                          # shared TS types (mirrored from proto)
│
├── backend/                                # ── Backend services & shared code ────────
│   │
│   ├── services/                           # Go microservices
│   │   ├── gateway/                        # API Gateway (Go · Fiber)
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── middleware/
│   │   │   │   │   ├── auth.go             # JWT validation
│   │   │   │   │   ├── ratelimit.go        # Redis sliding window
│   │   │   │   │   ├── tracing.go          # inject trace-id
│   │   │   │   │   └── logging.go
│   │   │   │   ├── proxy/                  # route → downstream service
│   │   │   │   └── config/
│   │   │   └── config.yaml
│   │   │
│   │   ├── identity/                       # Auth + Profile + Company + Verification
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── auth/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── jwt.go
│   │   │   │   ├── profile/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── company/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── verification.go     # state machine
│   │   │   │   └── grpc/
│   │   │   │       └── server.go           # GetUser, GetCompany gRPC endpoints
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── catalog/                        # Equipment catalog + comparison endpoint
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── equipment/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── category/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── attributes/             # dynamic schema per category
│   │   │   │   │   └── schema.go
│   │   │   │   ├── comparison/
│   │   │   │   │   └── handler.go          # GET /catalog/compare — no separate svc
│   │   │   │   └── grpc/
│   │   │   │       └── server.go
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── listing/                        # Ad lifecycle + inventory management
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── listing/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── statemachine.go
│   │   │   │   └── stats/
│   │   │   │       └── handler.go          # view counters, listing analytics
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── search/                         # OpenSearch wrapper + recommendations
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── search/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── indexer.go          # Kafka consumer → OpenSearch
│   │   │   │   ├── suggest/
│   │   │   │   │   └── handler.go          # autocomplete
│   │   │   │   └── geo/                    # pkg/geo — 2GIS wrapper, region table
│   │   │   │       ├── regions.go
│   │   │   │       └── twogis.go
│   │   │   └── config.yaml
│   │   │
│   │   ├── booking/                        # Rental availability + slot reservation
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── availability/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── calendar.go
│   │   │   │   └── reservation/
│   │   │   │       ├── service.go          # hold (Redis TTL) + confirm (PG)
│   │   │   │       └── repository.go
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── deal/                           # Transaction orchestrator
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── deal/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── statemachine.go     # inquiry→negotiation→confirmed→...
│   │   │   │   └── grpc/
│   │   │   │       └── client.go           # calls Payment, Booking, Document
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── payment/                        # Kaspi/Halyk/Uzcard + escrow + invoices
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── payment/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── escrow/
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── invoice/
│   │   │   │   │   └── service.go
│   │   │   │   ├── providers/              # payment provider adapters
│   │   │   │   │   ├── kaspi.go
│   │   │   │   │   ├── halyk.go
│   │   │   │   │   └── uzcard.go
│   │   │   │   └── webhook/
│   │   │   │       └── handler.go          # inbound provider callbacks
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── document/                       # PDF gen (Gotenberg) + ЭДО + storage
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── document/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   ├── templates/              # Go HTML templates for contracts
│   │   │   │   │   ├── contract_sale.html
│   │   │   │   │   ├── contract_rental.html
│   │   │   │   │   └── invoice.html
│   │   │   │   ├── pdf/
│   │   │   │   │   └── gotenberg.go        # Gotenberg HTTP client
│   │   │   │   └── edo/                    # ЭДО integration adapter
│   │   │   │       └── client.go
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── review/                         # Post-deal ratings
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   └── review/
│   │   │   │       ├── handler.go
│   │   │   │       ├── service.go
│   │   │   │       └── repository.go
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── services-marketplace/           # Addl services + logistics (TZ §2.8)
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── catalog/                # service offerings by providers
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   └── logistics/
│   │   │   │       ├── handler.go
│   │   │   │       ├── service.go          # carrier matching, status tracking
│   │   │   │       └── routing.go          # 2GIS routing API calls
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── chat/                           # WebSocket full-featured messaging
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── ws/
│   │   │   │   │   ├── hub.go              # connection hub, presence
│   │   │   │   │   ├── client.go           # per-connection handler
│   │   │   │   │   └── handler.go          # HTTP upgrade endpoint
│   │   │   │   ├── message/
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go       # MongoDB
│   │   │   │   └── conversation/
│   │   │   │       ├── service.go
│   │   │   │       └── repository.go       # MongoDB
│   │   │   └── config.yaml
│   │   │
│   │   ├── notification/                   # Kafka consumer → FCM/SMPP/Postal/in-app
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── consumer/
│   │   │   │   │   └── kafka.go            # topic subscriptions
│   │   │   │   ├── dispatcher/
│   │   │   │   │   └── router.go           # route event → channel
│   │   │   │   ├── channels/
│   │   │   │   │   ├── push.go             # FCM/APNs
│   │   │   │   │   ├── sms.go              # Beeline KZ / Kcell SMPP
│   │   │   │   │   ├── email.go            # Postal SMTP
│   │   │   │   │   └── inapp.go            # MongoDB feed
│   │   │   │   └── templates/
│   │   │   │       ├── ru/                 # Russian templates
│   │   │   │       └── kk/                 # Kazakh templates
│   │   │   └── config.yaml
│   │   │
│   │   ├── engagement/                     # Favorites + price history
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── favorites/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   └── repository.go
│   │   │   │   └── pricehistory/
│   │   │   │       ├── consumer.go         # Kafka: equipment.updated
│   │   │   │       ├── service.go
│   │   │   │       └── repository.go
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── integrity/                      # Dispute + Subscription + Audit log
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── dispute/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── statemachine.go
│   │   │   │   ├── subscription/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   ├── service.go
│   │   │   │   │   ├── repository.go
│   │   │   │   │   └── grpc_server.go      # feature-flag check for Listing
│   │   │   │   └── audit/
│   │   │   │       ├── consumer.go         # Kafka: all sensitive topics
│   │   │   │       └── repository.go       # write-once PG
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   ├── media/                          # Upload handling + imgproxy + MinIO
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── upload/
│   │   │   │   │   ├── handler.go          # presigned URL generation
│   │   │   │   │   └── service.go
│   │   │   │   └── process/
│   │   │   │       ├── consumer.go         # Kafka: media.uploaded
│   │   │   │       └── imgproxy.go         # imgproxy HTTP client
│   │   │   └── config.yaml
│   │   │
│   │   ├── analytics/                      # Event aggregation + seller/admin dashboards
│   │   │   ├── Dockerfile
│   │   │   ├── cmd/main.go
│   │   │   ├── internal/
│   │   │   │   ├── consumer/
│   │   │   │   │   └── kafka.go
│   │   │   │   ├── aggregator/
│   │   │   │   │   ├── platform.go         # GMV, users, listings
│   │   │   │   │   └── seller.go           # per-seller metrics
│   │   │   │   └── api/
│   │   │   │       └── handler.go          # REST endpoints for dashboards
│   │   │   ├── migrations/
│   │   │   └── config.yaml
│   │   │
│   │   └── admin/                          # Moderation + category management
│   │       ├── Dockerfile
│   │       ├── cmd/main.go
│   │       ├── internal/
│   │       │   ├── moderation/
│   │       │   │   ├── handler.go
│   │       │   │   └── service.go          # calls other services via gRPC
│   │       │   └── content/
│   │       │       ├── handler.go          # category/attribute management
│   │       │       └── service.go
│   │       └── config.yaml
│   │
│   ├── proto/                              # Protobuf definitions (shared)
│   │   ├── identity/
│   │   │   └── v1/
│   │   │       └── identity.proto          # GetUser, GetCompany, VerifyToken RPCs
│   │   ├── catalog/
│   │   │   └── v1/
│   │   │       └── catalog.proto           # GetEquipment, ValidateAttributes RPCs
│   │   ├── payment/
│   │   │   └── v1/
│   │   │       └── payment.proto           # InitiateEscrow, ReleaseEscrow RPCs
│   │   ├── booking/
│   │   │   └── v1/
│   │   │       └── booking.proto           # CreateHold, ConfirmBooking RPCs
│   │   ├── integrity/
│   │   │   └── v1/
│   │   │       └── integrity.proto         # CheckPlanLimits, GetSubscription RPCs
│   │   └── gen/                            # generated Go + TS code (committed)
│   │       ├── go/
│   │       └── ts/
│   │
│   ├── pkg/                                # Shared Go packages
│   │   ├── kafka/
│   │   │   ├── producer.go                 # shared Kafka producer wrapper
│   │   │   └── consumer.go                 # shared consumer with retry logic
│   │   ├── redis/
│   │   │   └── client.go                   # shared Redis client setup
│   │   ├── postgres/
│   │   │   └── client.go                   # shared PG connection + migrations runner
│   │   ├── mongo/
│   │   │   └── client.go
│   │   ├── minio/
│   │   │   └── client.go                   # presigned URL helpers
│   │   ├── jwt/
│   │   │   └── claims.go                   # shared JWT claims struct
│   │   ├── logger/
│   │   │   └── zerolog.go                  # zerolog setup with trace-id injection
│   │   ├── tracer/
│   │   │   └── jaeger.go                   # OpenTelemetry / Jaeger setup
│   │   ├── errors/
│   │   │   └── errors.go                   # typed error codes shared across services
│   │   └── geo/
│   │       ├── regions.go                  # KZ/CIS region lookup table
│   │       └── twogis.go                   # 2GIS API client
│   │
│   ├── migrations/                         # DB migrations (per service)
│   │   ├── identity/
│   │   │   ├── 001_init.sql
│   │   │   └── 002_add_company_docs.sql
│   │   ├── catalog/
│   │   ├── listing/
│   │   ├── booking/
│   │   ├── deal/
│   │   ├── payment/
│   │   ├── document/
│   │   ├── review/
│   │   ├── services-marketplace/
│   │   ├── engagement/
│   │   ├── integrity/
│   │   ├── media/
│   │   └── analytics/
│   │
│   └── scripts/
│       ├── proto-gen.sh                    # regenerate proto → Go + TS
│       ├── migrate.sh                      # run migrations for all services
│       ├── seed.sh                         # seed dev data
│       └── healthcheck.sh                  # check all containers up
│
├── infra/                                  # ── Infrastructure configs ───────────────
│   ├── nginx/
│   │   ├── nginx.conf                      # main config
│   │   ├── conf.d/
│   │   │   ├── gateway.conf                # proxy to API gateway
│   │   │   ├── frontend.conf               # Next.js SSR
│   │   │   ├── minio.conf                  # MinIO public bucket proxy
│   │   │   └── ws.conf                     # WebSocket upgrade for chat
│   │   └── ssl/                            # Let's Encrypt certs (gitignored)
│   ├── postgres/
│   │   └── init/
│   │       └── 00_create_databases.sql     # CREATE DATABASE per service
│   ├── kafka/
│   │   └── topics.sh                       # topic creation script (run on first start)
│   ├── opensearch/
│   │   └── mappings/
│   │       └── equipment.json              # index mapping with dynamic templates
│   ├── minio/
│   │   └── buckets.sh                      # bucket + policy creation on first start
│   ├── gotenberg/                          # Gotenberg (self-hosted PDF)
│   │   └── Dockerfile                      # or use official image directly
│   ├── imgproxy/
│   │   └── .env.imgproxy                   # imgproxy config (signing key, formats)
│   ├── postal/                             # self-hosted SMTP
│   │   └── postal.yml
│   └── grafana/
│       └── dashboards/
│           ├── platform-overview.json
│           └── per-service.json
│
├── docs/
│   ├── api/                                # OpenAPI specs per service
│   │   ├── gateway.yaml
│   │   ├── catalog.yaml
│   │   └── ...
│   ├── adr/                                # Architecture Decision Records
│   │   ├── 001-monorepo.md
│   │   ├── 002-kafka-over-redis-streams.md
│   │   └── 003-kz-data-residency.md
│   ├── architecture.md
│   ├── folder-tree.md
│   ├── PRD.md
│   ├── tz.md
│   └── runbooks/
│       ├── local-dev.md
│       └── kafka-reindex.md
│
└── README.md
```

## Key Structure Changes

### Frontend/Backend Separation
- `frontend/` - Next.js app (unchanged)
- `backend/` - All Go services and shared packages under one logical grouping
  - `services/` - 18 microservices
  - `proto/` - Protocol buffer definitions
  - `pkg/` - Shared Go packages
  - `migrations/` - Database migration files
  - `scripts/` - Backend automation scripts

### Shared Resources (At Root Level)
- `infra/` - Infrastructure configs (nginx, postgres, kafka, etc.)
- `docs/` - Documentation  
- Docker Compose files
- Makefile

## Docker Compose Paths
Update service build paths to reflect new structure:
```yaml
gateway:
  build: ./backend/services/gateway
identity:
  build: ./backend/services/identity
# ... etc
```

## Directory Organization Benefits
✅ Clear frontend/backend separation  
✅ Easy to navigate logical groupings  
✅ Monorepo pattern widely recognized  
✅ Services grouped together for discovery  
✅ Shared infrastructure at root level