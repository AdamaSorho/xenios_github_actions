# Spec 0002: AI-Powered CI/CD Platform with Claude Code

## Summary

A comprehensive CI/CD platform using GitHub Actions with Claude Code to automatically build features, fix bugs, and deploy changes across a monorepo containing a Go backend, Next.js web app, and React Native mobile app. Enforces Clean Architecture and Test-Driven Development throughout.

## Goals

### Primary Goals

1. **Automated Feature Building**: When issues are created with proper labels, Claude Code automatically implements features following specs
2. **TDD Enforcement**: All AI-generated code must follow test-first development with quality gates
3. **Clean Architecture Compliance**: Enforce architectural boundaries across all three codebases
4. **Multi-Platform Deployment**: Automated deployment to Vercel (web) and Fly.io (backend)
5. **Human-in-the-Loop**: Require human approval for production deployments while automating dev/staging

### Non-Goals

- Building a custom AI model (we use Claude Code as-is)
- Real-time mobile app deployment (app stores require manual release)
- Self-modifying CI/CD workflows

## Architecture

### Monorepo Structure

```
xenios/
├── .github/
│   └── workflows/
│       ├── claude-assistant.yml       # Interactive @claude responses
│       ├── claude-implement.yml       # Feature implementation automation
│       ├── claude-fix.yml             # Bug fix automation
│       ├── tdd-gate.yml               # TDD enforcement quality gate
│       ├── deploy-backend.yml         # Fly.io deployment
│       ├── deploy-web.yml             # Vercel deployment
│       └── deploy-mobile.yml          # Mobile build & test
├── apps/
│   ├── backend/                       # Go API (Clean Architecture)
│   │   ├── cmd/
│   │   │   └── api/
│   │   │       └── main.go            # Entry point
│   │   ├── internal/
│   │   │   ├── domain/                # Entities & business rules
│   │   │   │   ├── entities/
│   │   │   │   └── repository/        # Repository interfaces
│   │   │   ├── usecase/               # Application business logic
│   │   │   ├── adapter/               # Interface adapters
│   │   │   │   ├── handler/           # HTTP handlers
│   │   │   │   ├── repository/        # Repository implementations
│   │   │   │   └── presenter/         # Response formatters
│   │   │   └── infrastructure/        # Frameworks & drivers
│   │   │       ├── database/          # Supabase/PostgreSQL (raw SQL, no ORM)
│   │   │       ├── config/
│   │   │       └── middleware/
│   │   ├── pkg/                       # Shared utilities
│   │   ├── migrations/                # Raw SQL migrations (no ORM)
│   │   │   ├── 000001_create_users.up.sql
│   │   │   ├── 000001_create_users.down.sql
│   │   │   └── ...
│   │   ├── Dockerfile
│   │   ├── fly.toml
│   │   └── go.mod
│   ├── web/                           # Next.js app (Clean Architecture)
│   │   ├── src/
│   │   │   ├── app/                   # Next.js App Router
│   │   │   ├── domain/                # Entities & business rules
│   │   │   │   ├── entities/
│   │   │   │   └── repositories/      # Repository interfaces
│   │   │   ├── application/           # Use cases
│   │   │   │   └── usecases/
│   │   │   ├── infrastructure/        # External implementations
│   │   │   │   ├── api/               # API clients
│   │   │   │   └── repositories/      # Repository implementations
│   │   │   └── presentation/          # UI components
│   │   │       ├── components/
│   │   │       ├── hooks/
│   │   │       └── contexts/
│   │   ├── __tests__/
│   │   ├── vercel.json
│   │   └── package.json
│   └── mobile/                        # React Native (Clean Architecture)
│       ├── src/
│       │   ├── domain/
│       │   ├── application/
│       │   ├── infrastructure/
│       │   └── presentation/
│       │       ├── screens/
│       │       ├── components/
│       │       └── navigation/
│       ├── __tests__/
│       ├── app.json
│       └── package.json
├── packages/                          # Shared code across apps
│   ├── shared-types/                  # TypeScript types (shared between all apps)
│   ├── api-client/                    # API client for Web/Mobile to call Backend
│   │   └── src/
│   │       ├── client.ts              # HTTP client (fetch/axios wrapper)
│   │       ├── endpoints/             # Typed API endpoints
│   │       └── types.ts               # Request/Response types
│   └── ui-kit/                        # Shared UI components (Web + Mobile)
├── CLAUDE.md                          # Claude Code instructions
├── turbo.json                         # Turborepo config
└── package.json                       # Root workspace
```

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Frameworks & Drivers                      │
│  (HTTP, DB, UI Frameworks, External APIs)                   │
├─────────────────────────────────────────────────────────────┤
│                    Interface Adapters                        │
│  (Controllers, Gateways, Presenters, Repository Impls)      │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│  (Use Cases - Application Business Logic)                   │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                            │
│  (Entities, Business Rules, Repository Interfaces)          │
└─────────────────────────────────────────────────────────────┘

