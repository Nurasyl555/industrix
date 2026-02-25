
Industrial Equipment Marketplace — Implementation Plan

// 7 phases · 12 services + frontend + infra · KZ-compliant · monorepo
Backend (Go) Frontend (Next.js) Infra / DevOps Proto / Shared MUST SHOULD NICE
PHASE 1
Foundation
~48 tasks
Weeks 1–3
PHASE 2
Core Marketplace
~72 tasks
Weeks 4–8
PHASE 3
Transactions
~65 tasks
Weeks 9–13
PHASE 4
Communication
~42 tasks
Weeks 14–17
PHASE 5
Intelligence
~38 tasks
Weeks 18–21
PHASE 6
Integrity & Monetisation
~44 tasks
Weeks 22–26
PHASE 7
Hardening & Launch
~30 tasks
Weeks 27–30
Function / Task	Type	Priority	Notes
init monorepo
Initialize git repo, root go.work, .gitignore, .env.example with all vars documented
	INFRA	MUST	go.work for multi-module workspace
docker-compose.infra.yml
PostgreSQL 15, MongoDB 7, Redis 7 (×2 instances), Kafka+Zookeeper (3-node), OpenSearch, MinIO, imgproxy, Gotenberg, Postal SMTP
	INFRA	MUST	All self-hosted, KZ-compliant
docker-compose.yml
All 12 services + frontend + all infra in one compose for CI and full local testing
	INFRA	MUST	
docker-compose.override.yml
Dev overrides: port mappings, volume mounts, air hot-reload commands for Go services
	INFRA	MUST	
Makefile
up, down, up-infra, dev-svc, dev-frontend, proto, migrate, migrate-svc, seed, test, test-svc, lint, logs, health targets
	INFRA	MUST	
nginx.conf + conf.d/
gateway.conf (proxy to API gateway), frontend.conf (Next.js SSR), minio.conf (public bucket), ws.conf (WebSocket upgrade for chat)
	INFRA	MUST	TLS via Let's Encrypt
infra/postgres/init/00_create_databases.sql
CREATE DATABASE for each service: identity_db, catalog_db, listing_db, booking_db, deal_db, payment_db, doc_db, review_db, services_db, engagement_db, integrity_db, analytics_db, media_db
	INFRA	MUST	
infra/kafka/topics.sh
Create all 14 Kafka topics with correct partition counts and retention. Run once on first start.
	INFRA	MUST	
infra/minio/buckets.sh
Create buckets: equipment-media, documents, chat-files, dispute-evidence. Set public/private policies.
	INFRA	MUST	
infra/opensearch/mappings/equipment.json
Index mapping with dynamic templates for technical attributes, geo_point for location, keyword fields for filters
	INFRA	MUST	
scripts/healthcheck.sh
Poll all service /health endpoints, report which are up/down
	INFRA	SHOULD	
GitHub Actions: ci.yml
On PR: go test ./..., golangci-lint, ESLint, proto generation check (no diff)
	INFRA	MUST	
Grafana + Prometheus + Loki + Jaeger
docker-compose entries for observability stack. Base dashboards imported.
	INFRA	SHOULD	Add in phase 1, tune later
Function / Task	Type	Priority	Notes
pkg/postgres.NewClient()
PG connection pool with pgx, health check, migration runner (golang-migrate)
	SHARED	MUST	
pkg/redis.NewClient()
Redis client with sentinel support, typed key helpers
	SHARED	MUST	
pkg/mongo.NewClient()
MongoDB client with replica set connection, context timeout helpers
	SHARED	MUST	
pkg/kafka.NewProducer() / NewConsumer()
Sarama-based producer with retry, consumer with consumer group + DLQ (dead letter queue) support
	SHARED	MUST	
pkg/minio.NewClient() / PresignURL()
MinIO client, presigned PUT URL generation, presigned GET URL generation
	SHARED	MUST	
pkg/jwt.ParseClaims() / IssuePair()
Shared JWT claims struct (userID, companyID, role, verified), access+refresh token issuance
	SHARED	MUST	
pkg/logger.New()
zerolog setup: structured JSON, trace-id injection, log level from env, service name field
	SHARED	MUST	
pkg/tracer.Init()
OpenTelemetry tracer setup → Jaeger exporter. Provides TraceFromContext() helper.
	SHARED	SHOULD	
pkg/errors.New() / Wrap()
Typed error codes: NOT_FOUND, UNAUTHORIZED, VALIDATION, CONFLICT, INTERNAL. HTTP status mapping.
	SHARED	MUST	
pkg/geo.LookupRegion() / Geocode()
KZ/CIS region lookup table in PG, 2GIS API client for address geocoding (no PII sent)
	SHARED	MUST	
File / RPC	Type	Priority	Notes
identity/v1/identity.proto
GetUser(id) → User, GetCompany(id) → Company, VerifyToken(token) → Claims, GetUserBatch(ids) → []User
	PROTO	MUST	
catalog/v1/catalog.proto
GetEquipment(id) → Equipment, ValidateAttributes(categoryID, attrs) → ValidationResult, GetCategorySchema(id) → Schema
	PROTO	MUST	
payment/v1/payment.proto
InitiateEscrow(dealID, amount) → EscrowID, ReleaseEscrow(escrowID) → Result, RefundEscrow(escrowID) → Result, ChargeSubscription(companyID, planID) → Receipt
	PROTO	MUST	
booking/v1/booking.proto
CreateHold(listingID, dates) → HoldID, ConfirmBooking(holdID) → Booking, CancelBooking(bookingID) → Result, CheckAvailability(listingID, dates) → Available
	PROTO	MUST	
