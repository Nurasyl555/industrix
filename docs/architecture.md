
tiiny.host
Industrial Equipment Marketplace — System Architecture

// medium scale · CIS region · 1k–50k users · microservices · v1.0
Go · Gin/Fiber PostgreSQL MongoDB Redis Kafka MinIO OpenSearch NGINX gRPC · REST · WebSocket Gotenberg · imgproxy · 2GIS KZ data residency compliant 12 services · Docker/k8s
Client → Gateway (HTTPS/REST) Service ↔ Service (gRPC) Async (Kafka event) Cache read/write (Redis) DB read/write WebSocket
01 · Client Layer
web
Web App
React / Next.js · TypeScript
Main marketplace UI. Equipment catalog, listings, search, deal management, admin panel. SSR for SEO on catalog pages.
mobile
Mobile App
React Native · iOS + Android
Shared codebase with web where possible. Push notifications, camera for equipment photos, geolocation for region filtering.
3rd party
External Integrations
REST API · Webhooks
ERP/CRM integrations for enterprise clients. Payment provider callbacks. Electronic document flow (ЭДО) systems.
▼ HTTPS / REST · WebSocket
02 · Gateway & Edge Layer
edge
NGINX
nginx 1.25+
TLS termination, static file serving, load balancing across API Gateway instances, rate limiting at edge, WebSocket proxying for chat.
gateway
API Gateway
Go · Fiber · single deployable
JWT validation & auth middleware, request routing to downstream services, response aggregation, rate limiting per user/IP, request logging & tracing injection (trace-id headers).
Why a custom Go API Gateway? For medium scale, a thin Go/Fiber gateway avoids the operational complexity of Kong/Traefik while giving you full control over auth, rate limiting, and routing logic in the same language as your services. Scale to Kong if you need plugin ecosystem later.
▼ gRPC (internal) · REST (where appropriate)
03 · Core Domain Services  — 12 services · KZ data residency compliant · all self-hosted
KZ Data Residency: All services run on infrastructure physically located in Kazakhstan (Jusan Cloud / KazTransCom / Yandex Cloud KZ). No personal data transits to foreign SaaS. 3rd party tools used are either self-hosted or send only transient non-personal data (geocoding queries, push tokens at send-time only).
► Identity Service  merges: Auth + User & Profile + Company & Verification
svc-identity
Identity Service
Go · Gin · PostgreSQL (identity_db) · Redis
Single service owning all identity concerns — no reason to split at this scale, same team owns it all.

Auth module: Registration (email/phone), login, JWT issue/refresh/revoke, OTP via self-hosted SMS gateway (KZ operators: Beeline KZ, Kcell SMPP), password reset. Sessions in Redis.

Profile module: User profiles, account settings, notification preferences, reputation score (aggregated from Review Service via Kafka). Exposes gRPC GetUser / GetCompany endpoints consumed by all services.

