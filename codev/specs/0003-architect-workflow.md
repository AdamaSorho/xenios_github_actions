# Spec 0003: AI Architect Workflow - PRD to Issues Automation

## Summary

A GitHub Action workflow that uses Claude Code to act as an "Architect" - analyzing PRD documents, breaking them down into well-scoped implementation tasks, and automatically creating GitHub issues that the existing Builder workflow can implement.

## Goals

### Primary Goals

1. **PRD Parsing**: Accept PRD documents in multiple formats (PDF, Markdown, plain text prompts)
2. **Intelligent Decomposition**: Break large features into implementable, well-scoped issues (~1-3 day tasks)
3. **Dependency Tracking**: Identify and mark issue dependencies (blocking/blocked-by relationships)
4. **Issue Quality**: Generate issues following the project's feature request template with clear acceptance criteria
5. **Human-in-the-Loop**: Allow human review before issues are created or provide draft mode

### Non-Goals

- Implementing features directly (that's the Builder's job)
- Modifying existing issues
- Managing sprints or project timelines

## Workflow Triggers

| Trigger | Use Case |
|---------|----------|
| `workflow_dispatch` with file upload | Upload PRD PDF/MD and generate issues |
| `workflow_dispatch` with prompt | Paste requirements text directly |
| Issue comment `@claude-architect` | Analyze issue and break into sub-tasks |
| PR comment `@claude-architect` | Review PR scope and suggest issue breakdown |

## Technical Implementation

### claude-architect.yml - Main Workflow