integrity/v1/integrity.proto
CheckPlanLimits(companyID, action) → Allowed/Blocked, GetSubscription(companyID) → Plan, GetPlanFeatures(planID) → Features
	PROTO	MUST	
scripts/proto-gen.sh
buf generate all protos → pkg/gen/go/ and pkg/gen/ts/. Committed to repo.
	PROTO	MUST	
Auth Module
Function / Task	Type	Priority	Notes
POST /auth/register
Register with email+phone+password. Hash with bcrypt. Send OTP to phone (Kcell/Beeline SMPP). Return pending state until OTP confirmed.
	BE	MUST	
POST /auth/verify-otp
Validate OTP from Redis (TTL 5min). Activate account. Issue JWT access+refresh pair.
	BE	MUST	
POST /auth/login
Email+password login. bcrypt compare. Issue JWT pair. Store refresh token in Redis with device fingerprint.
	BE	MUST	
POST /auth/refresh
Validate refresh token from Redis. Rotate tokens (old refresh invalidated). Return new pair.
	BE	MUST	
POST /auth/logout
Invalidate refresh token in Redis. Optionally revoke all sessions for user.
	BE	MUST	
POST /auth/forgot-password
Send reset OTP/link to phone or email. Store token in Redis (TTL 15min).
	BE	MUST	
POST /auth/reset-password
Validate reset token, hash new password, invalidate all existing sessions.
	BE	MUST	
Profile Module
GET /users/me
Return authenticated user's profile with company info if attached.
	BE	MUST	
PUT /users/me
Update name, avatar (trigger media upload), contact details. Emit user.profile.updated to Kafka.
	BE	MUST	
PUT /users/me/avatar
Request presigned upload URL from Media service. Store resulting MinIO URL on profile.
	BE	SHOULD	
GET /users/:id/public
Public profile: name, avatar, company name, verified badge, rating, reviews count.
	BE	MUST	
PUT /users/me/notification-preferences
Toggle push/SMS/email per event type. Stored in PG, read by Notification service.
	BE	SHOULD	
grpc: GetUser() / GetUserBatch() / GetCompany()
Internal gRPC endpoints consumed by Deal, Chat, Review, Admin services.
	BE	MUST	
Company & Verification Module
POST /companies
Create company profile. БИН/ИНН format validation (12-digit KZ BIN). Set status = pending.
	BE	MUST	
PUT /companies/me
Update company info. If already verified, changes require re-verification for legal fields.
	BE	MUST	
POST /companies/me/documents
Upload verification docs (БИН cert, charter, director ID). Get presigned URL, store metadata. Move status to under_review.
	BE	MUST	
GET /companies/me/verification-status
Return current verification state, reviewer notes, list of submitted docs.
	BE	MUST	
verification state machine
pending → under_review → verified / rejected. On verified: emit company.verified to Kafka. On rejected: emit company.rejected with reason.
	BE	MUST	
Kafka consumer: review.created
On new review for a company member, recalculate and update company reputation score.
	BE	SHOULD	
DB Migrations
001_users.sql
users, sessions, otp_codes tables
	BE	MUST	
002_companies.sql
companies, company_documents, verification_history tables
	BE	MUST	
Function / Task	Type	Priority	Notes
middleware/auth.go: ValidateJWT()
Extract Bearer token, validate signature, inject userID+role into request context. Calls Identity grpc.VerifyToken() for revocation check.
	BE	MUST	
middleware/ratelimit.go: SlidingWindow()
Redis-based sliding window rate limiter. Per-user and per-IP limits. Configurable per route.
	BE	MUST	
middleware/tracing.go: InjectTraceID()
Generate or propagate X-Trace-ID header. Inject into downstream request headers and context.
	BE	SHOULD	
middleware/logging.go: RequestLogger()
Log method, path, status, latency, trace-id for every request using zerolog.
	BE	MUST	
proxy/router.go: RegisterRoutes()
Map all API routes to downstream services. Reverse proxy with header forwarding. Separate admin route group with stricter auth scope check.
	BE	MUST	
GET /health
Gateway liveness check. Optionally aggregate downstream /health responses.
	BE	MUST	
Function / Task	Type	Priority	Notes
Next.js App Router scaffold
Route groups: (auth), (marketplace), (deals), (account), (chat), (admin). Root layout with providers.
	FE	MUST	
lib/api/ typed client
Typed fetch wrappers per service (identity.ts, catalog.ts, search.ts…). Automatic JWT injection. 401 → refresh → retry logic.
	FE	MUST	
Zustand auth store
user, company, token, isVerified state. Persist access token in memory, refresh token in httpOnly cookie.
	FE	MUST	
components/ui/ design system
Button, Input, Select, Modal, Badge, Spinner, Toast, Table, Pagination, Tabs — base components
	FE	MUST	
components/layout/ Header / Sidebar / Footer
Responsive layout. Auth-aware nav (show/hide links by role). Notification bell with unread count.
	FE	MUST	
Login page / Register page
Forms with validation. OTP verification flow. Role selection (buyer/seller/service provider). i18n: RU + KK.
	FE	MUST	
Total scope: ~339 tracked tasks across 7 phases, 12 backend services, 1 frontend, shared packages, infra, and proto definitions. Phases 1–3 constitute the shippable MVP. Phases 4–7 add full-featured communication, intelligence, monetisation, and production hardening. Each phase produces a deployable, testable increment.
IEM · 