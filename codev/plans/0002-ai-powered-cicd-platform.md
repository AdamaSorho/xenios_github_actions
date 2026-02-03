# Plan 0002: AI-Powered CI/CD Platform with Claude Code

## Overview

This plan implements the AI-powered CI/CD platform as specified in `codev/specs/0002-ai-powered-cicd-platform.md`. The platform enables automated feature building, TDD enforcement, and multi-platform deployment using GitHub Actions with Claude Code.

## Prerequisites

Before starting implementation:

1. **Accounts & Access**
   - [ ] GitHub repository created
   - [ ] Supabase project created (staging + production)
   - [ ] Fly.io account with app created
   - [ ] Vercel account with project linked
   - [ ] Claude Max subscription ($200/month)

2. **Local Tools**
   - [ ] Node.js 24 LTS installed
   - [ ] Go 1.25 installed
   - [ ] Supabase CLI installed (`npm install -g supabase`)
   - [ ] Fly CLI installed (`brew install flyctl`)
   - [ ] Vercel CLI installed (`npm install -g vercel`)
   - [ ] Claude Code CLI installed (`npm install -g @anthropic-ai/claude-code`)

3. **Secrets Generated**
   - [ ] `CLAUDE_CODE_OAUTH_TOKEN` via `claude setup-token`
   - [ ] `FLY_API_TOKEN` via `fly tokens create deploy`
   - [ ] `VERCEL_TOKEN` from Vercel dashboard
   - [ ] `STAGING_DATABASE_URL` from Supabase
   - [ ] `PRODUCTION_DATABASE_URL` from Supabase

---

## Phase 1: Monorepo Foundation

**Goal**: Set up the monorepo structure with Turborepo and basic configuration.

### Tasks

#### 1.1 Initialize Monorepo

```bash
# Create root package.json with workspaces
npm init -y
```

**Files to create:**

- `package.json` - Root workspace configuration
- `turbo.json` - Turborepo pipeline configuration
- `.gitignore` - Comprehensive ignore rules
- `.nvmrc` - Node version (24)
- `.prettierrc` - Code formatting
- `.eslintrc.js` - Linting rules

#### 1.2 Create Directory Structure

```
mkdir -p apps/backend/{cmd/api,internal/{domain/{entities,repository},usecase,adapter/{handler,repository,presenter},infrastructure/{database,config,middleware}},pkg,migrations}
mkdir -p apps/web/src/{app,domain/{entities,repositories},application/usecases,infrastructure/{api,repositories},presentation/{components,hooks,contexts}}
mkdir -p apps/web/__tests__
mkdir -p apps/mobile/src/{domain/{entities,repositories},application/usecases,infrastructure/{api,repositories},presentation/{screens,components,navigation}}
mkdir -p apps/mobile/__tests__
mkdir -p packages/{shared-types/src,api-client/src/endpoints,ui-kit/src}
mkdir -p .github/{workflows,ISSUE_TEMPLATE}
```

#### 1.3 Configure Turborepo

**turbo.json:**
```json
{
  "$schema": "https://turbo.build/schema.json",
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", ".next/**", "build/**"]
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": []
    },
    "test:coverage": {
      "dependsOn": ["build"],
      "outputs": ["coverage/**"]
    },
    "lint": {
      "outputs": []
    },
    "typecheck": {
      "dependsOn": ["^build"],
      "outputs": []
    },
    "dev": {
      "cache": false,
      "persistent": true
    }
  }
}
```

### Acceptance Criteria

- [ ] `npm install` works from root
- [ ] `npx turbo run build` completes (even if apps are empty)
- [ ] Directory structure matches spec exactly

---

## Phase 2: Backend (Go) Setup

**Goal**: Scaffold the Go backend with Clean Architecture structure.

### Tasks

#### 2.1 Initialize Go Module

```bash
cd apps/backend
go mod init github.com/xenios/backend
```

#### 2.2 Create Entry Point

**apps/backend/cmd/api/main.go:**
```go
package main

import (
    "log"
    "os"

    "github.com/xenios/backend/internal/infrastructure/config"
    "github.com/xenios/backend/internal/infrastructure/database"
)

func main() {
    cfg := config.Load()

    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // TODO: Wire up dependencies and start server
    log.Printf("Server starting on port %s", cfg.Port)
}
```

#### 2.3 Create Domain Layer

