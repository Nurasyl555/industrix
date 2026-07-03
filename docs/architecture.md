# Industrial Equipment Marketplace — System Architecture


**Stack:** Go 1.24 · Fiber · PostgreSQL · Redis · Kafka · MinIO · OpenSearch · MongoDB · NGINX  
**Architecture:** Single Go binary · 12 domain modules · contracts layer · vertical slice  

---

## Connection Flow

```
Client → NGINX (HTTPS/TLS) → Backend Monolith (:8080) → PostgreSQL / Redis / Kafka
                                      │
                              ┌───────┴────────┐
                              │   Fiber HTTP    │
                              │   Rate Limit    │
                              │   JWT Auth      │
                              │   Logging       │
                              └───────┬────────┘
          ┌──────────┬──────────┬─────┴──────┬──────────┬──────────┐
          │          │          │            │          │          │
     ┌────▼───┐ ┌───▼────┐ ┌──▼─────┐ ┌───▼───┐ ┌───▼────┐ ┌──▼─────┐
     │Identity│ │Integrity│ │Catalog │ │Listing│ │ Search │ │  Deal  │
     └────────┘ └────────┘ └────────┘ └───────┘ └────────┘ └────────┘
     ┌────────┐ ┌────────┐ ┌────────┐ ┌───────┐ ┌────────┐ ┌────────┐
     │Payment │ │  Chat  │ │ Notif  │ │ Media │ │ Review │ │Analytics│
     └────────┘ └────────┘ └────────┘ └───────┘ └────────┘ └────────┘
```

---

## 01 · Client Layer

| Client         | Stack                        | Description                                                                             |
| -------------- | ---------------------------- | --------------------------------------------------------------------------------------- |
| **Web App**    | React / Next.js · TypeScript | Main marketplace UI. Equipment catalog, listings, search, deal management, admin panel. |
| **Mobile App** | React Native · iOS + Android | Shared codebase with web. Push notifications, camera, geolocation.                      |
| **External**   | REST API · Webhooks          | ERP/CRM integrations, payment provider callbacks, ЭДО systems.                          |

---

## 02 · Platform Layer (Middleware)

All requests flow through the platform middleware stack — no separate gateway service.

| Middleware     | Purpose                                                                                |
| -------------- | -------------------------------------------------------------------------------------- |
| **NGINX**      | TLS termination, static file serving, load balancing, WebSocket proxying               |
| **Logging**    | Request method, path, status, latency, trace-id for every request (zerolog)            |
| **Rate Limit** | Redis sliding window limiter per user/IP, configurable per route                       |
| **JWT Auth**   | Direct in-process JWT validation using `jwtClient.ParseClaims()` — no gRPC call needed |

---

## 03 · Domain Modules

12 vertically-isolated modules. Each module owns: **types → repository → service → handler → module entry point**. Modules communicate only through the `contracts/` layer — no direct inter-module imports.

### Identity Module ✅

> Auth + User Profile

| Concern       | Details                                                                                |
| ------------- | -------------------------------------------------------------------------------------- |
| **Auth**      | Registration (email/phone), OTP verification, login, JWT issue/refresh, password reset |
| **Profile**   | User profiles, account settings, avatar management                                     |
| **DB**        | `users` table (PostgreSQL) + OTP codes (Redis)                                         |
| **Contracts** | Implements `UserProvider` — exposes `GetUserBasic()` for cross-module use              |

**Public routes**: `/auth/email/register`, `/auth/email/login`, `/auth/phone/login`, `/auth/phone/verify`, `/auth/oauth/google`, `/auth/refresh`  
**Protected routes**: `GET /users/me`, `PUT /users/me`

### Integrity Module ✅

> Company management & verification

| Concern           | Details                                                                         |
| ----------------- | ------------------------------------------------------------------------------- |
| **Company**       | Company profiles, BIN validation (12-digit KZ format), verification workflow    |
| **Status**        | State machine: `pending → verified → rejected`                                  |
| **Disputes**      | Complaint filing, evidence upload (MinIO), arbitration                          |
| **Subscriptions** | Seller tariff plans (free/basic/pro/enterprise), billing via Payment module     |
| **Audit Log**     | Immutable append-only trail — CIS legal compliance                              |
| **DB**            | `companies`, `disputes`, `subscriptions`, `audit_logs` tables (PostgreSQL)      |
| **Contracts**     | Implements `CompanyProvider` — exposes `GetCompanyBasic()` for cross-module use |

**Protected routes**: `POST /companies`, `GET /companies/:id`, `PUT /companies/:id`

### Catalog Module 🟡 MVP

> Equipment CRUD & category taxonomy

**Implemented (MVP):** equipment CRUD with owner checks, flat category list,
SQL filtering by category/region/search, public browse + protected writes.
**Deferred:** dynamic attribute schemas, `/catalog/compare`, Kafka events.