DEPENDENCY RULE: Dependencies point INWARD only.
Inner layers know nothing about outer layers.
```

### Workflow Triggers

| Trigger | Workflow | Action |
|---------|----------|--------|
| Issue labeled `claude-implement` | claude-implement.yml | Claude implements feature from issue spec |
| Issue labeled `claude-fix` | claude-fix.yml | Claude fixes bug described in issue |
| PR comment `@claude` | claude-assistant.yml | Claude responds/reviews/helps |
| PR to main | tdd-gate.yml | Enforce TDD compliance |
| Merge to main | deploy-*.yml | Auto-deploy to staging |
| Release tag | deploy-*.yml | Deploy to production (with approval) |

## Technical Implementation

### 1. GitHub Actions: Claude Code Integration

#### claude-implement.yml - Automated Feature Implementation

```yaml
name: Claude Feature Implementation

on:
  issues:
    types: [labeled]

jobs:
  implement-feature:
    if: github.event.label.name == 'claude-implement'
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
      issues: write

    steps:
      - uses: actions/checkout@v6

      - name: Setup Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '24'

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'  # Stable as of Feb 2026, verify at go.dev/dl

      - name: Install dependencies
        run: npm ci

      - name: Claude Implements Feature
        uses: anthropics/claude-code-action@v1
        with:
          # Using Claude Max subscription via OAuth token (flat $200/month)
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          prompt: |
            ISSUE: #${{ github.event.issue.number }}
            TITLE: ${{ github.event.issue.title }}

            You are implementing a feature for the Xenios platform.

            ## Instructions

            1. READ the issue description thoroughly
            2. IDENTIFY which app(s) need changes (backend, web, mobile)
            3. FOLLOW Test-Driven Development:
               - Write failing tests FIRST
               - Implement minimum code to pass tests
               - Refactor while keeping tests green
            4. ENFORCE Clean Architecture:
               - Domain layer: entities and repository interfaces
               - Application layer: use cases only
               - Infrastructure: external implementations
               - Presentation: UI/API handlers
            5. CREATE a PR with:
               - Clear description of changes
               - Test coverage report
               - Screenshots/recordings for UI changes

            ## Architecture Rules

            - Dependencies flow INWARD only
            - Domain layer has NO external dependencies
            - Use cases call repository interfaces, not implementations
            - Handlers/Controllers are thin - delegate to use cases

            ## Database Rules (CRITICAL)

            - ONLY the Backend (Go) accesses the database (Supabase/PostgreSQL)
            - Web and Mobile NEVER access database directly - they call Backend APIs
            - NO ORMs allowed - use raw SQL with pgx/sqlx
            - Repository INTERFACES in domain layer (no DB imports)
            - Repository IMPLEMENTATIONS in infrastructure layer
            - FORBIDDEN in Backend: GORM, Prisma, TypeORM, Drizzle, Ent
            - FORBIDDEN in Web/Mobile: ANY database library (@supabase, pg, etc.)
            - Always use parameterized queries ($1, $2) for SQL injection prevention

            ## TDD Rules

            - Every new function needs a test
            - Test files: `*_test.go` (Go), `*.test.ts` (TypeScript)
            - Minimum 80% coverage for new code

            ## Version Verification (CRITICAL)

            Your knowledge about package versions may be OUTDATED.
            Before using ANY specific version:
            1. Check existing project files (package.json, go.mod, .nvmrc)
            2. Use WebSearch to verify current LTS/stable versions
            3. Check official docs (nodejs.org, go.dev, npmjs.com)
            4. If uncertain, post a comment on the issue asking for clarification
               before proceeding, or document your assumption in the PR
          claude_args: |
            --max-turns 50
            --model claude-opus-4-5-20251101
            --allowedTools Bash,Edit,Read,Write,Glob,Grep

      - name: Add implementation label
        if: success()
        run: |
          gh issue edit ${{ github.event.issue.number }} --add-label "implementation-complete"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### claude-assistant.yml - Interactive Code Assistant

```yaml
name: Claude Assistant

on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]
  issues:
    types: [opened, assigned]
  pull_request_review:
    types: [submitted]

jobs:
  claude-response:
    if: |
      contains(github.event.comment.body, '@claude') ||
      (github.event_name == 'issues' && github.event.action == 'assigned' && github.event.assignee.login == 'claude-bot')
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
      issues: write

    steps:
      - uses: actions/checkout@v6

      - uses: anthropics/claude-code-action@v1
        with:
          # Using Claude Max subscription via OAuth token
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          trigger_phrase: "@claude"
          claude_args: |
            --max-turns 20
            --model claude-opus-4-5-20251101
```

#### tdd-gate.yml - TDD Enforcement Quality Gate