**apps/backend/internal/domain/entities/user.go:**
```go
package entities

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID        uuid.UUID
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**apps/backend/internal/domain/repository/user.go:**
```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user data access
// NOTE: This is an INTERFACE - no database imports here!
type UserRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
    Create(ctx context.Context, user *entities.User) error
    Update(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### 2.4 Create Infrastructure Layer

**apps/backend/internal/infrastructure/database/postgres.go:**
```go
package database

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {
    return pgxpool.New(context.Background(), databaseURL)
}
```

**apps/backend/internal/adapter/repository/postgres_user.go:**
```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/xenios/backend/internal/domain/entities"
    "github.com/xenios/backend/internal/domain/repository"
)

type PostgresUserRepository struct {
    db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) repository.UserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    var u entities.User
    err := r.db.QueryRow(ctx,
        "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
        id,
    ).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &u, nil
}

// ... implement other methods
```

#### 2.5 Create First Migration

**apps/backend/migrations/000001_create_users_table.up.sql:**
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

**apps/backend/migrations/000001_create_users_table.down.sql:**
```sql
DROP TABLE IF EXISTS users;
```

#### 2.6 Create Dockerfile

**apps/backend/Dockerfile:**
```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]
```

#### 2.7 Create Fly.io Config

**apps/backend/fly.toml:**
```toml
app = "xenios-api"
primary_region = "iad"

[build]

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[env]
  PORT = "8080"
```

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (with placeholder tests)
- [ ] Clean Architecture layers are properly separated
- [ ] No ORM imports (only pgx/database-sql)
- [ ] Domain layer has zero external dependencies

---

## Phase 3: Web (Next.js) Setup

**Goal**: Scaffold the Next.js web app with Clean Architecture.

### Tasks

#### 3.1 Initialize Next.js App

```bash
cd apps/web
npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir
```

#### 3.2 Create Clean Architecture Structure

**apps/web/src/domain/entities/User.ts:**
```typescript
export interface User {
  id: string
  email: string
  name: string
  createdAt: Date
  updatedAt: Date
}
```

**apps/web/src/domain/repositories/UserRepository.ts:**
```typescript
import { User } from '../entities/User'

// Interface only - NO api client imports here!
export interface UserRepository {
  findById(id: string): Promise<User | null>
  findByEmail(email: string): Promise<User | null>
  create(user: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User>
  update(id: string, user: Partial<User>): Promise<User>
  delete(id: string): Promise<void>
}
```

**apps/web/src/application/usecases/GetUserUseCase.ts:**
```typescript
import { User } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'

export class GetUserUseCase {
  constructor(private userRepo: UserRepository) {}

  async execute(id: string): Promise<User | null> {
    return this.userRepo.findById(id)
  }
}
```

**apps/web/src/infrastructure/repositories/ApiUserRepository.ts:**
```typescript
import { User } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'
import { apiClient } from '@xenios/api-client'

export class ApiUserRepository implements UserRepository {
  async findById(id: string): Promise<User | null> {
    const response = await apiClient.get<User>(`/users/${id}`)
    return response.data ?? null
  }

  // ... implement other methods
}
```

#### 3.3 Configure Package.json Scripts

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "typecheck": "tsc --noEmit",
    "test": "jest",
    "test:coverage": "jest --coverage"
  }
}
```

### Acceptance Criteria

- [ ] `npm run build` succeeds
- [ ] `npm run test` passes
- [ ] `npm run typecheck` passes
- [ ] No database imports (no @supabase, pg, prisma, etc.)
- [ ] Infrastructure layer uses API client only

---

## Phase 4: Mobile (React Native) Setup

**Goal**: Scaffold the React Native app with Clean Architecture.

### Tasks

#### 4.1 Initialize React Native App

```bash
cd apps/mobile
npx create-expo-app . --template blank-typescript
```

#### 4.2 Mirror Clean Architecture from Web

Same structure as web app:
- `src/domain/` - Entities and repository interfaces
- `src/application/` - Use cases
- `src/infrastructure/` - API client repository implementations
- `src/presentation/` - Screens and components

#### 4.3 Configure for Monorepo

Update `metro.config.js` to resolve workspace packages.

### Acceptance Criteria

- [ ] `npm run start` launches Expo
- [ ] TypeScript compilation succeeds
- [ ] No database imports
- [ ] Shares types with `@xenios/shared-types`

---

## Phase 5: Shared Packages

**Goal**: Create shared packages for types, API client, and UI components.

### Tasks

#### 5.1 shared-types Package

**packages/shared-types/src/index.ts:**
```typescript
// User types shared across all apps
export interface User {
  id: string
  email: string
  name: string
  createdAt: string  // ISO string for JSON serialization
  updatedAt: string
}