| Concern        | Details                                                        |
| -------------- | -------------------------------------------------------------- |
| **Equipment**  | Create, read, update, delete equipment items ✅                |
| **Categories** | Category list (hierarchy column exists, flat for now) ✅        |
| **Attributes** | Dynamic technical attribute schemas per category type — planned |
| **Comparison** | Compare endpoint: `GET /catalog/compare?ids=1,2,3` — planned   |
| **DB**         | `equipment`, `categories` tables (PostgreSQL) ✅                |
| **Events**     | `equipment.created/updated/deleted` to Kafka — planned         |

### Listing Module 🟡 MVP

> Ad lifecycle & pricing

**Implemented (MVP):** sale/rental listings with price + optional rental
period, owner checks, public browse (joined with equipment) + protected
CRUD/publish/archive. **Deferred:** moderation state, stats, plan limits, Kafka.

| Concern       | Details                                                     |
| ------------- | ----------------------------------------------------------- |
| **Lifecycle** | State machine: `draft → active → archived` ✅ (no moderation step yet) |
| **Pricing**   | Fixed price (sale) or per day/week/month (rental) ✅         |
| **Stats**     | View counters, contact rate, listing analytics — planned    |
| **Limits**    | Subscription plan enforcement via Integrity contract — planned |
| **DB**        | `listings` table (PostgreSQL) ✅                             |
| **Events**    | `listing.created/deactivated` to Kafka — planned            |

### Search Module

> Full-text + faceted search

| Concern          | Details                                                                 |
| ---------------- | ----------------------------------------------------------------------- |
| **Search**       | OpenSearch full-text + faceted filters (price, region, category, specs) |
| **Geo**          | Region filtering via 2GIS API (address geocoding only, no PII)          |
| **Autocomplete** | Query suggestions, history                                              |
| **Cache**        | Redis cache for hot queries (TTL 60s)                                   |
| **Indexer**      | Kafka consumer for OpenSearch index sync from Catalog/Listing events    |

### Deal Module 🟡 MVP

> Transaction orchestrator

**Implemented (MVP):** a buyer inquiry on a listing with a **two-way realtime
message thread** — both parties view it in `/my-deals`, chat live over a
WebSocket, either can close it. Validates the listing is active and that
buyers can't inquire on their own listings. **Deferred:** the full negotiation
state machine, Booking/Payment/Document orchestration, Kafka events.

| Concern           | Details                                                                   |
| ----------------- | ------------------------------------------------------------------------- |
| **State machine** | `inquiry → closed` ✅ (full `negotiation → confirmed → … → completed` planned) |
| **Messaging**     | Two-way thread, realtime via Fiber WebSocket `/ws/deals/:id` ✅ (cookie-auth, in-memory hub) |
| **Coordination**  | Orchestrates Booking, Payment, Document modules — planned                  |
| **DB**            | `deals`, `deal_messages` tables (PostgreSQL) ✅                            |
| **Events**        | `deal.status.changed`, `deal.completed` to Kafka — planned                 |

### Payment Module

> Payment processing & escrow

| Concern           | Details                                              |
| ----------------- | ---------------------------------------------------- |
| **Providers**     | Kaspi Pay, Halyk Bank, Uzcard/Humo (CIS-local only)  |
| **Escrow**        | Hold/release pattern for secure transactions         |
| **Invoices**      | Invoice generation, transaction history, refunds     |
| **Subscriptions** | Subscription billing for seller tariff plans         |
| **DB**            | `payments`, `invoices`, `escrow` tables (PostgreSQL) |

### Booking Module

> Calendar-based availability

| Concern          | Details                                                          |
| ---------------- | ---------------------------------------------------------------- |
| **Availability** | Calendar-based availability for rentals                          |
| **Holds**        | Redis TTL for temporary holds, PostgreSQL for confirmed bookings |
| **Conflicts**    | Optimistic locking for conflict prevention                       |
| **DB**           | `bookings`, `availability` tables (PostgreSQL)                   |

### Chat Module

> Real-time messaging

| Concern       | Details                                                                 |
| ------------- | ----------------------------------------------------------------------- |
| **WebSocket** | Real-time connections via Fiber                                         |
| **Messages**  | Conversation threads scoped to deals                                    |
| **Features**  | Read receipts, typing indicators, file sharing via MinIO presigned URLs |
| **DB**        | MongoDB (chat_db) — messages, conversations                             |
| **Presence**  | Redis pub/sub for online/offline status                                 |

### Notification Module

> Multi-channel alerts

| Concern            | Details                                                         |
| ------------------ | --------------------------------------------------------------- |
| **Kafka consumer** | Fans out events to channels — no REST API                       |
| **In-app**         | MongoDB notification feed per user                              |
| **Push**           | FCM/APNs — device token + message only, no PII stored at Google |
| **Email**          | Self-hosted Postal SMTP server (open-source)                    |

### Media Module