```yaml
name: Claude Architect - PRD to Issues

on:
  workflow_dispatch:
    inputs:
      prd_url:
        description: 'URL to PRD document (GitHub raw URL, or public URL)'
        required: false
        type: string
      prd_content:
        description: 'Paste PRD content directly (if no URL)'
        required: false
        type: string
      mode:
        description: 'Output mode'
        required: true
        type: choice
        options:
          - draft    # Creates draft issues for review
          - create   # Creates real issues immediately
          - preview  # Only outputs plan, no issues created
        default: 'preview'
      priority:
        description: 'Default priority for generated issues'
        required: false
        type: choice
        options:
          - high
          - medium
          - low
        default: 'medium'

  issue_comment:
    types: [created]

jobs:
  architect-analyze:
    if: |
      github.event_name == 'workflow_dispatch' ||
      (github.event_name == 'issue_comment' && contains(github.event.comment.body, '@claude-architect'))
    runs-on: ubuntu-latest
    permissions:
      contents: read
      issues: write
      pull-requests: write

    steps:
      - uses: actions/checkout@v6

      - name: Setup Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '24'

      - name: Fetch PRD Content (if URL provided)
        id: fetch-prd
        if: github.event.inputs.prd_url != ''
        run: |
          curl -sL "${{ github.event.inputs.prd_url }}" -o prd_content.txt
          echo "prd_file=prd_content.txt" >> $GITHUB_OUTPUT

      - name: Claude Architect Analysis
        uses: anthropics/claude-code-action@v1
        with:
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          prompt: |
            You are an AI Architect. Your job is to analyze product requirements and create well-structured GitHub issues for implementation.

            ## Input

            ${{ github.event.inputs.prd_content || '' }}
            ${{ github.event.inputs.prd_url && format('PRD loaded from: {0}', github.event.inputs.prd_url) || '' }}
            ${{ github.event_name == 'issue_comment' && format('Analyze this issue for breakdown: #{0}', github.event.issue.number) || '' }}

            ## Your Task

            1. **Analyze the Requirements**
               - Identify distinct features/capabilities
               - Note technical dependencies
               - Identify risk areas that need spikes

            2. **Break Down into Issues**
               - Each issue should be 1-3 days of work (Small/Medium scope)
               - Group related work logically
               - Order by dependencies (foundation first)

            3. **For Each Issue, Define:**
               - Clear, actionable title
               - User story (As a [role], I want [capability] so that [benefit])
               - Acceptance criteria (testable checkboxes)
               - Technical notes (architecture hints, API contracts)
               - Affected apps (Backend, Web, Mobile)
               - Test scenarios (happy path, edge cases, errors)
               - Dependencies (which issues must complete first)
               - Estimated scope: XS (<4hrs), S (4-8hrs), M (1-2 days), L (3-5 days)

            4. **Output Format**

               Mode: ${{ github.event.inputs.mode || 'preview' }}

               If mode is 'preview':
               - Output a markdown summary of all planned issues
               - Include dependency graph
               - Do NOT create any issues

               If mode is 'draft' or 'create':
               - Use the gh CLI to create issues
               - Add label: claude-implement
               - Add label based on scope: size/XS, size/S, size/M, size/L
               - Add priority label: priority/${{ github.event.inputs.priority || 'medium' }}
               - For 'draft' mode, create as draft issues if supported

            ## Architecture Rules (from CLAUDE.md)

            - Follow Clean Architecture (Domain → Application → Infrastructure → Presentation)
            - Backend (Go) accesses database; Web/Mobile call Backend API only
            - No ORMs - raw SQL with pgx/sqlx
            - TDD required - 80% coverage minimum
            - All issues must include test scenarios

            ## Issue Template to Follow

            ```markdown
            ## Summary
            <!-- One-sentence description -->

            ## User Story
            As a [role], I want [capability] so that [benefit].

            ## Acceptance Criteria
            - [ ] Criterion 1
            - [ ] Criterion 2
            - [ ] Criterion 3

            ## Technical Notes
            <!-- Architecture guidance, API contracts, data models -->

            ## Affected Apps
            - [ ] Backend (Go)
            - [ ] Web (Next.js)
            - [ ] Mobile (React Native)
            - [ ] Shared packages

            ## Test Scenarios
            1. **Happy path**: ...
            2. **Edge case**: ...
            3. **Error case**: ...

            ## Dependencies
            <!-- List issue numbers this depends on -->
            Blocked by: #XX, #YY

            ## Scope
            <!-- XS (<4hrs), S (4-8hrs), M (1-2 days), L (3-5 days) -->
            Size: M

            ---
            🤖 Generated by Claude Architect from PRD analysis
            ```

            ## Creating Issues (when mode != preview)

            Use the GitHub CLI to create issues:
            ```bash
            gh issue create \
              --title "TITLE" \
              --body "BODY" \
              --label "claude-implement,size/M,priority/medium"
            ```

            For issues with dependencies, after creating all issues, update them:
            ```bash
            gh issue edit ISSUE_NUMBER --body "Updated body with dependency links"
            ```

          claude_args: |
            --max-turns 100
            --model claude-opus-4-5-20251101
            --allowedTools Bash,Read,WebFetch

      - name: Post Summary
        if: always()
        uses: actions/github-script@v8
        with:
          script: |
            const summary = `## 🏗️ Architect Analysis Complete

            Mode: \`${{ github.event.inputs.mode || 'preview' }}\`

            Check the workflow logs for the full analysis and created issues.

            [View Workflow Run](${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }})
            `;

            if (context.eventName === 'issue_comment') {
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: summary
              });
            }
```

### Breaking Down the Xenios PRD

Based on the PRD you shared, here's how the Architect would break it into issues:

#### Week 1: Foundation (Issues 1-7)
```
Issue #1: Database Schema Setup
  - Create all PostgreSQL tables with RLS
  - Scope: M (1-2 days)
  - Dependencies: None

Issue #2: Authentication System
  - JWT tokens with refresh rotation
  - Scope: M
  - Dependencies: #1

Issue #3: pg-boss Job Queue Setup
  - Async job infrastructure
  - Scope: S (4-8 hrs)
  - Dependencies: #1

Issue #4: S3/R2 Storage with Presigned URLs
  - File upload infrastructure
  - Scope: S
  - Dependencies: None

Issue #5: Mobile App Skeleton (Auth + Navigation)
  - React Native foundation
  - Scope: M
  - Dependencies: #2

Issue #6: Web Dashboard Skeleton
  - Next.js with login
  - Scope: M
  - Dependencies: #2

Issue #7: CI/CD Pipeline Setup
  - GitHub Actions → staging
  - Scope: S
  - Dependencies: None
