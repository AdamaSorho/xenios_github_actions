# Xenios Platform - Claude Code Instructions

## Architecture: Clean Architecture (MANDATORY)

This monorepo enforces Clean Architecture across all apps.
**Dependencies flow INWARD only.**

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLEAN ARCHITECTURE                          │
│                    (Applied to EACH app separately)                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐        │
│   │   Mobile    │      │     Web     │      │   Backend   │        │
│   │ (API Client)│─────▶│ (API Client)│─────▶│ (Database)  │        │
│   └─────────────┘      └─────────────┘      └─────────────┘        │
│         │                    │                    │                 │
│         ▼                    ▼                    ▼                 │
│   Infrastructure:      Infrastructure:      Infrastructure:        │
│   ApiUserRepository    ApiUserRepository    PostgresUserRepository │
│   (calls Backend API)  (calls Backend API)  (calls Supabase/pgx)   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

Web & Mobile: External dependency = Backend REST API
Backend: External dependency = Supabase (PostgreSQL)
```

### Layer Rules (All Apps)

1. **Domain Layer** (`domain/`)
   - Contains: Entities, Value Objects, Repository Interfaces
   - Dependencies: NONE (pure business logic)
   - Example: `User` entity, `UserRepository` interface
   - **Same across all apps** - can be shared via `packages/shared-types`

2. **Application Layer** (`application/` or `usecase/`)
   - Contains: Use Cases (application-specific business rules)
   - Dependencies: Domain only
   - Example: `CreateUserUseCase`, `AuthenticateUserUseCase`

3. **Infrastructure Layer** (`infrastructure/`)
   - Contains: External implementations
   - Dependencies: Domain, Application
   - **Differs per app:**

   | App | Infrastructure Contains | Example |
   |-----|------------------------|---------|
   | Backend (Go) | Database access (pgx, raw SQL) | `PostgresUserRepository` |
   | Web (Next.js) | API client (HTTP to Backend) | `ApiUserRepository` |
   | Mobile (RN) | API client (HTTP to Backend) | `ApiUserRepository` |

4. **Presentation Layer** (`presentation/`, `adapter/handler/`)
   - Contains: UI components, HTTP handlers, CLI
   - Dependencies: Domain, Application, Infrastructure
   - **Differs per app:**

   | App | Presentation Contains |
   |-----|----------------------|
   | Backend (Go) | REST API handlers |
   | Web (Next.js) | React components, pages |
   | Mobile (RN) | React Native screens, components |

## TDD Requirements (MANDATORY)

1. **Test First**: Write failing test before implementation
2. **Red-Green-Refactor**:
   - RED: Write failing test
   - GREEN: Minimum code to pass
   - REFACTOR: Clean up while tests pass
3. **Coverage**: Minimum 80% for all new code
4. **Test Naming**: `Test<Function>_<Scenario>_<Expected>`

### File Patterns

- Go tests: `*_test.go` in same package
- TS tests: `*.test.ts` or `*.spec.ts`
- Test fixtures: `testdata/` or `__fixtures__/`

## Database: Backend-Only Access (MANDATORY)

**ONLY the Backend (Go) accesses the database. Web and Mobile call the Backend API.**

### Clean Architecture + Data Access (Per App)

**Backend (Go) - Infrastructure = Database**
```
┌─────────────────────────────────────────────────────────────────────┐
│  DOMAIN: UserRepository interface { FindByID(id) (*User, error) }   │
├─────────────────────────────────────────────────────────────────────┤
│  APPLICATION: GetUserUseCase { repo.FindByID(id) }                  │
├─────────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE: PostgresUserRepository                             │
│    → db.QueryRow("SELECT * FROM users WHERE id = $1", id)          │
│    → Raw SQL with pgx (talks to Supabase)                          │
├─────────────────────────────────────────────────────────────────────┤
│  PRESENTATION: HTTP Handler                                         │
│    → GET /api/users/:id → calls GetUserUseCase                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Web & Mobile (TypeScript) - Infrastructure = API Client**
```
┌─────────────────────────────────────────────────────────────────────┐
│  DOMAIN: UserRepository interface { findById(id): Promise<User> }   │
├─────────────────────────────────────────────────────────────────────┤
│  APPLICATION: GetUserUseCase { repo.findById(id) }                  │
├─────────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE: ApiUserRepository                                  │
│    → apiClient.get('/api/users/' + id)                             │
│    → HTTP call to Backend (NO database access!)                    │
├─────────────────────────────────────────────────────────────────────┤
│  PRESENTATION: React Component / React Native Screen                │
│    → useQuery() → calls GetUserUseCase                             │
└─────────────────────────────────────────────────────────────────────┘
```