// API response types
export interface ApiResponse<T> {
  data: T | null
  error: string | null
}

// Add more shared types as needed
```

#### 5.2 api-client Package

**packages/api-client/src/client.ts:**
```typescript
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

export interface ApiResponse<T> {
  data: T | null
  error: string | null
  ok: boolean
}

export const apiClient = {
  async get<T>(path: string): Promise<ApiResponse<T>> {
    const response = await fetch(`${API_BASE_URL}${path}`)
    if (!response.ok) {
      return { data: null, error: response.statusText, ok: false }
    }
    const data = await response.json()
    return { data, error: null, ok: true }
  },

  async post<T>(path: string, body: unknown): Promise<ApiResponse<T>> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!response.ok) {
      return { data: null, error: response.statusText, ok: false }
    }
    const data = await response.json()
    return { data, error: null, ok: true }
  },

  // ... put, delete methods
}
```

### Acceptance Criteria

- [ ] Packages can be imported from apps: `import { User } from '@xenios/shared-types'`
- [ ] API client works for basic requests
- [ ] TypeScript types are properly exported

---

## Phase 6: GitHub Actions - Claude Code Integration

**Goal**: Set up GitHub Actions workflows for Claude Code automation.

### Tasks

#### 6.1 Claude Assistant Workflow

**Create `.github/workflows/claude-assistant.yml`** as specified in spec.

#### 6.2 Claude Implement Workflow

**Create `.github/workflows/claude-implement.yml`** as specified in spec.

#### 6.3 Claude Fix Workflow

**Create `.github/workflows/claude-fix.yml`** (similar to implement, for bugs).

### Acceptance Criteria

- [ ] `@claude` mention triggers response
- [ ] `claude-implement` label triggers feature implementation
- [ ] `claude-fix` label triggers bug fix

---

## Phase 7: GitHub Actions - TDD Quality Gate

**Goal**: Enforce TDD and Clean Architecture compliance on all PRs.

### Tasks

#### 7.1 Create TDD Gate Workflow

**Create `.github/workflows/tdd-gate.yml`** as specified in spec, including:
- Test coverage checks (80% minimum)
- Clean Architecture validation
- No ORM checks
- No direct database access in Web/Mobile

#### 7.2 Create dependency-cruiser Config

**Create `.dependency-cruiser.js`** as specified in spec.

### Acceptance Criteria

- [ ] PRs with <80% coverage fail
- [ ] PRs with ORM imports fail
- [ ] PRs with architecture violations fail
- [ ] PRs with database imports in Web/Mobile fail

---

## Phase 8: GitHub Actions - Deployment

**Goal**: Set up automated deployments to Fly.io and Vercel.

### Tasks

#### 8.1 Migration Workflow

**Create `.github/workflows/migrate-db.yml`** as specified in spec.

#### 8.2 Backend Deployment

**Create `.github/workflows/deploy-backend.yml`** as specified in spec.

#### 8.3 Web Deployment

**Create `.github/workflows/deploy-web.yml`** as specified in spec.

#### 8.4 Mobile Build

**Create `.github/workflows/deploy-mobile.yml`** for build/test only.

### Acceptance Criteria

- [ ] Migrations run on staging automatically
- [ ] Migrations require approval for production
- [ ] Backend deploys to Fly.io on merge to main
- [ ] Web deploys to Vercel on merge to main
- [ ] Production requires manual approval

---

## Phase 9: Issue Templates

**Goal**: Create issue templates that guide Claude on implementation.

### Tasks

#### 9.1 Feature Request Template

**Create `.github/ISSUE_TEMPLATE/feature-request.yml`:**
```yaml
name: Feature Request (Claude Implement)
description: Request a new feature for Claude to implement
labels: ["claude-implement"]
body:
  - type: textarea
    id: summary
    attributes:
      label: Summary
      description: One-sentence description
    validations:
      required: true
  - type: textarea
    id: user-story
    attributes:
      label: User Story
      description: As a [role], I want [capability] so that [benefit]
    validations:
      required: true
  - type: textarea
    id: acceptance-criteria
    attributes:
      label: Acceptance Criteria
      description: List the criteria that must be met
    validations:
      required: true
  - type: checkboxes
    id: affected-apps
    attributes:
      label: Affected Apps
      options:
        - label: Backend (Go)
        - label: Web (Next.js)
        - label: Mobile (React Native)
  - type: textarea
    id: test-scenarios
    attributes:
      label: Test Scenarios
      description: Key test cases Claude should implement
