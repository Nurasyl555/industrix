# Industrial Equipment Marketplace — System Architecture

> Modular Monolith · CIS region · 1k–50k users · v2.0

**Stack:** Go · Fiber · PostgreSQL · Redis · Kafka · MinIO · OpenSearch · NGINX  
**Architecture:** Single Go binary · 3 domain modules · contracts layer · vertical slice  
**Compliance:** KZ data residency compliant · all self-hosted

---

## Connection Flow

```
Client → NGINX (HTTPS/TLS) → Backend Monolith (REST) → PostgreSQL / Redis / Kafka
                                      │
                              ┌───────┴────────┐
                              │   Fiber HTTP    │
                              │   Rate Limit    │
                              │   JWT Auth      │
                              │   Logging       │
                              └───────┬────────┘
                    ┌─────────────────┼─────────────────┐
                    │                 │                   │
              ┌─────▼─────┐   ┌──────▼──────┐   ┌──────▼──────┐
              │  Identity  │   │  Integrity  │   │ Marketplace │
              │  Module    │   │  Module     │   │   Module    │
              └────────────┘   └─────────────┘   └─────────────┘
```

---

## 01 · Client Layer

| Client         | Stack                        | Description                                                                             |
| -------------- | ---------------------------- | --------------------------------------------------------------------------------------- |
| **Web App**    | React / Next.js · TypeScript | Main marketplace UI. Equipment catalog, listings, search, deal management, admin panel. |
| **Mobile App** | React Native · iOS + Android | Shared codebase with web. Push notifications, camera, geolocation.                      |
| **External**   | REST API · Webhooks          | ERP/CRM integrations, payment provider callbacks.                                       |

---

## 02 · Platform Layer (Middleware)

All requests flow through the platform middleware stack — no separate gateway service.

| Middleware     | Purpose                                                                                |
| -------------- | -------------------------------------------------------------------------------------- |
| **NGINX**      | TLS termination, static file serving, load balancing, WebSocket proxying               |
| **Logging**    | Request method, path, status, latency, trace-id for every request (zerolog)            |
| **Rate Limit** | Redis sliding window limiter per user/IP, configurable per route                       |
| **JWT Auth**   | Direct in-process JWT validation using `jwtClient.ParseClaims()` — no gRPC call needed |

> **Why no API Gateway?** In a modular monolith, the gateway proxy layer is unnecessary. Middleware handles auth, rate limiting, and logging directly in the same process. This eliminates the gRPC roundtrip for token verification and removes an entire service from the deployment.

---

## 03 · Domain Modules

Three vertically-isolated modules. Each module owns: **types → repository → service → handler → module entry point**. Modules communicate only through the `contracts/` layer — no direct inter-module imports.

### Identity Module

> Auth + Profile (merged)

| Concern       | Details                                                                                |
| ------------- | -------------------------------------------------------------------------------------- |
| **Auth**      | Registration (email/phone), OTP verification, login, JWT issue/refresh, password reset |
| **Profile**   | User profiles, account settings, avatar management                                     |
| **DB**        | `users` table (PostgreSQL) + OTP codes (Redis)                                         |
| **Contracts** | Implements `UserProvider` — exposes `GetUserBasic()` for cross-module use              |

**Public routes** (no auth): `/auth/register`, `/auth/login`, `/auth/verify-otp`, `/auth/refresh`  
**Protected routes**: `GET /users/me`, `PUT /users/me`

### Integrity Module

> Company management & verification

| Concern       | Details                                                                         |
| ------------- | ------------------------------------------------------------------------------- |
| **Company**   | Company profiles, BIN validation (12-digit KZ format), verification workflow    |
| **Status**    | State machine: `pending → verified → rejected`                                  |
| **DB**        | `companies` table (PostgreSQL)                                                  |
| **Contracts** | Implements `CompanyProvider` — exposes `GetCompanyBasic()` for cross-module use |

**Protected routes**: `POST /companies`, `GET /companies/:id`, `PUT /companies/:id`

### Marketplace Module

> Reviews & reputation

| Concern        | Details                                                      |
| -------------- | ------------------------------------------------------------ |
| **Reviews**    | Post-deal reviews with star ratings                          |
| **Reputation** | Auto-calculated scores per entity (gold/silver/bronze tiers) |
| **DB**         | `reviews`, `reputation_scores` tables (PostgreSQL)           |