Company & Verification module: Company profiles, БИН/ИНН validation against KZ tax authority format, document upload to MinIO (verification docs stored on-prem). Verification state machine: pending → under_review → verified / rejected. Admin review queue. Emits company.verified to unlock listing rights.
► Catalog & Discovery
svc-catalog
Catalog Service
Go · Gin · PostgreSQL (catalog_db) · Redis
Equipment CRUD, category taxonomy, dynamic technical attribute schemas per category type. Canonical source of truth. Also handles comparison endpoint — no separate service needed: GET /catalog/compare?ids=1,2,3 normalizes specs inline. Publishes all mutations to Kafka.
svc-listing
Listing Service
Go · Gin · PostgreSQL (listing_db)
Ad lifecycle: draft → moderation → active → archived. Pricing rules, view counters, listing stats. Enforces subscription plan limits via gRPC call to Integrity Service. Manages rental booking slots. Emits to Kafka for OpenSearch sync.
svc-search
Search Service
Go · Gin · OpenSearch · Redis · 2GIS API*
Full-text + faceted search with dynamic filters. Geo region filtering via internal KZ/CIS region table (PostgreSQL). 2GIS API used only for address geocoding — sends only address strings, no personal data. Autocomplete, recommendations, query history in Redis. Geo is a pkg/geo library, not a standalone service.
svc-booking
Booking & Availability Service
Go · Gin · PostgreSQL (booking_db) · Redis
Calendar-based availability for rentals. Hold + confirm slot pattern (Redis TTL for holds, PostgreSQL for confirmed). Optimistic locking for conflict prevention. Cancellation policies. Linked to Deal Service — booking created on deal init, released on cancel.
► Marketplace & Transactions
svc-deal
Deal Service
Go · Gin · PostgreSQL (deal_db)
Core transaction orchestrator. State machine: inquiry → negotiation → confirmed → in_progress → completed / cancelled. Coordinates Booking, Payment, Document services via gRPC. Emits deal lifecycle events to Kafka.
svc-payment
Payment Service
Go · Gin · PostgreSQL (payment_db)
Kaspi Pay, Halyk Bank, Uzcard/Humo integrations (all CIS-local providers — no foreign payment data storage). Escrow hold/release. Invoice generation. Transaction history, refunds. Webhook ingestion. Also handles subscription billing for seller tariff plans — Integrity Service calls Payment for charges.
svc-document
Document Service
Go · Gin · PostgreSQL (doc_db) · MinIO · Gotenberg*
Contract/oferta template rendering → PDF via Gotenberg (self-hosted, open-source HTML→PDF via headless Chrome — runs on your infra, zero data leaves KZ). ЭДО integration. Signed document storage in MinIO. Invoice PDFs. Called by Deal Service on confirmation.
svc-review
Review & Rating Service
Go · Gin · PostgreSQL (review_db)
Post-deal mutual reviews. Star ratings per dimension. Anti-fraud: only post-completed-deal reviews. Aggregate score recalculated on new review, pushed to Identity Service via gRPC. Unlocked by deal.completed Kafka event.
► Services Marketplace  merges: Additional Services + Logistics & Delivery
svc-services-mkt
Services Marketplace
Go · Gin · PostgreSQL (services_db) · 2GIS API* (route calc)
Single service for all third-party service offerings (TZ §2.8) — same domain, same team, different entity types in the same DB.

Service catalog module: Logistics/delivery, installation & commissioning, maintenance contracts, insurance offerings. Providers register with pricing models (fixed/per-km/per-day). Attached to deals as line items.

Logistics module: Delivery requests, carrier matching, route cost estimation via 2GIS Routing API (only route coordinates sent — no personal data), status tracking: pickup → in_transit → delivered. Carrier profiles and ratings.
► Communication
svc-chat
Chat Service
Go · Fiber · MongoDB (chat_db) · Redis · MinIO
WebSocket connections via Fiber. Conversation threads scoped to deals. Read receipts, delivery status, typing indicators. File/image sharing via MinIO presigned URLs. Message history in MongoDB. User presence in Redis pub/sub. All message data stays on-prem.
svc-notification
Notification Service
Go · Kafka consumer · MongoDB (notif_db) · FCM* · SMPP*
Pure Kafka consumer, no REST API. Fans out to channels:
In-app: stored in MongoDB (on-prem) ✓
Push: FCM/APNs — sends device token + message text only, no profile data stored at Google. Acceptable under KZ law as transient transport.
SMS: Beeline KZ / Kcell SMPP relay — KZ operators, data stays in KZ ✓
Email: Self-hosted Postal SMTP server (open-source) — no SendGrid/Mailgun ✓
► Engagement Service  merges: Favorites & Watchlist + Price History
svc-engagement
Engagement Service
Go · Gin · PostgreSQL (engagement_db) · Redis · Kafka consumer
Lightweight read-heavy service — two closely related features, one deployment.

Favorites module: User watchlists, collection grouping, watching for price drops. Emits favorite.price_dropped → Notification Service. Hot lists cached in Redis.

Price History module: Consumes equipment.updated events, records price snapshots. Exposes price charts per listing. Market benchmarks: avg/median per category+region. Powers "price dropped" badge and seller market insight panel.
► Platform Integrity  merges: Dispute + Subscription & Tariff + Audit Log
svc-integrity
Platform Integrity Service
Go · Gin · PostgreSQL (integrity_db: separate schemas per module) · Kafka consumer · MinIO
Three closely coupled concerns owned by the same team — platform trust and monetization enforcement.