```

#### 9.2 Bug Report Template

**Create `.github/ISSUE_TEMPLATE/bug-report.yml`** as specified in spec.

### Acceptance Criteria

- [ ] Templates appear when creating new issues
- [ ] Labels are auto-applied
- [ ] Claude can parse the structured format

---

## Phase 10: CLAUDE.md Configuration

**Goal**: Create the CLAUDE.md file with all rules and instructions.

### Tasks

#### 10.1 Create CLAUDE.md

**Create `CLAUDE.md`** at the repository root with all content from the spec:
- Clean Architecture rules per app
- Database rules (backend-only)
- TDD requirements
- Version verification rules
- Forbidden patterns

### Acceptance Criteria

- [ ] All rules from spec are included
- [ ] Examples are clear and correct
- [ ] Claude follows the rules when implementing features

---

## Phase 11: GitHub Secrets Configuration

**Goal**: Configure all required secrets in GitHub.

### Tasks

#### 11.1 Add Repository Secrets

Navigate to Repository → Settings → Secrets → Actions and add:

| Secret | Source |
|--------|--------|
| `CLAUDE_CODE_OAUTH_TOKEN` | `claude setup-token` |
| `FLY_API_TOKEN` | `fly tokens create deploy` |
| `VERCEL_TOKEN` | Vercel dashboard |
| `VERCEL_ORG_ID` | Vercel dashboard |
| `VERCEL_PROJECT_ID` | Vercel dashboard |
| `STAGING_DATABASE_URL` | Supabase dashboard |
| `PRODUCTION_DATABASE_URL` | Supabase dashboard |

#### 11.2 Configure GitHub Environments

Create environments with protection rules:
- **staging**: No protection (auto-deploy)
- **production**: Required reviewers, manual approval

### Acceptance Criteria

- [ ] All secrets are configured
- [ ] Production environment requires approval
- [ ] Workflows can access secrets

---

## Phase 12: Integration Testing

**Goal**: Verify the entire pipeline works end-to-end.

### Tasks

#### 12.1 Test Claude Assistant

1. Create a PR
2. Comment `@claude What does this code do?`
3. Verify Claude responds

#### 12.2 Test Feature Implementation

1. Create issue with `claude-implement` label
2. Wait for Claude to create PR
3. Verify PR passes TDD gate
4. Merge and verify deployment

#### 12.3 Test Bug Fix

1. Create issue with `claude-fix` label
2. Wait for Claude to create PR
3. Verify fix is correct

#### 12.4 Test Migration Pipeline

1. Add a new migration file
2. Verify it runs on staging
3. Approve production deployment
4. Verify schema is updated

### Acceptance Criteria

- [ ] Full cycle: Issue → Claude PR → Tests Pass → Deploy
- [ ] Migrations work correctly
- [ ] All quality gates function

---

## Implementation Order

```
Phase 1: Monorepo Foundation     ─┐
Phase 2: Backend (Go)             │  Week 1-2
Phase 3: Web (Next.js)            │
Phase 4: Mobile (React Native)   ─┘

Phase 5: Shared Packages         ─┐
Phase 6: Claude Code Workflows    │  Week 3
Phase 7: TDD Quality Gate        ─┘

Phase 8: Deployment Workflows    ─┐
Phase 9: Issue Templates          │  Week 4
Phase 10: CLAUDE.md              ─┘

Phase 11: GitHub Secrets         ─┐
Phase 12: Integration Testing     │  Week 5
                                 ─┘
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Claude generates code with wrong architecture | TDD gate blocks PRs with violations |
| Migrations break production | Staging runs first, production requires approval |
| OAuth token expires | Use `/install-github-app` for auto-refresh |
| Rate limits on Claude Max | Set reasonable `--max-turns` limits |
| ORM accidentally added | CI checks block any ORM imports |

---

## Success Metrics

1. **Automation Rate**: Target 80% of simple features auto-implemented
2. **TDD Compliance**: 100% of PRs must pass quality gate
3. **Deployment Frequency**: Daily deploys to staging
4. **Migration Safety**: Zero production incidents from migrations
5. **Test Coverage**: Maintained above 80% across all apps

---

**Status**: Ready for implementation
**Estimated Duration**: 5 weeks
**Author**: Architect
**Date**: 2026-02-03
