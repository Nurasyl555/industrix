# Industrix - Industrial Equipment Marketplace

A comprehensive digital platform for listing, searching, buying, and renting industrial equipment and related services in the CIS region.

https://industrix-ztktsn.tiiny.site/

## Overview

Industrix is a microservices-based marketplace platform designed for industrial equipment transactions. The platform serves:

- **Industrial enterprises** seeking to buy or rent equipment
- **Equipment suppliers** listing new and used machinery
- **Service companies** offering delivery, installation, and maintenance
- **Construction and energy companies**
- **Oil and gas sector businesses**
- **Manufacturing SMEs**

### Key Capabilities

- Equipment catalog with advanced search and filtering
- Listing management (sale and rental)
- Real-time messaging between buyers and sellers
- Deal lifecycle management
- Payment processing (Kaspi Pay, Halyk Bank, Uzcard/Humo)
- Document generation and e-signature
- Rating and review system
- Admin moderation panel

---

## Architecture

### Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | React / Next.js · TypeScript |
| **Backend** | Go · Gin/Fiber |
| **Databases** | PostgreSQL · MongoDB · Redis |
| **Search** | OpenSearch |
| **Message Broker** | Apache Kafka |
| **Object Storage** | MinIO |
| **API** | gRPC · REST · WebSocket |
| **Infrastructure** | Docker/Kubernetes · NGINX |

### Microservices (12 Services)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                            │
│   Web (Next.js)  ·  Mobile (React Native)  ·  3rd Party APIs  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Gateway & Edge Layer                        │
│         NGINX (TLS, Load Balancing)  ·  API Gateway           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Core Domain Services                       │
├─────────────────────────────────────────────────────────────────┤
│ Identity        │ Auth, Profile, Company & Verification       │
│ Catalog         │ Equipment CRUD, Categories, Comparison      │
│ Listing         │ Ad lifecycle, Pricing, Statistics           │
│ Search          │ Full-text search, Recommendations, Geo      │
│ Booking         │ Rental availability, Slot reservation        │
│ Deal            │ Transaction orchestrator                    │
│ Payment         │ Escrow, Invoices, Payment providers        │
│ Document        │ PDF generation (Gotenberg), ЭДО            │
│ Review          │ Post-deal ratings and reviews               │
│ Services Market │ Additional services & logistics              │
│ Chat            │ WebSocket messaging                         │
│ Notification    │ Push, SMS, Email, In-app                    │
│ Engagement      │ Favorites, Price history                    │
│ Integrity       │ Disputes, Subscriptions, Audit log          │
│ Media           │ Upload handling, Image processing           │
│ Analytics       │ Event aggregation, Dashboards              │
│ Admin           │ Moderation, Content management              │
└─────────────────────────────────────────────────────────────────┘
```

### KZ Data Residency Compliance

All services run on infrastructure physically located in Kazakhstan. Personal data does not transit to foreign SaaS services.

**Self-hosted Components:**
- Gotenberg (PDF generation)
- imgproxy (image transforms)
- Postal (SMTP server)
- PostgreSQL, MongoDB, Redis, Kafka, MinIO

**Permitted External APIs:**
- 2GIS API (geocoding/routing - addresses only)
- FCM/APNs (push delivery - transient tokens)
- Beeline KZ / Kcell (SMS delivery)

---

## Project Structure

```
industrix/
├── .github/workflows/          # CI/CD pipelines
│
├── backend/                    # Go microservices
│   ├── services/              # 16 microservices
│   │   ├── gateway/           # API Gateway (Fiber)
│   │   ├── identity/          # Auth & User management
│   │   ├── catalog/           # Equipment catalog
│   │   ├── listing/           # Ad management
│   │   ├── search/            # OpenSearch service
│   │   ├── booking/           # Availability & reservations
│   │   ├── deal/              # Transaction handling
│   │   ├── payment/           # Payments & escrow
│   │   ├── document/          # PDF & contracts
│   │   ├── review/            # Ratings & reviews
│   │   ├── services-marketplace/ # Additional services
│   │   ├── chat/              # WebSocket messaging
│   │   ├── notification/      # Multi-channel notifications
│   │   ├── engagement/        # Favorites & price history
│   │   ├── integrity/         # Disputes & subscriptions
│   │   ├── media/             # File uploads
│   │   ├── analytics/         # Metrics & dashboards
│   │   └── admin/             # Admin panel
│   │
│   ├── proto/                  # Protocol buffer definitions
│   ├── pkg/                   # Shared Go packages
│   ├── migrations/            # Database migrations
│   └── scripts/               # Build & deployment scripts
│
├── frontend/                   # Next.js web application
│   ├── src/app/              # App Router pages
│   ├── src/components/       # UI components
│   ├── src/lib/              # API clients, hooks
│   └── src/types/            # TypeScript definitions
│
├── infra/                     # Infrastructure configs
│   ├── nginx/                # NGINX configuration
│   ├── postgres/             # PostgreSQL init scripts
│   ├── kafka/                # Kafka topics
│   ├── opensearch/           # Search index mappings
│   ├── gotenberg/            # PDF service
│   ├── imgproxy/             # Image processing
│   └── grafana/              # Monitoring dashboards
│
├── docs/                      # Documentation
│   ├── api/                  # OpenAPI specs
│   ├── adr/                  # Architecture decisions
│   └── runbooks/             # Operational guides
│
├── docker-compose.yml         # Full local stack
├── docker-compose.infra.yml   # Infrastructure only
├── docker-compose.override.yml # Dev overrides
├── Makefile                   # Development commands
└── README.md
```

---

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for frontend development)
- Make

### Quick Start

1. **Clone the repository**

2. **Start the infrastructure stack**
   
```
bash
   make up
   # Or just infrastructure:
   docker-compose -f docker-compose.infra.yml up -d
   