```yaml
name: TDD Quality Gate

on:
  pull_request:
    branches: [main, develop]

jobs:
  tdd-enforcement:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0  # Full history for diff

      - name: Setup Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '24'

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'  # Stable as of Feb 2026, verify at go.dev/dl

      - name: Install dependencies
        run: npm ci

      # Backend TDD Check
      - name: Go - Check test coverage
        working-directory: apps/backend
        run: |
          go test -coverprofile=coverage.out ./...
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Backend coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "::error::Backend test coverage ($COVERAGE%) is below 80%"
            exit 1
          fi

      - name: Go - Verify Clean Architecture
        working-directory: apps/backend
        run: |
          # Check domain has no external imports
          if grep -r "infrastructure\|adapter" internal/domain/; then
            echo "::error::Domain layer must not import from outer layers"
            exit 1
          fi

          # Check usecase only imports domain
          if grep -r "infrastructure\|adapter/handler" internal/usecase/; then
            echo "::error::Usecase layer can only import from domain"
            exit 1
          fi

      - name: Go - No ORM Check
        working-directory: apps/backend
        run: |
          # Detect forbidden ORM imports
          FORBIDDEN_ORMS="gorm.io|entgo.io|github.com/uptrace/bun|github.com/volatiletech/sqlboiler"
          if grep -rE "$FORBIDDEN_ORMS" --include="*.go" .; then
            echo "::error::ORM detected! Use raw SQL with pgx/sqlx only. No GORM, Ent, Bun, or SQLBoiler."
            exit 1
          fi

      - name: Go - Database in Infrastructure Only
        working-directory: apps/backend
        run: |
          # Check that pgx/database imports only exist in infrastructure/adapter layers
          if grep -rE "github.com/jackc/pgx|database/sql" internal/domain/ internal/usecase/; then
            echo "::error::Database imports found outside infrastructure layer!"
            echo "Domain and UseCase layers must not import database packages."
            echo "Define repository INTERFACES in domain, IMPLEMENTATIONS in infrastructure."
            exit 1
          fi

      # Web TDD Check
      - name: Web - Run tests with coverage
        working-directory: apps/web
        run: |
          npm run test:coverage
          COVERAGE=$(npx nyc report --reporter=text-summary | grep "All files" | awk '{print $4}' | sed 's/%//')
          echo "Web coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "::error::Web test coverage ($COVERAGE%) is below 80%"
            exit 1
          fi

      - name: Web - No Direct Database Access
        working-directory: apps/web
        run: |
          # Web must NEVER access database directly - only consume Backend API
          FORBIDDEN_DEPS='@supabase/supabase-js|prisma|typeorm|drizzle-orm|sequelize|@mikro-orm|objection|pg|mysql|mongodb'
          if grep -E "$FORBIDDEN_DEPS" package.json; then
            echo "::error::Database library detected in Web app!"
            echo "Web must NOT access database directly."
            echo "Use the Backend API via @xenios/api-client instead."
            exit 1
          fi

      # Mobile TDD Check
      - name: Mobile - Run tests with coverage
        working-directory: apps/mobile
        run: |
          npm run test:coverage

      - name: Mobile - No Direct Database Access
        working-directory: apps/mobile
        run: |
          # Mobile must NEVER access database directly - only consume Backend API
          FORBIDDEN_DEPS='@supabase/supabase-js|prisma|typeorm|drizzle-orm|sequelize|@mikro-orm|objection|pg|mysql|mongodb'
          if grep -E "$FORBIDDEN_DEPS" package.json; then
            echo "::error::Database library detected in Mobile app!"
            echo "Mobile must NOT access database directly."
            echo "Use the Backend API via @xenios/api-client instead."
            exit 1
          fi

      # Architecture Lint
      - name: Lint architecture dependencies
        run: |
          npx dependency-cruiser --config .dependency-cruiser.js apps/

      - name: TDD Compliance Report
        uses: anthropics/claude-code-action@v1
        with:
          # Using Claude Max subscription via OAuth token
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          prompt: |
            Review this PR for TDD compliance:

            1. Are tests written BEFORE implementation? (check commit order)
            2. Do tests cover edge cases?
            3. Is test coverage adequate?
            4. Does code follow Clean Architecture?

            Provide a compliance score (0-100) and specific feedback.
          claude_args: |
            --max-turns 5
            --json-schema '{"type":"object","properties":{"score":{"type":"number"},"compliant":{"type":"boolean"},"issues":{"type":"array","items":{"type":"string"}},"suggestions":{"type":"array","items":{"type":"string"}}},"required":["score","compliant"]}'
```

### 2. Deployment Workflows

#### deploy-backend.yml - Fly.io Deployment