```

#### Week 2: Data Extraction (Issues 8-13)
```
Issue #8: IBM Docling Integration for PDF Extraction
  - InBody scans, lab results
  - Scope: L (3-5 days)
  - Dependencies: #1, #4

Issue #9: CSV/JSON Parsers for Wearables
  - Garmin, WHOOP, Apple Health
  - Scope: M
  - Dependencies: #1

Issue #10: Async Extraction Workers
  - Trigger extraction on upload
  - Scope: M
  - Dependencies: #3, #8, #9

Issue #11: Client Profile UI
  - Display extracted measurements
  - Scope: M
  - Dependencies: #5, #6, #10

Issue #12: Basic Insight Generation
  - Flag out-of-range values
  - Scope: M
  - Dependencies: #10

Issue #13: Coach Approval Queue UI
  - Draft insights with approve/reject
  - Scope: M
  - Dependencies: #6, #12
```

#### Week 3-6: Continues similarly...

## Dependency Graph Example

```
                    ┌──────────┐
                    │ #1 DB    │
                    └────┬─────┘
           ┌─────────────┼─────────────┐
           ▼             ▼             ▼
      ┌────────┐    ┌────────┐    ┌────────┐
      │ #2 Auth│    │ #3 Jobs│    │ #4 S3  │
      └───┬────┘    └───┬────┘    └───┬────┘
          │             │             │
    ┌─────┴─────┐       │             │
    ▼           ▼       │             │
┌──────┐   ┌──────┐     │             │
│#5 Mob│   │#6 Web│     │             │
└──────┘   └──────┘     │             │
                        ▼             │
                   ┌─────────────┐    │
                   │ #8 Docling  │◄───┘
                   └──────┬──────┘
                          │
                          ▼
                   ┌─────────────┐
                   │ #10 Workers │
                   └─────────────┘
```

## Security Considerations

1. **PRD Content Handling**
   - PRD URLs must be from trusted sources (GitHub, internal docs)
   - No automatic execution of code from PRDs
   - Sanitize issue content to prevent injection

2. **Issue Creation Limits**
   - Maximum 50 issues per run (prevent runaway creation)
   - Rate limiting between issue creation calls
   - Human approval required for 'create' mode (via branch protection)

3. **Sensitive Data**
   - PRDs may contain business-sensitive information
   - Issues are public by default - warn if repo is public
   - Option to create issues in private project board first

## Usage Examples

### 1. Upload PRD and Preview

```bash
# Trigger via GitHub UI or CLI
gh workflow run claude-architect.yml \
  -f prd_url="https://raw.githubusercontent.com/org/repo/main/docs/PRD.md" \
  -f mode="preview"
```

### 2. Paste Requirements Directly

```bash
gh workflow run claude-architect.yml \
  -f prd_content="Build a user authentication system with email/password and OAuth support" \
  -f mode="draft" \
  -f priority="high"
```

### 3. Break Down Existing Issue

Comment on any issue:
```
@claude-architect Please break this feature into smaller implementable tasks
```

## Success Metrics

| Metric | Target |
|--------|--------|
| Issue Quality Score | Issues require <2 clarification comments before implementation |
| Scope Accuracy | 80% of issues completed within estimated scope |
| Dependency Accuracy | <10% of issues blocked by missing dependencies |
| Builder Success Rate | 90% of architect-created issues successfully implemented |

## Integration with Existing Workflows

```
PRD Document
     │
     ▼
┌─────────────────────────┐
│  claude-architect.yml   │  ← NEW (this spec)
│  Analyzes requirements  │
│  Creates issues         │
└───────────┬─────────────┘
            │ Creates issues with
            │ label: claude-implement
            ▼
┌─────────────────────────┐
│  claude-implement.yml   │  ← EXISTING (Spec 0002)
│  Implements features    │
│  Creates PRs            │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  tdd-gate.yml           │  ← EXISTING (Spec 0002)
│  Validates quality      │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  deploy-*.yml           │  ← EXISTING (Spec 0002)
│  Deploys to production  │
└─────────────────────────┘
```

## Batch Orchestration (Hybrid Approach)

Issues are created in batches (foundation first, then dependents) with human control over batch progression.

### How It Works

1. **Master Plan Issue**: Architect creates a "Master Plan" issue that tracks all batches
2. **Batch 1 Immediate**: First batch (foundation) created immediately
3. **Event Monitoring**: Monitor workflow watches for issue close events
4. **Batch Completion**: When all issues in a batch are closed, notify human
5. **Human Triggers Next**: Human approves next batch via comment or workflow dispatch

### Master Plan Issue Template

```markdown
# 🏗️ Master Plan: [PRD Title]

