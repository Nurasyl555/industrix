
Industrial Equipment Marketplace — Implementation Plan

// 7 phases · 9 services + frontend + infra · KZ-compliant · monorepo
// Service Integration: identity+integrity+review → trust-service, catalog+listing+search → inventory-service,
// booking+deal+payment → transaction-service, document+media+engagement → content-service, chat+notification → communication-service

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
All 9 services + frontend + all infra in one compose for CI and full local testing
	INFRA	MUST	Service count reduced from 12 to 9 via domain integration

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
CREATE DATABASE for each service: trust_db, inventory_db, transaction_db, content_db, communication_db, services_db, analytics_db
	INFRA	MUST	Consolidated from 12 to 7 databases via domain grouping

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
trust/v1/trust.proto
GetUser(id) → User, GetCompany(id) → Company, VerifyToken(token) → Claims, GetUserBatch(ids) → []User, GetVerificationStatus(companyID) → Status, SubmitDocument(companyID, doc) → Result, GetReputation(companyID) → Score
	PROTO	MUST	Merged: identity + integrity + review
inventory/v1/inventory.proto
GetEquipment(id) → Equipment, ValidateAttributes(categoryID, attrs) → ValidationResult, GetCategorySchema(id) → Schema, SearchEquipment(query) → Results, IndexEquipment(equipment) → Result
	PROTO	MUST	Merged: catalog + listing + search
transaction/v1/transaction.proto
CreateHold(listingID, dates) → HoldID, ConfirmBooking(holdID) → Booking, CancelBooking(bookingID) → Result, CheckAvailability(listingID, dates) → Available, CreateDeal(equipmentID, buyerID, terms) → Deal, ConfirmDeal(dealID) → Result, InitiateEscrow(dealID, amount) → EscrowID, ReleaseEscrow(escrowID) → Result, RefundEscrow(escrowID) → Result
	PROTO	MUST	Merged: booking + deal + payment
content/v1/content.proto
UploadDocument(metadata) → PresignedURL, GetDocument(id) → Document, UploadMedia(metadata) → PresignedURL, ProcessMedia(id) → Result, GetEngagementStats(entityID) → Stats
	PROTO	MUST	Merged: document + media + engagement
communication/v1/communication.proto
SendMessage(roomID, content) → Message, GetMessages(roomID) → Messages, SendNotification(userID, template, channel) → Result, GetPreferences(userID) → Preferences
	PROTO	MUST	Merged: chat + notification

scripts/proto-gen.sh
buf generate all protos → pkg/gen/go/ and pkg/gen/ts/. Committed to repo.
	PROTO	MUST	
Trust Service Module (Merged: identity + integrity + review)
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
Update name, avatar (trigger content service upload), contact details. Emit user.profile.updated to Kafka.
	BE	MUST	
PUT /users/me/avatar
Request presigned upload URL from Content service. Store resulting MinIO URL on profile.
	BE	SHOULD	
GET /users/:id/public
Public profile: name, avatar, company name, verified badge, rating, reviews count.
	BE	MUST	
PUT /users/me/notification-preferences
Toggle push/SMS/email per event type. Stored in PG, read by Communication service.
	BE	SHOULD	
grpc: GetUser() / GetUserBatch() / GetCompany()
Internal gRPC endpoints consumed by Transaction, Communication, Admin services.
	BE	MUST	
Company & Verification Module
POST /companies
Create company profile. БИН/ИНН format validation (12-digit KZ BIN). Set status = pending.
	BE	MUST	
PUT /companies/me
Update company info. If already verified, changes require re-verification for legal fields.
	BE	MUST	
POST /companies/me/documents
Upload verification docs (БИН cert, charter, director ID). Get presigned URL from Content service, store metadata. Move status to under_review.
	BE	MUST	
GET /companies/me/verification-status
Return current verification state, reviewer notes, list of submitted docs.
	BE	MUST	
verification state machine
pending → under_review → verified / rejected. On verified: emit company.verified to Kafka. On rejected: emit company.rejected with reason.
	BE	MUST	
Review & Reputation Module
POST /reviews
Create review for company/equipment. Validates transaction history. Updates reputation score in-process.
	BE	MUST	
GET /reviews/:entityID
List reviews with pagination, filtering by rating.
	BE	MUST	
reputation calculation
Real-time score update on new review. Weighted by reviewer verification status, transaction history.
	BE	MUST	Previously async via Kafka, now in-process for consistency
DB Migrations
001_users.sql
users, sessions, otp_codes tables
	BE	MUST	
002_companies.sql
companies, company_documents, verification_history tables
	BE	MUST	
003_reviews.sql
reviews, reputation_scores tables
	BE	MUST	New: merged from separate review service

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
Service Integration Summary

Consolidated Services (9 total, down from 12):
┌─────────────────────┬─────────────────────────────────────────────┬──────────┐
│ Service             │ Contains                                    │ Status   │
├─────────────────────┼─────────────────────────────────────────────┼──────────┤
│ gateway             │ API routing (unchanged)                     │ Keep     │
│ trust-service       │ identity + integrity + review               │ Merge    │
│ inventory-service   │ catalog + listing + search                  │ Merge    │
│ transaction-service │ booking + deal + payment                    │ Merge    │
│ content-service     │ document + media + engagement               │ Merge    │
│ communication-svc   │ chat + notification                         │ Merge    │
│ marketplace-service │ service provider listings                   │ Keep     │
│ analytics-service   │ reporting, dashboards                       │ Keep     │
│ admin-service       │ back-office tools                           │ Keep     │
└─────────────────────┴─────────────────────────────────────────────┴──────────┘

Integration Benefits:
+ Performance: In-process calls vs network (10-100x faster cross-module)
+ Consistency: ACID transactions across related operations
+ Development: Single codebase per domain, faster feature delivery
+ Testing: Integration tests become unit tests
+ Operations: Fewer containers, simpler deployment matrix

Integration Risks:
- Scaling: Can't scale components independently (mitigate: module-level autoscaling)
- Blast radius: Bug affects entire domain (mitigate: circuit breakers per module)
- Complexity: Larger codebase per service (mitigate: strict internal boundaries)

Total scope: ~339 tracked tasks across 7 phases, 9 backend services, 1 frontend, shared packages, infra, and proto definitions. Phases 1–3 constitute the shippable MVP. Phases 4–7 add full-featured communication, intelligence, monetisation, and production hardening. Each phase produces a deployable, testable increment.

IEM ·