```yaml
name: Deploy Backend

on:
  push:
    branches: [main]
    paths:
      - 'apps/backend/**'
  release:
    types: [published]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '1.25'  # Stable as of Feb 2026, verify at go.dev/dl
      - name: Run tests
        working-directory: apps/backend
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v5
        with:
          files: apps/backend/coverage.out

  deploy-staging:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: staging

    steps:
      - uses: actions/checkout@v6

      - name: Setup Fly.io
        uses: superfly/flyctl-action/setup-flyctl@master

      - name: Deploy to staging
        working-directory: apps/backend
        run: flyctl deploy --remote-only --config fly.staging.toml
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}

  deploy-production:
    needs: test
    if: github.event_name == 'release'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.xenios.app

    steps:
      - uses: actions/checkout@v6

      - name: Setup Fly.io
        uses: superfly/flyctl-action/setup-flyctl@master

      - name: Deploy to production
        working-directory: apps/backend
        run: flyctl deploy --remote-only
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

#### migrate-db.yml - Database Migrations

**Migrations run BEFORE backend deployment to ensure schema is ready.**

```yaml
name: Database Migrations

on:
  push:
    branches: [main]
    paths:
      - 'apps/backend/migrations/**'
  workflow_dispatch:  # Allow manual trigger

jobs:
  migrate-staging:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: staging

    steps:
      - uses: actions/checkout@v6

      - name: Setup Supabase CLI
        uses: supabase/setup-cli@v1
        with:
          version: latest

      - name: Run Migrations (Staging)
        run: |
          supabase db push --db-url "$DATABASE_URL"
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}

  migrate-production:
    needs: migrate-staging
    runs-on: ubuntu-latest
    environment:
      name: production
      # Requires manual approval in GitHub Environment settings

    steps:
      - uses: actions/checkout@v6

      - name: Setup Supabase CLI
        uses: supabase/setup-cli@v1
        with:
          version: latest

      - name: Run Migrations (Production)
        run: |
          supabase db push --db-url "$DATABASE_URL"
        env:
          DATABASE_URL: ${{ secrets.PRODUCTION_DATABASE_URL }}
```

**Alternative: Using golang-migrate (Go-native)**

If you prefer a Go-native migration tool:

```yaml
      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'

      - name: Install golang-migrate
        run: |
          go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

      - name: Run Migrations
        run: |
          migrate -path apps/backend/migrations -database "$DATABASE_URL" up
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### Migration File Structure

```
apps/backend/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_user_email_index.up.sql
│   ├── 000002_add_user_email_index.down.sql
│   └── ...
├── cmd/
└── internal/
```

### Migration File Examples

**000001_create_users_table.up.sql**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- Enable Row Level Security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
```

**000001_create_users_table.down.sql** (rollback)
```sql
DROP TABLE IF EXISTS users;
```

### Migration Rules

1. **Always write both up AND down migrations** - Enable rollback
2. **One change per migration** - Easier to debug and rollback
3. **Never modify existing migrations** - Create new ones instead
4. **Test locally first** - Run against local Supabase before pushing
5. **No data in migrations** - Use seed files separately
6. **Use transactions** - Wrap DDL in BEGIN/COMMIT when possible

### Migration Workflow

```
Developer creates migration
        │
        ▼
PR includes migration file
        │
        ▼
CI runs migration on staging (automatic)
        │
        ▼
PR merged to main
        │
        ▼
Migration runs on production (manual approval required)
        │
        ▼
Backend deploys to Fly.io (uses new schema)
```

#### deploy-web.yml - Vercel Deployment

```yaml
name: Deploy Web

on:
  push:
    branches: [main]
    paths:
      - 'apps/web/**'
      - 'packages/**'
  pull_request:
    branches: [main]
    paths:
      - 'apps/web/**'
      - 'packages/**'
  release:
    types: [published]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-node@v6
        with:
          node-version: '24'
          cache: 'npm'

      - run: npm ci
      - run: npm run test --workspace=apps/web
      - run: npm run lint --workspace=apps/web
      - run: npm run typecheck --workspace=apps/web

  deploy-preview:
    needs: test
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v6

      - name: Deploy Preview
        uses: amondnet/vercel-action@v41
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          working-directory: ./apps/web

      - name: Comment Preview URL
        uses: actions/github-script@v8
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🚀 Preview deployed: ${{ steps.deploy.outputs.preview-url }}'
            })

  deploy-production:
    needs: test
    if: github.event_name == 'release'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://xenios.app

    steps:
      - uses: actions/checkout@v6

      - name: Deploy Production
        uses: amondnet/vercel-action@v41
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
          working-directory: ./apps/web
```

### 3. CLAUDE.md - Agent Instructions

The root `CLAUDE.md` must instruct Claude on architecture rules:

```markdown
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
   | Backend (Go) | REST API handlers (Gin/Echo) |
   | Web (Next.js) | React components, pages |
   | Mobile (RN) | React Native screens, components |

### TDD Requirements (MANDATORY)

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

### Before Making Changes

1. Identify which layer the change belongs to
2. Write tests for the expected behavior
3. Implement the minimum code to pass tests
4. Verify no layer violations introduced
5. Run full test suite before committing