```

3. **Verify services are running**
   
```
bash
   make health
   
```

4. **Access the application**
   - Web UI: http://localhost:3000
   - API Gateway: http://localhost:8080
   - MinIO Console: http://localhost:9001

### Development Commands

| Command | Description |
|---------|-------------|
| `make up` | Start all services |
| `make down` | Stop all services |
| `make logs` | View logs |
| `make test` | Run tests |
| `make lint` | Run linters |
| `make proto` | Generate protobuf code |
| `make migrate` | Run database migrations |
| `make seed` | Seed development data |
| `make health` | Check service health |

---

## API Documentation

REST API documentation is available in `docs/api/`:

- `gateway.yaml` - API Gateway endpoints
- `catalog.yaml` - Catalog service endpoints
- `identity.yaml` - Authentication endpoints

OpenAPI specs can be imported into Swagger UI or Postman.

---

## Key Request Flows

### Equipment Search
1. Client sends search request → NGINX → API Gateway
2. Gateway validates JWT, routes to Search Service
3. Search Service checks Redis cache
4. Cache miss → queries OpenSearch
5. Results cached and returned

### Listing Creation
1. Seller creates listing via API Gateway → Listing Service
2. Listing Service validates category/attributes via Catalog Service
3. Client uploads media directly to MinIO (presigned URLs)
4. Media Service processes images asynchronously
5. Listing indexed in OpenSearch

### Real-time Chat
1. Client establishes WebSocket connection to Chat Service
2. Messages stored in MongoDB
3. Offline users receive push notifications via Notification Service

### Deal & Payment
1. Seller confirms deal → Deal Service
2. Payment Service initiates escrow hold
3. Deal status updates propagate via Kafka
4. Document Service generates contract PDF
5. Payment completion releases escrow

---

## Monitoring & Observability

- **Metrics**: Prometheus + Grafana dashboards
- **Logging**: Loki with structured JSON logs
- **Tracing**: Jaeger distributed tracing
- **Alerting**: Configured for critical metrics

Access Grafana at http://localhost:3002 (default credentials: admin/admin).

---

## License

Proprietary - All rights reserved

---

## Contributing

1. Create a feature branch
2. Make changes and add tests
3. Ensure CI passes
4. Submit a pull request
5. Wait for code review

### Code Generation

After modifying protobuf definitions:
```
bash
make proto
```

This generates Go and TypeScript code from `.proto` files.