**Generated from**: [PRD URL or description]
**Generated at**: [timestamp]
**Total Issues**: [N]
**Batches**: [M]

## Progress

- [x] **Batch 1: Foundation** (Issues #1-#7) ✅ Complete
- [ ] **Batch 2: Data Extraction** (Issues #8-#13) 🔄 In Progress
- [ ] **Batch 3: AI Integration** (Issues #14-#20) ⏳ Waiting
- [ ] **Batch 4: Polish & Deploy** (Issues #21-#25) ⏳ Waiting

## Batch Details

### Batch 1: Foundation (Created ✅)
| Issue | Title | Status |
|-------|-------|--------|
| #1 | Database Schema Setup | ✅ Closed |
| #2 | Authentication System | ✅ Closed |
| #3 | pg-boss Job Queue Setup | ✅ Closed |
| #4 | S3/R2 Storage with Presigned URLs | ✅ Closed |
| #5 | Mobile App Skeleton | ✅ Closed |
| #6 | Web Dashboard Skeleton | ✅ Closed |
| #7 | CI/CD Pipeline Setup | ✅ Closed |

### Batch 2: Data Extraction (Pending)
| Issue | Title | Dependencies |
|-------|-------|--------------|
| - | IBM Docling Integration | #1, #4 |
| - | CSV/JSON Parsers for Wearables | #1 |
| - | Async Extraction Workers | #3, Docling, Parsers |
| - | Client Profile UI | #5, #6, Workers |
| - | Basic Insight Generation | Workers |
| - | Coach Approval Queue UI | #6, Insights |

### Batch 3: AI Integration (Planned)
...

### Batch 4: Polish & Deploy (Planned)
...

---

## Commands

- Comment `@claude-architect next` to create the next batch
- Comment `@claude-architect status` to get current progress
- Trigger manually: [Create Next Batch](workflow_dispatch_link)

---
🤖 Generated by Claude Architect
```

### claude-architect-monitor.yml - Batch Progress Monitor

```yaml
name: Claude Architect - Batch Monitor

on:
  issues:
    types: [closed]

jobs:
  check-batch-completion:
    # Only run for issues created by the architect
    if: contains(github.event.issue.labels.*.name, 'claude-implement')
    runs-on: ubuntu-latest
    permissions:
      contents: read
      issues: write

    steps:
      - uses: actions/checkout@v6

      - name: Find Master Plan Issue
        id: find-master
        uses: actions/github-script@v8
        with:
          script: |
            // Find the Master Plan issue that tracks this batch
            const issues = await github.rest.issues.listForRepo({
              owner: context.repo.owner,
              repo: context.repo.repo,
              labels: 'architect-master-plan',
              state: 'open'
            });

            if (issues.data.length === 0) {
              console.log('No master plan issue found');
              return { found: false };
            }

            const masterPlan = issues.data[0];
            return {
              found: true,
              number: masterPlan.number,
              body: masterPlan.body
            };

      - name: Check Batch Completion
        if: steps.find-master.outputs.result != '{"found":false}'
        id: check-batch
        uses: actions/github-script@v8
        with:
          script: |
            const master = ${{ steps.find-master.outputs.result }};
            if (!master.found) return { complete: false };

            // Parse the master plan to find current batch issues
            const body = master.body;

            // Extract issue numbers from "In Progress" batch
            const inProgressMatch = body.match(/🔄 In Progress\n\|[^\n]+\n\|[^\n]+\n((?:\|[^\n]+\n?)+)/);
            if (!inProgressMatch) {
              console.log('No in-progress batch found');
              return { complete: false };
            }

            // Extract issue numbers from the table
            const issueNumbers = [...inProgressMatch[1].matchAll(/#(\d+)/g)].map(m => parseInt(m[1]));

            // Check if all issues in the batch are closed
            let allClosed = true;
            for (const num of issueNumbers) {
              const issue = await github.rest.issues.get({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: num
              });
              if (issue.data.state !== 'closed') {
                allClosed = false;
                break;
              }
            }

            return {
              complete: allClosed,
              masterNumber: master.number,
              batchIssues: issueNumbers
            };

      - name: Notify Batch Complete
        if: steps.check-batch.outputs.result.complete == true
        uses: actions/github-script@v8
        with:
          script: |
            const result = ${{ steps.check-batch.outputs.result }};

            // Comment on the master plan issue
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: result.masterNumber,
              body: `## 🎉 Batch Complete!

            All issues in the current batch have been closed.

            **Ready for next batch?**

            - Comment \`@claude-architect next\` to create the next batch of issues
            - Or [trigger manually](${{ github.server_url }}/${{ github.repository }}/actions/workflows/claude-architect.yml)

            ---
            🤖 Automated batch completion notification`
            });

      - name: Update Master Plan Status
        if: steps.check-batch.outputs.result.complete == true
        uses: anthropics/claude-code-action@v1
        with:
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          prompt: |
            Update the Master Plan issue #${{ fromJSON(steps.check-batch.outputs.result).masterNumber }} to:
            1. Mark the current batch as complete (change 🔄 to ✅)
            2. Mark the next batch as "In Progress" (change ⏳ to 🔄)
            3. Update the issue table with closed status for completed issues

            Use the gh CLI to update the issue body.
          claude_args: |
            --max-turns 10
            --allowedTools Bash,Read
```

### claude-architect.yml Updates for Batch Support

Add to the main workflow's prompt section:

```yaml
# Add to workflow_dispatch inputs
inputs:
  # ... existing inputs ...
  batch_action:
    description: 'Batch action (for ongoing projects)'
    required: false
    type: choice
    options:
      - new      # Start new project from PRD
      - next     # Create next batch for existing project
      - status   # Report current status
    default: 'new'
  master_plan_issue:
    description: 'Master Plan issue number (for next/status actions)'
    required: false
    type: string
```

Add batch handling to the prompt:

```yaml
prompt: |
  # ... existing prompt content ...

  ## Batch Creation Mode

  Action: ${{ github.event.inputs.batch_action || 'new' }}
  Master Plan Issue: ${{ github.event.inputs.master_plan_issue || 'N/A' }}

  If action is 'new':
  - Create the Master Plan issue first with label: architect-master-plan
  - Create only Batch 1 (foundation) issues
  - Document all future batches in the Master Plan

  If action is 'next':
  - Read the Master Plan issue to find the next pending batch
  - Create issues for that batch only
  - Update the Master Plan to reflect new issues

  If action is 'status':
  - Read the Master Plan and report current progress
  - List which batches are complete, in progress, and pending
```

### Issue Comment Trigger for Next Batch

Add to `claude-architect.yml` triggers:

```yaml
on:
  # ... existing triggers ...

  issue_comment:
    types: [created]

jobs:
  architect-analyze:
    if: |
      github.event_name == 'workflow_dispatch' ||
      (github.event_name == 'issue_comment' &&
       contains(github.event.issue.labels.*.name, 'architect-master-plan') &&
       contains(github.event.comment.body, '@claude-architect'))
```

### Batch Orchestration Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                        BATCH ORCHESTRATION                          │
└─────────────────────────────────────────────────────────────────────┘

1. INITIAL TRIGGER (Human uploads PRD)
   │
   ▼
┌─────────────────────────┐
│  claude-architect.yml   │
│  action: new            │
└───────────┬─────────────┘
            │
            ├──► Creates Master Plan Issue (#100)
            │    - Label: architect-master-plan
            │    - Contains all batches overview
            │
            └──► Creates Batch 1 Issues (#101-#107)
                 - Label: claude-implement, batch/1
                 - Foundation issues only

2. BUILDER IMPLEMENTS (Automatic via claude-implement.yml)
   │
   ▼
┌─────────────────────────┐
│  claude-implement.yml   │  For each issue in Batch 1
│  Triggered by label     │
└───────────┬─────────────┘
            │
            └──► Creates PRs, merges when approved

3. BATCH COMPLETION (Automatic detection)
   │
   ▼
┌─────────────────────────┐
│  claude-architect-      │
│  monitor.yml            │  Triggered on issue close
└───────────┬─────────────┘
            │
            ├──► Checks if all Batch 1 issues closed
            │
            └──► If complete:
                 - Comments on Master Plan: "Batch complete!"
                 - Updates Master Plan checkboxes
                 - Suggests: "@claude-architect next"

4. HUMAN APPROVAL (Required)
   │
   ▼
┌─────────────────────────┐
│  Human reviews Master   │
│  Plan and comments:     │
│  "@claude-architect     │
│   next"                 │
└───────────┬─────────────┘
            │
            ▼

5. NEXT BATCH (Triggered by comment)
   │
   ▼
┌─────────────────────────┐
│  claude-architect.yml   │
│  action: next           │
│  Reads Master Plan      │
└───────────┬─────────────┘
            │
            └──► Creates Batch 2 Issues (#108-#113)
                 - Updates Master Plan

6. REPEAT steps 2-5 until all batches complete
```

## Handling PRD Updates (Amendment Issues)

When a PRD is updated after issues have already been created, create **amendment issues** rather than modifying originals.

### Why Amendments?

- Original issue stays stable for in-progress work
- Clear audit trail of requirement changes
- Amendment can be scheduled for a later batch
- Builder working on original isn't surprised mid-implementation

### Amendment Issue Template

```markdown
# [Amendment] [Brief description of change]

**Amends**: #[original_issue_number]
**PRD Change Date**: [date]

## Change Summary

**Original scope**: [what the original issue specified]
**Change**: [what's being added/modified/removed]

## New/Modified Acceptance Criteria

- [ ] New criterion 1
- [ ] New criterion 2
- [ ] Modified criterion (was: "old text")

## Technical Impact

<!-- How does this change affect the original implementation? -->
- Database changes: [yes/no, details]
- API changes: [yes/no, details]
- UI changes: [yes/no, details]

## Dependencies

- Requires #[original] to be completed first
- Blocked by: #XX (if any)

## Scope

Size: [XS/S/M/L]

---
🤖 Generated by Claude Architect (PRD Amendment)
```

### Amendment Workflow

```
PRD Updated
     │
     ▼
┌─────────────────────────┐
│  claude-architect.yml   │
│  action: amend          │
│  original_prd: URL      │
│  updated_prd: URL       │
└───────────┬─────────────┘
            │
            ├──► Diff the two PRDs
            │
            ├──► Identify affected issues
            │
            └──► Create amendment issues
                 - Label: claude-implement, amendment
                 - References original issue
                 - Added to next available batch

Original Issue #5          Amendment Issue #25
┌─────────────────┐        ┌─────────────────────┐
│ Auth System     │◄───────│ [Amendment] Add     │
│ - Email/pass    │ Amends │ OAuth to Auth       │
│ - JWT tokens    │        │ - Google OAuth      │
│                 │        │ - Apple OAuth       │
│ Status: Done ✅ │        │ Status: Pending     │
└─────────────────┘        └─────────────────────┘
```

### Workflow Dispatch Input for Amendments

```yaml
inputs:
  # ... existing inputs ...
  batch_action:
    description: 'Batch action'
    type: choice
    options:
      - new      # Start new project from PRD
      - next     # Create next batch for existing project
      - amend    # Create amendments from PRD update
      - status   # Report current status
    default: 'new'
  original_prd_url:
    description: 'Original PRD URL (for amend action)'
    required: false
    type: string
  updated_prd_url:
    description: 'Updated PRD URL (for amend action)'
    required: false
    type: string
```

## Open Questions

1. ~~Should issues be created in batches (foundation first, then dependents) or all at once?~~ **Resolved: Hybrid batch approach with human control**
2. Should there be a "sprint" concept where only N issues are created at a time?
   - *Current approach*: Batches are logical groupings, not time-boxed sprints
3. ~~How to handle PRD updates - create new issues or update existing ones?~~ **Resolved: Amendment issues that reference originals**

---

**Status**: conceived
**Author**: Architect
**Date**: 2026-02-04