### Version Verification (MANDATORY)

**Your knowledge about package versions may be outdated. ALWAYS verify before using.**

Before specifying ANY version of a package, framework, or runtime:

1. **CHECK the official documentation** for the current LTS/stable version:
   - Node.js: https://nodejs.org/
   - Go: https://go.dev/dl/
   - Next.js: https://nextjs.org/docs
   - React Native: https://reactnative.dev/
   - Any npm package: https://www.npmjs.com/package/<name>
   - Any Go module: https://pkg.go.dev/<module>

2. **Use WebFetch or WebSearch** to verify versions before writing code:
   ```
   Example: Before writing any node-version, search for "Node.js LTS version" + current year
   ```

3. **Check existing project files** for version hints:
   - `package.json` → `engines.node` field
   - `go.mod` → Go version
   - `.nvmrc` or `.node-version` → Node version
   - `Dockerfile` → Base image versions

4. **When uncertain in GitHub Actions** (can't ask interactively):
   - Post a comment on the issue/PR explaining the uncertainty
   - Use the most recent stable/LTS version from official docs
   - Document the assumption in the PR description

5. **Document the version** in comments when it matters:
   ```yaml
   node-version: '24'  # LTS as of Feb 2026, verify at nodejs.org
   ```

**Why this matters**: AI knowledge has a cutoff date. Package ecosystems move fast.
Using outdated versions can cause security vulnerabilities, compatibility issues, or
miss important features.

### Database: Backend-Only Access (MANDATORY)

**ONLY the Backend (Go) accesses the database. Web and Mobile call the Backend API.**

#### Clean Architecture + Data Access (Per App)

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
│  PRESENTATION: HTTP Handler (Gin/Echo)                              │
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

**Key Insight**: The repository interface is the SAME, but implementations differ:
- Backend: `PostgresUserRepository` → SQL queries
- Web/Mobile: `ApiUserRepository` → HTTP requests to Backend

#### Rules

**All Apps (Clean Architecture):**
1. **Repository interfaces in Domain** - Define WHAT, not HOW
2. **Repository implementations in Infrastructure** - External access here only
3. **Dependency injection** - Use cases receive interfaces, not concrete repos
4. **No cross-layer imports** - Domain never imports Infrastructure

**Backend Only:**
5. **Raw SQL only** - No ORMs (GORM, Ent, Prisma, etc.)
6. **Prepared statements** - Always use parameterized queries ($1, $2)
7. **Migrations** - Raw SQL files, applied via CI/CD (see Migrations section below)

**Web & Mobile Only:**
8. **API Client only** - Infrastructure calls Backend API, never database
9. **No database libraries** - No @supabase, pg, prisma, typeorm, etc.
10. **Shared API client** - Use `@xenios/api-client` package

#### Why No ORM?

- **Performance**: Raw SQL is faster, no query generation overhead
- **Control**: Full control over query optimization
- **Transparency**: Know exactly what queries hit the database
- **Supabase features**: Direct access to RLS, triggers, functions
- **Clean Architecture**: ORMs blur layer boundaries

#### Allowed Libraries

| App | Allowed | Forbidden |
|-----|---------|-----------|
| **Backend (Go)** | `database/sql`, `pgx`, `sqlx` | GORM, Ent, Bun, SQLBoiler |
| **Web (Next.js)** | `fetch`, `axios`, API client | **ALL database libs** (no Supabase, no Prisma, etc.) |
| **Mobile (React Native)** | `fetch`, `axios`, API client | **ALL database libs** (no Supabase, no Prisma, etc.) |

#### Architecture: Backend-Only Database Access

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Mobile    │     │     Web     │     │   Backend   │     ┌──────────┐
│ React Native│────▶│   Next.js   │────▶│     Go      │────▶│ Supabase │
│             │     │             │     │             │     │PostgreSQL│
└─────────────┘     └─────────────┘     └─────────────┘     └──────────┘
      │                   │                   │
      │   HTTP/REST API   │   HTTP/REST API   │   SQL (pgx)
      └───────────────────┴───────────────────┘

      Web & Mobile NEVER talk to database directly.
      All data flows through the Backend API.
```

**Why?**
- **Security**: Database credentials only on backend, not exposed in client bundles
- **Consistency**: Single source of truth for business logic
- **Flexibility**: Can change database without touching clients
- **Rate limiting**: Backend controls all database access

#### Example: Go Backend (Clean Architecture)

```go
// ─── DOMAIN LAYER: internal/domain/entity/user.go ───
type User struct {
    ID        uuid.UUID
    Email     string
    Name      string
    CreatedAt time.Time
}

// ─── DOMAIN LAYER: internal/domain/repository/user.go ───
// Interface only - NO database imports here!
type UserRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
}