Dispute module: Complaint filing against deals. Evidence in MinIO. State machine: filed → under_review → resolved / escalated. Admin arbitration. Escrow release/refund decisions back to Payment via gRPC.

Subscription & Tariff module: Seller plan management (free/basic/pro/enterprise). Listing limits, featured placement, analytics access. Billing via Payment Service gRPC. Feature-flag gRPC endpoint polled by Listing Service to enforce plan limits.

Audit Log module: Immutable append-only trail — write-once schema, no UPDATE/DELETE permissions. Consumes all sensitive Kafka events: deal changes, payments, verifications, moderation, disputes, subscriptions. Records actor + action + timestamp + IP + old→new state. CIS legal compliance requirement.
► Platform Support
svc-media
Media Service
Go · Fiber · MinIO · imgproxy*
Presigned URL generation for direct browser→MinIO uploads. Image transforms (resize, WebP, thumbnails) delegated to imgproxy (self-hosted, open-source — runs on your infra, zero data leaves KZ). Video thumbnail extraction. Media metadata in PG. Public URLs proxied via NGINX.
svc-analytics
Analytics Service
Go · Kafka consumer · PostgreSQL (analytics_db) · Redis
Self-hosted analytics — no PostHog cloud (personal data concern). Consumes all Kafka events. Seller dashboard: views, contact rate, deal conversion, price vs market. Admin: GMV, active listings, user growth, regional heatmaps. Pre-aggregated in PG, hot counters in Redis. Grafana dashboards on top.
svc-admin
Admin Service
Go · Gin · PostgreSQL
Moderation orchestrator. Listing review queue, user/company management, ban/suspend via gRPC to other services. Category & attribute schema management. Separate JWT scope — admin tokens not valid on user-facing routes. Reads from Analytics and Integrity services.
* Permitted 3rd Party Tools — KZ Compliant
Gotenberg — self-hosted HTML→PDF. Runs on your infra. ✓
imgproxy — self-hosted image transforms. Runs on your infra. ✓
2GIS API — geocoding + routing. Sends address/coords only, no PII. ✓
FCM / APNs — push delivery only. Token + message at send-time, no storage. ✓
Beeline KZ / Kcell SMPP — KZ operators for OTP/SMS. Data stays in KZ. ✓
Postal (self-hosted) — open-source SMTP server. Runs on your infra. ✓
✗ Supabase / Clerk — stores personal data outside KZ. Not permitted.
✗ SendGrid / Mailgun — logs email content on foreign servers. Not permitted.
✗ PostHog Cloud — user behavioral data on foreign servers. Not permitted.
04 · Async & Event Bus
broker
Apache Kafka
3-node cluster · replication factor 2
Central event bus. Consumers use Kafka consumer groups — each service subscribes independently.

equipment.created/updated/deleted → Search (index sync), Engagement (price watch), Analytics
listing.created/deactivated → Search (index sync), Analytics
company.verified/rejected → Catalog (unlock), Listing (unlock), Notification, Integrity (audit)
deal.status.changed → Notification, Booking, Listing, Integrity (audit), Analytics
deal.completed → Review (unlock), Services Marketplace, Integrity (audit), Analytics
payment.completed/failed → Deal, Notification, Integrity (audit), Analytics
message.sent → Notification (offline push/SMS/email fallback)
review.created → Identity (score update), Analytics
media.uploaded → Media (imgproxy transform pipeline)
dispute.filed/resolved → Notification, Payment (escrow action), Integrity (audit)
favorite.price_dropped → Notification
subscription.activated/expired → Notification, Listing (re-check limits), Integrity (audit)
delivery.status.changed → Notification, Analytics
moderation.action.taken → Notification, Integrity (audit)
cache
Redis
Redis 7 · 2 instances (cache + session)
Dual-purpose:

Session store: JWT refresh tokens, OTP codes, active user sessions

Cache: Hot catalog pages (TTL 5min), search result caching (TTL 1min), user profile cache, equipment availability status

Chat presence: Online/offline status, active WebSocket connections registry

