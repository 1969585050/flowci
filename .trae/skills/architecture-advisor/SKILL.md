---
name: "architecture-advisor"
description: "Helps with technical architecture decisions and tech stack selection. Invoke when starting a new project, evaluating technologies, or making architectural choices."
---

# Architecture Advisor

Professional technical architecture decision support system.

## Decision Framework

### Tech Stack Selection Matrix

When selecting technology, evaluate on:

| Dimension | Weight | Evaluation Criteria |
|-----------|--------|---------------------|
| Team Familiarity | 30% | Existing knowledge, learning curve |
| Community Activity | 20% | GitHub stars, recent updates, issues |
| Ecosystem | 20% | Libraries, tools, integrations |
| Performance | 15% | Benchmarks, scalability |
| Security | 15% | CVE history, security features |

### Decision Process

1. **Gather Requirements**
   - Functional needs
   - Non-functional requirements
   - Constraints (time, budget, skills)

2. **Generate Options**
   - 2-4 viable alternatives
   - Include "status quo" option

3. **Evaluate**
   - Score each option
   - Document trade-offs

4. **Decide**
   - Select highest score
   - Document rationale

5. **Review**
   - Revisit after 6 months
   - Track actual vs predicted

---

## Technology Recommendations (2026)

### Desktop Application

| Layer | Recommended | Alternative |
|-------|-------------|-------------|
| Framework | Tauri 2.10+ | Electron |
| Frontend | Vue 3.5+ | React 19 |
| State | Pinia | Zustand |
| UI Library | Naive UI | shadcn/vue |
| Build | Vite 6+ | - |

**Why Tauri?**
- Smaller binary (10MB vs 150MB Electron)
- Native performance
- Lower memory usage
- Rust backend (memory safe)

### Web Backend (Go)

| Layer | Recommended | Alternative |
|-------|-------------|-------------|
| HTTP Framework | Chi 5.1+ | Gin, Echo |
| ORM | GORM | sqlx |
| Migration | golang-migrate | goose |
| Validation | go-playground/validator | asaskevich |
| Logging | zerolog | zap |
| Config | Viper | standard JSON |

**Why Chi?**
- Lightweight, no middleware bloat
- Compatible with net/http
- Easy to extend
- Good performance

### Web Backend (Node.js)

| Layer | Recommended | Alternative |
|-------|-------------|-------------|
| Runtime | Node.js 20+ | Bun, Deno |
| Framework | Fastify | NestJS, Express |
| ORM | Prisma | Drizzle |
| Validation | Zod | Yup |
| Logging | Pino | winston |

### Frontend (Vue)

| Layer | Recommended | Alternative |
|-------|-------------|-------------|
| Framework | Vue 3.5+ | - |
| Type System | TypeScript 5.7+ | - |
| Build | Vite 6+ | - |
| Router | Vue Router 4 | - |
| State | Pinia | Vuex |
| HTTP | Axios / ky | fetch |

### Databases

| Use Case | Recommended | Alternative |
|----------|-------------|-------------|
| Relational | PostgreSQL 16 | MySQL 8 |
| Document | MongoDB 7 | - |
| Cache | Redis 7 | Memcached |
| Time Series | InfluxDB | TimescaleDB |
| Search | Elasticsearch | Meilisearch |
| Embedded | SQLite | Badger |

**Why PostgreSQL?**
- ACID compliance
- Rich data types
- JSON support
- Full-text search
- Active community

### Container & DevOps

| Category | Recommended |
|---------|-------------|
| Container Runtime | Docker 24+ |
| Orchestration | Kubernetes 1.28+ |
| CI/CD | GitHub Actions |
| Container Registry | GitHub Container Registry |
| Monitoring | Prometheus + Grafana |
| Logging | Loki |
| Tracing | Jaeger |

---

## Architecture Patterns

### 1. Clean Architecture (Backend)

```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/          # HTTP Layer
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── router.go
│   ├── domain/       # Business Entities
│   │   ├── entities/
│   │   └── errors/
│   ├── service/      # Business Logic
│   │   └── usecases/
│   └── repository/   # Data Access
│       └── interfaces/
├── pkg/
│   └── shared/       # Shared packages
└── configs/
```

**Benefits**:
- Separation of concerns
- Testable business logic
- Independent of frameworks
- Clear dependency direction

### 2. Feature-Based Structure (Frontend)