// ─── APPLICATION LAYER: internal/usecase/get_user.go ───
type GetUserUseCase struct {
    userRepo repository.UserRepository  // Depends on INTERFACE
}

func (uc *GetUserUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    return uc.userRepo.FindByID(ctx, id)  // Calls interface method
}

// ─── INFRASTRUCTURE LAYER: internal/adapter/repository/postgres_user.go ───
// Implementation with actual Supabase/PostgreSQL details
type PostgresUserRepository struct {
    db *pgxpool.Pool
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    row := r.db.QueryRow(ctx,
        "SELECT id, email, name, created_at FROM users WHERE id = $1", id)
    var u entity.User
    err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
    return &u, err
}
```

#### Example: TypeScript Web/Mobile (Clean Architecture + API Client)

**Web and Mobile do NOT access the database. They consume the Backend API.**

```typescript
// ─── DOMAIN LAYER: src/domain/entities/User.ts ───
export interface User {
  id: string
  email: string
  name: string
  createdAt: Date
}

// ─── DOMAIN LAYER: src/domain/repositories/UserRepository.ts ───
// Interface only - defines what data operations are needed
export interface UserRepository {
  findById(id: string): Promise<User | null>
  create(user: Omit<User, 'id' | 'createdAt'>): Promise<User>
}

// ─── APPLICATION LAYER: src/application/usecases/GetUserUseCase.ts ───
export class GetUserUseCase {
  constructor(private userRepo: UserRepository) {}  // Depends on INTERFACE

  async execute(id: string): Promise<User | null> {
    return this.userRepo.findById(id)
  }
}

// ─── INFRASTRUCTURE LAYER: src/infrastructure/repositories/ApiUserRepository.ts ───
// Implementation calls BACKEND API (not database!)
import { apiClient } from '@xenios/api-client'  // Shared API client package

export class ApiUserRepository implements UserRepository {
  async findById(id: string): Promise<User | null> {
    // Calls Backend API, NOT database directly
    const response = await apiClient.get(`/users/${id}`)
    if (!response.ok) return null
    return response.data
  }

  async create(user: Omit<User, 'id' | 'createdAt'>): Promise<User> {
    const response = await apiClient.post('/users', user)
    return response.data
  }
}

// ❌ FORBIDDEN in Web/Mobile:
// import { createClient } from '@supabase/supabase-js'
// import { PrismaClient } from '@prisma/client'
// Any direct database access
```

#### Dependency Injection (Wiring It Together)

**Backend (Go) - Injects database repository:**
```go
// cmd/api/main.go
func main() {
    // Infrastructure: Connect to Supabase
    dbPool := connectToSupabase(os.Getenv("DATABASE_URL"))
    userRepo := repository.NewPostgresUserRepository(dbPool)  // SQL implementation

    // Application: Inject repository interface
    getUserUseCase := usecase.NewGetUserUseCase(userRepo)

    // Presentation: Wire up HTTP handlers
    handler := handler.NewUserHandler(getUserUseCase)
    router.GET("/api/users/:id", handler.GetUser)
}
```

**Web (Next.js) - Injects API client repository:**
```typescript
// src/infrastructure/container.ts
import { ApiUserRepository } from './repositories/ApiUserRepository'
import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'

// Infrastructure: API client (calls Backend, not database)
const userRepo = new ApiUserRepository()  // HTTP implementation

// Application: Inject repository interface
export const getUserUseCase = new GetUserUseCase(userRepo)

// Presentation: React hook uses the use case
// src/presentation/hooks/useUser.ts
export function useUser(id: string) {
  return useQuery(['user', id], () => getUserUseCase.execute(id))
}
```

**Same interface, different implementations:**
```
UserRepository (interface)
    ├── PostgresUserRepository (Backend) → SQL queries to Supabase
    └── ApiUserRepository (Web/Mobile)   → HTTP calls to Backend
```

### Forbidden Patterns

**Clean Architecture Violations (All Apps):**
- Domain layer importing from Infrastructure
- Use Cases importing external specifics (HTTP, DB, API client)
- Use Cases directly instantiating repository implementations
- Circular dependencies between modules
- Business logic in handlers/controllers/components
- Direct external calls from use cases (must go through repository interface)

**Backend (Go) Violations:**
- Using any ORM (GORM, Ent, Prisma, TypeORM, Drizzle, etc.)
- Raw SQL without parameterized queries (SQL injection risk)
- Database imports (pgx, sql) in domain or usecase layers
- Exposing database errors directly to clients

**Web & Mobile (TypeScript) Violations:**
- Importing ANY database library (@supabase, pg, prisma, mysql, mongodb, etc.)
- Accessing database directly (bypassing Backend API)
- API client imports in domain or application layers
- Hardcoding Backend API URLs (use environment config)
```

### 4. TDD Enforcement Tooling

#### dependency-cruiser.js - Architecture Validation