**Protected routes**: `POST /reviews`, `GET /reviews/:entityID`, `GET /reviews/:entityID/reputation`

---

## 04 · Contracts Layer

Modules never import each other directly. All cross-module communication goes through interfaces defined in `contracts/`:

```go
// contracts/contracts.go
type UserProvider interface {
    GetUserBasic(ctx context.Context, userID string) (*UserBasic, error)
}

type CompanyProvider interface {
    GetCompanyBasic(ctx context.Context, companyID string) (*CompanyBasic, error)
}
```

This enforces strict vertical isolation: if a future module needs user data, it receives a `UserProvider` — not a direct import of the identity package.

---

## 05 · Async & Event Bus

| Component | Config                               | Purpose                                       |
| --------- | ------------------------------------ | --------------------------------------------- |
| **Kafka** | 3-node cluster, replication factor 2 | Central event bus for cross-cutting concerns  |
| **Redis** | Single instance (cache + session)    | JWT sessions, OTP codes, rate limiting, cache |

**Key Kafka events** (planned):

- `company.verified` / `company.rejected` → unlock/block listing rights
- `deal.completed` → unlock review flow
- `review.created` → recalculate reputation
- `equipment.updated` → sync OpenSearch index

---

## 06 · Data Layer

| Store          | Technology                            | Purpose                                        |
| -------------- | ------------------------------------- | ---------------------------------------------- |
| **PostgreSQL** | PostgreSQL 15, single DB (`trust_db`) | Users, companies, reviews, reputation scores   |
| **Redis**      | Redis 7                               | Sessions, OTP codes, rate limiting, cache      |
| **MongoDB**    | MongoDB 7 (planned)                   | Chat messages, notifications                   |
| **OpenSearch** | OpenSearch 2.x (planned)              | Equipment full-text + faceted search           |
| **MinIO**      | S3-compatible (planned)               | Equipment photos, verification docs, contracts |

---

## 07 · Observability & Deployment

| Concern        | Stack                                    |
| -------------- | ---------------------------------------- |
| **Metrics**    | Prometheus + Grafana dashboards          |
| **Logging**    | Loki + structured JSON logs (zerolog)    |
| **Tracing**    | Jaeger distributed tracing (planned)     |
| **Deployment** | Docker Compose (dev) → Kubernetes (prod) |

**Single Dockerfile** with optimized layer caching:

1. Copy `go.mod + go.sum` → `go mod download` (cached unless deps change)
2. Copy source → `go build` (fast rebuild on code-only changes)
3. Alpine runtime image (~15MB final image)

---

## 08 · Key Architectural Decisions

| Decision | Rationale |
| -------- | --------- |

| **Vertical module isolation** | Each module is a self-contained vertical slice (types → repo → service → handler). Prevents spaghetti imports as codebase grows. |
| **Contracts layer** | Shared interfaces prevent direct inter-module coupling. Modules can be extracted to separate services later by implementing the same contract remotely. |
| **Kafka retained** | Even with a monolith, Kafka is needed for OpenSearch indexing, notification fanout, and future service extraction. |

---

## 09 · Future Modules (Planned)

| Module           | Contains                                              | Phase   |
| ---------------- | ----------------------------------------------------- | ------- |
| **Catalog**      | Equipment CRUD, categories, dynamic attribute schemas | Phase 2 |
| **Listing**      | Ad lifecycle, pricing, booking slots                  | Phase 2 |
| **Search**       | OpenSearch full-text + faceted + geo                  | Phase 2 |
| **Deal**         | Transaction orchestrator, state machine               | Phase 3 |
| **Payment**      | Kaspi Pay, Halyk Bank, escrow                         | Phase 3 |
| **Chat**         | WebSocket messaging (MongoDB backend)                 | Phase 4 |
| **Notification** | Push (FCM), Email (Postal)                            | Phase 4 |
| **Media**        | MinIO uploads, imgproxy transforms                    | Phase 5 |
| **Analytics**    | Event aggregation, seller dashboards                  | Phase 6 |

> All future modules will follow the same vertical slice pattern: `modules/<name>/` with types, repository, service, handler, module.go.

---

_Industrix · Architecture v2.0 — Modular Monolith · KZ-compliant · Generated 2026-02-28_