> File uploads & transforms

| Concern       | Details                                                     |
| ------------- | ----------------------------------------------------------- |
| **Upload**    | Presigned URL generation for direct browser → MinIO uploads |
| **Transform** | imgproxy (self-hosted) for resize, WebP, thumbnails         |
| **Video**     | Thumbnail extraction for video uploads                      |
| **Storage**   | MinIO buckets: `equipment-media`, `documents`, `chat-files` |

### Review & Marketplace Module ✅

> Reviews, ratings & reputation

| Concern           | Details                                                             |
| ----------------- | ------------------------------------------------------------------- |
| **Reviews**       | Post-deal reviews with star ratings, anti-fraud validation          |
| **Reputation**    | Auto-calculated scores per entity (gold/silver/bronze tiers)        |
| **Engagement**    | Favorites, watchlists, price drop alerts                            |
| **Price History** | Price snapshots, market benchmarks per category+region              |
| **DB**            | `reviews`, `reputation_scores`, `favorites`, `price_history` tables |

### Analytics Module

> Self-hosted analytics

| Concern    | Details                                               |
| ---------- | ----------------------------------------------------- |
| **Events** | Kafka consumer — aggregates all platform events       |
| **Seller** | Views, contact rate, deal conversion, price vs market |
| **Admin**  | GMV, active listings, user growth, regional heatmaps  |
| **DB**     | Pre-aggregated in PostgreSQL, hot counters in Redis   |

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

type EquipmentProvider interface { // implemented by catalog
    GetEquipmentBasic(ctx context.Context, equipmentID string) (*EquipmentBasic, error)
}

type ListingProvider interface { // implemented by listing
    GetListingBasic(ctx context.Context, listingID string) (*ListingBasic, error)
}

// Future contracts added as modules grow:
// PaymentProvider, BookingProvider, etc.
```

This enforces strict vertical isolation: if a future module needs user data, it receives a `UserProvider` — not a direct import of the identity package.

---

## 05 · Async & Event Bus

| Component | Config                               | Purpose                                       |
| --------- | ------------------------------------ | --------------------------------------------- |
| **Kafka** | 3-node cluster, replication factor 2 | Central event bus for cross-cutting concerns  |
| **Redis** | Single instance (cache + session)    | JWT sessions, OTP codes, rate limiting, cache |

**Key Kafka events:**

- `company.verified` / `company.rejected` → unlock/block listing rights
- `deal.completed` → unlock review flow
- `deal.status.changed` → Notification, Booking, Analytics
- `review.created` → recalculate reputation
- `equipment.created/updated/deleted` → OpenSearch index sync
- `listing.created/deactivated` → Search index, Analytics
- `payment.completed/failed` → Deal state, Notification, Audit
- `message.sent` → offline push/email fallback
- `media.uploaded` → imgproxy transform pipeline
- `dispute.filed/resolved` → Notification, Payment escrow action
- `favorite.price_dropped` → Notification
- `subscription.activated/expired` → Listing limit enforcement

---

## 06 · Data Layer

| Store          | Technology                            | Purpose                                                            |
| -------------- | ------------------------------------- | ------------------------------------------------------------------ |
| **PostgreSQL** | PostgreSQL 15, single DB (`trust_db`) | Users, companies, equipment, listings, deals, payments, reviews    |
| **Redis**      | Redis 7                               | Sessions, OTP codes, rate limiting, cache, booking holds, presence |
| **MongoDB**    | MongoDB 7                             | Chat messages, conversations, notification feed                    |
| **OpenSearch** | OpenSearch 2.x                        | Equipment full-text + faceted + geo search                         |
| **MinIO**      | S3-compatible                         | Equipment photos, verification docs, contracts, chat files         |

---

## 07 · Observability & Deployment

| Concern        | Stack                                    |
| -------------- | ---------------------------------------- |
| **Metrics**    | Prometheus + Grafana dashboards          |
| **Logging**    | Loki + structured JSON logs (zerolog)    |
| **Tracing**    | Jaeger distributed tracing               |
| **Deployment** | Docker Compose (dev) → Kubernetes (prod) |

**Single Dockerfile** with optimized layer caching:

1. Copy `go.mod + go.sum` → `go mod download` (cached unless deps change)
2. Copy source → `go build` (fast rebuild on code-only changes)
3. Alpine runtime image (~15MB final image)

---

## 08 · Key Architectural Decisions

| Decision                      | Rationale                                                                                                                                               |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Vertical module isolation** | Each module is a self-contained vertical slice (types → repo → service → handler). Prevents spaghetti imports as codebase grows.                        |
| **Contracts layer**           | Shared interfaces prevent direct inter-module coupling. Modules can be extracted to separate services later by implementing the same contract remotely. |
| **Kafka retained**            | Even with a monolith, Kafka is needed for OpenSearch indexing, notification fanout, and future service extraction.                                      |