```
src/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── composables/
│   │   ├── api/
│   │   ├── types/
│   │   └── index.ts
│   └── builds/
│       └── ...
├── shared/
│   ├── components/
│   ├── composables/
│   └── utils/
├── router/
└── App.vue
```

**Benefits**:
- Colocated related code
- Easy to find features
- Clear module boundaries
- Scalable structure

### 3. Tauri + Go Backend Architecture

```
┌─────────────────────────────────────┐
│         Desktop UI (Vue 3)          │
│   ┌─────────────────────────────┐   │
│   │    Tauri Commands (Rust)    │   │
│   │    HTTP Client → Go API     │   │
│   └─────────────────────────────┘   │
└─────────────────────────────────────┘
                    ↓ HTTP
┌─────────────────────────────────────┐
│        Go API Server (Chi)          │
│   ┌─────────────────────────────┐   │
│   │   Handlers → Services →     │   │
│   │   Repositories → Database   │   │
│   └─────────────────────────────┘   │
└─────────────────────────────────────┘
                    ↓
            Docker Daemon
```

---

## Non-Functional Requirements Guidelines

### Performance

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Response (p99) | <200ms | APM tool |
| Page Load | <2s | Lighthouse |
| Time to Interactive | <3s | Lighthouse |
| Memory Usage | <200MB | Profiling |

### Security

| Area | Requirement |
|------|-------------|
| Authentication | JWT with refresh |
| Authorization | RBAC/ABAC |
| Data | Encrypted at rest, TLS in transit |
| Secrets | Never in code, use vault |
| Dependencies | Regular audit, no critical CVEs |

### Reliability

| Metric | Target |
|--------|--------|
| Uptime | 99.9% (8.7h downtime/year) |
| Error Rate | <0.1% |
| Recovery Time | <15 min |
| Backup Frequency | Daily + WAL |

### Scalability

| Metric | Target |
|--------|--------|
| Concurrent Users | 1000+ |
| API Throughput | 1000 RPS |
| Database | Support sharding if needed |
| CDN | Static assets cached |

---

## Common Architecture Decisions

### 1. Monolith vs Microservices

**Choose Monolith when:**
- Small team (<10)
- Fast iteration needed
- Shared data model
- Simple deployment

**Choose Microservices when:**
- Large team (>20)
- Multiple independent deploys
- Different scaling needs
- Technology diversity needed

### 2. SQL vs NoSQL

**Choose SQL when:**
- Structured data
- ACID transactions needed
- Complex queries
- Data integrity critical

**Choose NoSQL when:**
- Unstructured/semi-structured
- High write throughput
- Flexible schema
- Horizontal scaling needed

### 3. REST vs GraphQL vs gRPC

**REST**:
- Pros: Simple, widely understood, good tooling
- Cons: Over/under-fetching, multiple roundtrips

**GraphQL**:
- Pros: Flexible queries,减少 roundtrips
- Cons: Complexity, caching harder

**gRPC**:
- Pros: Binary protocol, streaming, type-safe
- Cons: Browser support, debugging harder

---

## Documentation Template

### ADR (Architecture Decision Record)

\`\`\`markdown
# ADR-001: Use Tauri for Desktop Framework

## Status
Accepted

## Context
We need to build a desktop application that:
- Works on Windows, macOS, Linux
- Has native performance
- Has small binary size
- Uses existing web skills

## Decision
We will use Tauri 2.0 with Vue 3 frontend.

## Consequences

### Positive
- Smaller binary size (~10MB vs 150MB Electron)
- Lower memory usage
- Native performance
- Leverage existing Vue skills

### Negative
- Less mature ecosystem than Electron
- More complex build setup
- Debugging requires multiple tools

## Alternatives Considered

### Electron
- Larger binary size
- Higher memory usage
- More mature but heavier

### Flutter
- Different skill set required (Dart)
- Heavier UI toolkit
\`\`\`

---

## Usage

### Get Tech Recommendation

```
/arch recommend --category "desktop" --constraints "small-binary,windows-linux"
```

### Evaluate Architecture

```
/arch evaluate --architecture "clean-hexagonal" --language "go"
```

### Create ADR

```
/arch create-adr --title "Use Tauri for Desktop" --context "need cross-platform desktop"
```

---

## Review Checklist

- [ ] Tech stack matches team skills
- [ ] Architecture supports requirements
- [ ] Non-functional requirements achievable
- [ ] Trade-offs documented
- [ ] ADRs created for major decisions
- [ ] Scalability considered
- [ ] Security reviewed