Rate limiting: Sliding window counters per user/IP
05 · Data Layer
relational
PostgreSQL
PostgreSQL 15 · per-service DBs
auth_db: users, sessions, verifications
catalog_db: equipment, categories, attributes
listing_db: ads, pricing, bookings
deal_db: transactions, contracts, status
payment_db: payments, invoices, escrow
review_db: ratings, comments
Each service owns its own schema. No cross-DB joins.
document
MongoDB
MongoDB 7 · replica set
chat_db: messages, conversations, read receipts — flexible schema, high write throughput, easy embedding of reactions/attachments

notifications_db: in-app notification feed per user, unread counts, notification preferences
search
OpenSearch
OpenSearch 2.x · 3-node cluster
Equipment index with dynamic mappings for technical attributes per category. Handles: full-text search, faceted filters (price range, region, category, specs), geo-distance queries (equipment location), autocomplete suggestions. Fed by Kafka consumer from Catalog events.
object store
MinIO
MinIO · S3-compatible · distributed mode
equipment-media: photos, videos, thumbnails
documents: company verification docs, contracts, invoices
chat-files: files/images shared in chat
Presigned URLs for direct browser upload. NGINX proxies public bucket URLs.
06 · Observability & Infrastructure
observability
Monitoring Stack
Prometheus · Grafana · Jaeger
Prometheus scrapes all Go services (stdlib metrics). Grafana dashboards per service. Jaeger distributed tracing — trace-id injected at gateway, propagated via gRPC metadata through entire call chain.
logging
Logging
Loki · structured JSON logs
All services emit structured JSON logs (zerolog in Go). Loki aggregates logs, queryable from Grafana. Log levels per service configurable at runtime via env. Correlation by trace-id across services.
deploy
Deployment
Docker Compose (dev) → Kubernetes (prod)
Each service containerized. Docker Compose for local dev with all dependencies. K8s manifests for production with HPA (horizontal pod autoscaling) per service. CI/CD via GitHub Actions → container registry → k8s rollout.
07 · Key Request Flows
Equipment Search with Filters
01
Client sends GET /search?q=crane®ion=almaty&price_max=5M → NGINX → API Gateway
02
Gateway validates JWT, injects user context headers, routes to Search Service REST
03
Search Service checks Redis cache for identical query hash (TTL 60s)
04
Cache miss → queries OpenSearch with bool query + range filters + geo filter
05
Results cached in Redis, returned to client. Query saved to user history.
Listing Creation + Media Upload
01
Seller creates listing via POST /listings → Gateway → Listing Service
02
Listing Service calls Catalog Service gRPC to validate category/attributes schema
03
Client requests presigned upload URL from Media Service, uploads photos directly to MinIO
04
Media Service emits media.uploaded to Kafka → async image processing (resize, WebP)
05
Listing saved → equipment.created event → OpenSearch indexer consumer updates search index
Real-time Chat Message
01
Buyer opens chat on deal page → establishes WebSocket to Chat Service via NGINX proxy
02
Connection authenticated via JWT query param. Presence registered in Redis
03
Message sent → stored in MongoDB, delivered to recipient WS if online
04
Recipient offline → Chat Service emits message.sent event to Kafka
05
Notification Service consumes event → sends push notification (FCM/APNs) or email
Deal Confirmation + Payment
01
Seller confirms deal → PUT /deals/:id/confirm → Deal Service
02
Deal Service calls Payment Service gRPC to initiate escrow hold
03
Deal status machine transitions → emits deal.status.changed to Kafka
04
Notification Service sends confirmation to both parties (email + push)
05
Payment provider webhook → Payment Service releases escrow, emits payment.completed
08 · Service Boundaries & Data Ownership
SERVICE 	OWNS (DB) 	EMITS (Kafka) 	CONSUMES (Kafka) 	CALLS (gRPC)
IDENTITY
Identity
auth+profile+company	identity_db (PG) + Redis + MinIO	user.registered, company.verified, company.rejected	review.created (score update)	— (called by all services)
CATALOG & DISCOVERY
Catalog
+comparison endpoint	catalog_db (PG) + Redis	equipment.created, .updated, .deleted	company.verified	Identity (verify check)
Listing	listing_db (PG)	listing.created, listing.deactivated	deal.status.changed, subscription.expired	Catalog (schema), Integrity (plan limits)
Search
+pkg/geo library	OpenSearch index + Redis	—	equipment.*, listing.*	— (2GIS API for geocoding)
Booking	booking_db (PG) + Redis	booking.confirmed, booking.cancelled	deal.status.changed	Listing (availability update)
MARKETPLACE & TRANSACTIONS
Deal	deal_db (PG)	deal.status.changed, deal.completed	payment.completed, payment.failed	Payment, Booking, Document, Identity
Payment
+subscription billing	payment_db (PG)	payment.completed, payment.failed	—	— (called by Deal, Integrity)
Document
Gotenberg ✓	doc_db (PG) + MinIO	document.generated, document.signed	deal.status.changed	— (called by Deal)
Review	review_db (PG)	review.created	deal.completed	Identity (score update)
SERVICES MARKETPLACE
Services Marketplace
addl services + logistics	services_db (PG)	service.order.created, delivery.status.changed	deal.completed	Identity (provider verify), 2GIS (routing)
COMMUNICATION
Chat	chat_db (Mongo)	message.sent	—	Identity (profile), Deal (auth)
Notification
Postal+FCM+SMPP ✓	notif_db (Mongo)	—	message.sent, deal.*, payment.*, company.*, dispute.*, delivery.*, favorite.*	—
ENGAGEMENT
Engagement
favorites + price history	engagement_db (PG) + Redis	favorite.price_dropped	equipment.updated	Catalog (item details)
PLATFORM INTEGRITY
Platform Integrity
dispute+subscription+audit	integrity_db (PG, 3 schemas) + MinIO	dispute.filed, dispute.resolved, subscription.activated, subscription.expired	deal.*, payment.*, company.*, moderation.*, dispute.*, subscription.*	Payment (escrow/billing), Listing (plan limits)
PLATFORM SUPPORT
Media
imgproxy ✓	media_db (PG) + MinIO	media.uploaded, media.processed	—	—
Analytics
self-hosted	analytics_db (PG) + Redis	—	all events (consumer group)	—
Admin	— (reads from others)	moderation.action.taken	—	All services (mod actions)
09 · Key Architectural Decisions
Kafka vs Redis Streams: At medium scale (1k–50k users), Redis Streams would technically suffice, but Kafka is the right choice here because: (1) OpenSearch indexing requires reliable, replayable events — Kafka's log retention is critical if OpenSearch needs reindexing; (2) Notification fanout to multiple consumers (push, email, in-app) is exactly the consumer group pattern Kafka excels at; (3) You won't need to migrate when you grow.
PostgreSQL per service, not shared: Each service owns its DB entirely — no cross-service DB queries. Inter-service data needs go through gRPC or Kafka events. This keeps services independently deployable and prevents coupling. Consistency is eventual where needed (e.g., OpenSearch index), strong where required (deals, payments).
MongoDB for Chat: Message schema is inherently flexible (reactions, attachments, thread replies evolve). MongoDB's document model + high write throughput suits chat better than PG. Conversations are denormalized per-thread, making read patterns efficient without joins.
gRPC between services: Typed contracts via protobuf prevent API drift between services. Binary protocol reduces latency vs REST for internal calls. Easy code generation for Go. Keep REST only at the gateway boundary facing clients.
MinIO for media: S3-compatible API means you can migrate to AWS S3 / Yandex Object Storage with zero code changes if needed. Presigned URL pattern offloads file upload bandwidth from your app servers entirely — client uploads direct to MinIO, only metadata goes through your API.
OpenSearch over Elasticsearch: Fully open-source (Apache 2.0), no licensing concerns. AWS fork with active development. Equivalent feature set for your use case. Critical for complex equipment search: nested attributes (a crane has different filterable specs than a compressor) handled elegantly with dynamic mappings and nested queries.
MVP SCOPE — SHIP FIRST
For MVP, collapse to 5 deployments: Identity + Catalog/Listing (merged) + Search + Chat + Notification. Single PostgreSQL with separate schemas. Redis Streams instead of Kafka. MinIO from day one. Gotenberg + imgproxy from day one (free, self-hosted). Skip Booking, Document, Services Marketplace, Engagement, Integrity, Analytics — add post-launch. The architecture above is the target state (12 services, KZ-compliant); build toward it incrementally.
Industrial Equipment Marketplace · Architecture v1.3 — consolidated & KZ-compliant · Generated 2026-02-25