```javascript
module.exports = {
  forbidden: [
    // ─── CLEAN ARCHITECTURE RULES ───
    {
      name: 'domain-no-infrastructure',
      severity: 'error',
      comment: 'Domain layer cannot import from infrastructure',
      from: { path: 'domain' },
      to: { path: 'infrastructure' }
    },
    {
      name: 'domain-no-presentation',
      severity: 'error',
      comment: 'Domain layer cannot import from presentation/handlers',
      from: { path: 'domain' },
      to: { path: 'presentation|adapter/handler' }
    },
    {
      name: 'usecase-no-infrastructure',
      severity: 'error',
      comment: 'Use cases cannot import from infrastructure (use interfaces)',
      from: { path: 'usecase|application' },
      to: { path: 'infrastructure' }
    },
    {
      name: 'usecase-no-presentation',
      severity: 'error',
      comment: 'Use cases cannot import from presentation layer',
      from: { path: 'usecase|application' },
      to: { path: 'presentation|adapter/handler' }
    },

    // ─── DATABASE RULES ───
    // Web and Mobile must NEVER access database - only Backend does
    {
      name: 'web-no-database',
      severity: 'error',
      comment: 'Web app cannot import any database library. Use Backend API.',
      from: { path: 'apps/web' },
      to: { path: '@supabase|prisma|typeorm|drizzle|sequelize|pg|mysql|mongodb' }
    },
    {
      name: 'mobile-no-database',
      severity: 'error',
      comment: 'Mobile app cannot import any database library. Use Backend API.',
      from: { path: 'apps/mobile' },
      to: { path: '@supabase|prisma|typeorm|drizzle|sequelize|pg|mysql|mongodb' }
    },
    // Backend: No ORMs allowed, raw SQL only
    {
      name: 'no-orm-prisma',
      severity: 'error',
      comment: 'ORMs are forbidden. Use pgx/sqlx with raw SQL.',
      from: {},
      to: { path: '@prisma/client|prisma' }
    },
    {
      name: 'no-orm-typeorm',
      severity: 'error',
      comment: 'ORMs are forbidden. Use pgx/sqlx with raw SQL.',
      from: {},
      to: { path: 'typeorm' }
    },
    {
      name: 'no-orm-drizzle',
      severity: 'error',
      comment: 'ORMs are forbidden. Use pgx/sqlx with raw SQL.',
      from: {},
      to: { path: 'drizzle-orm' }
    },
    {
      name: 'no-orm-sequelize',
      severity: 'error',
      comment: 'ORMs are forbidden. Use pgx/sqlx with raw SQL.',
      from: {},
      to: { path: 'sequelize' }
    }
  ],
  options: {
    doNotFollow: {
      path: 'node_modules'
    }
  }
};
```

## Issue Templates

### Feature Request Template

```markdown
---
name: Feature Request (Claude Implement)
about: Request a new feature for Claude to implement
title: '[FEATURE] '
labels: claude-implement
---

## Summary
<!-- One-sentence description -->

## User Story
As a [role], I want [capability] so that [benefit].

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Technical Notes
<!-- Any architectural guidance, API contracts, etc. -->

## Affected Apps
- [ ] Backend (Go)
- [ ] Web (Next.js)
- [ ] Mobile (React Native)

## Test Scenarios
<!-- Describe key test cases Claude should implement -->
1. Happy path: ...
2. Edge case: ...
3. Error case: ...
```

### Bug Report Template

```markdown
---
name: Bug Report (Claude Fix)
about: Report a bug for Claude to fix
title: '[BUG] '
labels: claude-fix
---

## Description
<!-- What's broken? -->

## Steps to Reproduce
1. ...
2. ...
3. ...

## Expected Behavior
<!-- What should happen -->

## Actual Behavior
<!-- What happens instead -->

## Environment
- App: Backend / Web / Mobile
- Version:
- Browser/OS:

## Logs/Screenshots
<!-- Include relevant error messages -->
```

## Security Considerations

1. **Claude Max OAuth Token Management**
   - Store `CLAUDE_CODE_OAUTH_TOKEN` in GitHub Secrets
   - Generate token via `claude setup-token` locally
   - Use `/install-github-app` for automatic token refresh
   - Token expires ~7 days; GitHub App handles refresh automatically

2. **Deployment Tokens**
   - `FLY_API_TOKEN` and `VERCEL_TOKEN` in secrets
   - Scope tokens to minimum required permissions
   - Use GitHub Environments for production approvals

3. **Code Review**
   - All AI-generated PRs require human review
   - Production deployments require manual approval
   - Sensitive files (`.env`, credentials) excluded from Claude's scope

4. **Rate Limiting**
   - Claude Max has generous rate limits (not per-token billing)
   - Set `--max-turns` to prevent runaway execution time
   - Monitor usage via Claude dashboard

## Costs & Limits