### Database Rules

**All Apps (Clean Architecture):**
1. **Repository interfaces in Domain** - Define WHAT, not HOW
2. **Repository implementations in Infrastructure** - External access here only
3. **Dependency injection** - Use cases receive interfaces, not concrete repos
4. **No cross-layer imports** - Domain never imports Infrastructure

**Backend Only:**
5. **Raw SQL only** - No ORMs (GORM, Ent, Prisma, etc.)
6. **Prepared statements** - Always use parameterized queries ($1, $2)
7. **Migrations** - Raw SQL files, applied via CI/CD

**Web & Mobile Only:**
8. **API Client only** - Infrastructure calls Backend API, never database
9. **No database libraries** - No @supabase, pg, prisma, typeorm, etc.
10. **Shared API client** - Use `@xenios/api-client` package

### Allowed Libraries

| App | Allowed | Forbidden |
|-----|---------|-----------|
| **Backend (Go)** | `database/sql`, `pgx`, `sqlx` | GORM, Ent, Bun, SQLBoiler |
| **Web (Next.js)** | `fetch`, `axios`, API client | **ALL database libs** |
| **Mobile (React Native)** | `fetch`, `axios`, API client | **ALL database libs** |

## Version Verification (MANDATORY)

**Your knowledge about package versions may be outdated. ALWAYS verify before using.**

Before specifying ANY version of a package, framework, or runtime:

1. **CHECK the official documentation** for the current LTS/stable version:
   - Node.js: https://nodejs.org/
   - Go: https://go.dev/dl/
   - Next.js: https://nextjs.org/docs
   - React Native: https://reactnative.dev/

2. **Check existing project files** for version hints:
   - `package.json` → `engines.node` field
   - `go.mod` → Go version
   - `.nvmrc` or `.node-version` → Node version

3. **When uncertain in GitHub Actions**:
   - Post a comment on the issue/PR explaining the uncertainty
   - Use the most recent stable/LTS version from official docs
   - Document the assumption in the PR description

## Forbidden Patterns

**Clean Architecture Violations (All Apps):**
- Domain layer importing from Infrastructure
- Use Cases importing external specifics (HTTP, DB, API client)
- Use Cases directly instantiating repository implementations
- Circular dependencies between modules
- Business logic in handlers/controllers/components

**Backend (Go) Violations:**
- Using any ORM (GORM, Ent, Prisma, TypeORM, Drizzle, etc.)
- Raw SQL without parameterized queries (SQL injection risk)
- Database imports (pgx, sql) in domain or usecase layers

**Web & Mobile (TypeScript) Violations:**
- Importing ANY database library (@supabase, pg, prisma, mysql, mongodb, etc.)
- Accessing database directly (bypassing Backend API)
- API client imports in domain or application layers

## Before Making Changes

1. Identify which layer the change belongs to
2. Write tests for the expected behavior
3. Implement the minimum code to pass tests
4. Verify no layer violations introduced
5. Run full test suite before committing

## Project Structure

```
xenios/
├── apps/
│   ├── backend/           # Go API (Clean Architecture)
│   │   ├── cmd/api/       # Entry point
│   │   ├── internal/
│   │   │   ├── domain/    # Entities, repository interfaces
│   │   │   ├── usecase/   # Application business logic
│   │   │   ├── adapter/   # Handlers, repository implementations
│   │   │   └── infrastructure/  # Database, config
│   │   └── migrations/    # Raw SQL migrations
│   ├── web/               # Next.js (Clean Architecture)
│   │   └── src/
│   │       ├── domain/    # Entities, repository interfaces
│   │       ├── application/   # Use cases
│   │       ├── infrastructure/  # API repositories
│   │       └── presentation/    # Components, hooks
│   └── mobile/            # React Native (Clean Architecture)
│       └── src/           # Same structure as web
├── packages/
│   ├── shared-types/      # Common TypeScript types
│   ├── api-client/        # HTTP client for Backend API
│   └── ui-kit/            # Shared UI components
└── .github/
    └── workflows/         # CI/CD automation
```

## Git Workflow

**NEVER use `git add -A` or `git add .`** - Always add files explicitly.

Commit messages format:
```
[Feature] Brief description
[Bug] Fix description
[Chore] Maintenance description
```