| Resource | Limit | Cost Estimate |
|----------|-------|---------------|
| Claude Max Subscription | Unlimited Opus 4.5 (rate limited) | **$200/month flat** |
| Supabase (Database) | Free tier: 500MB, Pro: 8GB | $0-25/month |
| Fly.io (Backend) | 3 shared-cpu-1x | ~$15-30/month |
| Vercel (Web) | Pro plan | ~$20/month |
| GitHub Actions | 3000 min/month (free tier) | $0-50/month |

**Total estimated: ~$235-325/month** (vs $400-1000+ with API billing)

### Claude Max vs API Billing

| Aspect | Claude Max ($200/mo) | API (Pay-per-token) |
|--------|----------------------|---------------------|
| Opus 4.5 access | Unlimited* | ~$15/1M in, $75/1M out |
| Predictable cost | Yes | No |
| Rate limits | Yes (generous) | No |
| Best for | Heavy automation | Light/variable usage |

*Rate limited but sufficient for most CI/CD workloads

## Success Metrics

1. **Automation Rate**: % of issues implemented by Claude without human intervention
2. **TDD Compliance**: % of PRs passing TDD quality gate on first attempt
3. **Deployment Frequency**: Deployments per week
4. **Lead Time**: Time from issue creation to production
5. **Test Coverage**: Maintained above 80% across all apps

## Prerequisites

### Claude Max Subscription Setup

1. **Subscribe to Claude Max** ($200/month) at https://claude.ai

2. **Install Claude Code CLI**:
   ```bash
   npm install -g @anthropic/claude-code
   ```

3. **Generate OAuth Token**:
   ```bash
   claude setup-token
   ```
   Copy the generated token.

4. **Install GitHub App** (recommended for auto-refresh):
   ```bash
   claude /install-github-app
   ```
   This sets up automatic token refresh so you don't need to manually regenerate every 7 days.

5. **Add Secret to GitHub**:
   - Go to Repository → Settings → Secrets → Actions
   - Add `CLAUDE_CODE_OAUTH_TOKEN` with the token value

### Required GitHub Secrets

| Secret | Source | Used By | Purpose |
|--------|--------|---------|---------|
| `CLAUDE_CODE_OAUTH_TOKEN` | `claude setup-token` | CI/CD | Claude Max authentication |
| `FLY_API_TOKEN` | `fly tokens create deploy` | CI/CD | Fly.io deployment |
| `VERCEL_TOKEN` | Vercel dashboard | CI/CD | Vercel deployment |
| `VERCEL_ORG_ID` | Vercel dashboard | CI/CD | Vercel org identifier |
| `VERCEL_PROJECT_ID` | Vercel dashboard | CI/CD | Vercel project identifier |
| `STAGING_DATABASE_URL` | Supabase dashboard | **Migrations** | Staging PostgreSQL connection |
| `PRODUCTION_DATABASE_URL` | Supabase dashboard | **Migrations** | Production PostgreSQL connection |
| `DATABASE_URL` | Supabase dashboard | **Backend runtime** | PostgreSQL connection (set in Fly.io) |
| `SUPABASE_URL` | Supabase dashboard | **Backend runtime** | Supabase API URL |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase dashboard | **Backend runtime** | Private API key |

**Important:**
- Database secrets are for **Backend only**
- Web and Mobile never access the database - they consume the Backend API
- Migrations use separate staging/production URLs for safety

## Rollout Plan

### Phase 1: Foundation (Week 1-2)
- Set up monorepo structure
- Configure GitHub Actions workflows
- Set up Fly.io and Vercel projects
- Create CLAUDE.md with architecture rules
- **Set up Claude Max OAuth token**

### Phase 2: TDD Pipeline (Week 3)
- Implement TDD quality gate
- Set up dependency-cruiser for architecture validation
- Create issue templates

### Phase 3: Claude Integration (Week 4)
- Install claude-code-action
- Test with simple issues
- Tune prompts based on results

### Phase 4: Production (Week 5+)
- Enable production deployments
- Monitor and iterate on Claude prompts
- Document lessons learned

## Open Questions

1. Should Claude be able to request clarification on issues, or reject unclear specs?
2. What's the maximum PR size Claude should produce before requiring split?
3. How do we handle Claude-generated code that passes tests but has performance issues?

## References

- [Claude Code Action](https://github.com/anthropics/claude-code-action)
- [Clean Architecture in Go](https://github.com/Ikhlashmulya/golang-clean-architecture)
- [Next.js Clean Architecture](https://github.com/nikolovlazar/nextjs-clean-architecture)
- [Fly.io Continuous Deployment](https://fly.io/docs/launch/continuous-deployment-with-github-actions/)
- [Vercel GitHub Actions](https://github.com/marketplace/actions/deploy-to-vercel-action)
- [TDD Enforcement for AI Agents](https://www.brgr.one/blog/ai-coding-agents-tdd-enforcement)

---

**Status**: conceived
**Author**: Architect
**Date**: 2026-02-03
