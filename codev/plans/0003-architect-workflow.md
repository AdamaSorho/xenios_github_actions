# Plan 0003: AI Architect Workflow - Implementation Plan

## Overview

Implement the Claude Architect GitHub Action workflow that analyzes PRD documents and creates well-structured issues in batches for the Builder workflow to implement.

**Spec**: `codev/specs/0003-architect-workflow.md`

## Implementation Phases

### Phase 1: Core Architect Workflow

**Goal**: Create the main `claude-architect.yml` workflow that can analyze PRDs and create issues.

**Files to create**:
- `.github/workflows/claude-architect.yml`

**Tasks**:
1. Create workflow file with `workflow_dispatch` trigger
2. Add inputs: `prd_url`, `prd_content`, `mode`, `priority`, `batch_action`, `master_plan_issue`
3. Add `issue_comment` trigger for `@claude-architect` commands
4. Implement PRD fetch step (curl for URLs)
5. Configure Claude Code action with architect prompt
6. Add issue template in the prompt
7. Add architecture rules from CLAUDE.md in the prompt
8. Implement mode handling (preview/draft/create)

**Acceptance Criteria**:
- [ ] Workflow can be triggered from GitHub UI
- [ ] PRD content can be provided via URL or direct paste
- [ ] Preview mode outputs markdown summary without creating issues
- [ ] Create mode creates issues with proper labels

### Phase 2: Master Plan Issue & Batch Creation

**Goal**: Implement the Master Plan issue pattern for tracking batches.

**Files to modify**:
- `.github/workflows/claude-architect.yml`

**Tasks**:
1. Add Master Plan issue template to prompt
2. Implement `action: new` - creates Master Plan + Batch 1
3. Implement `action: next` - reads Master Plan, creates next batch
4. Implement `action: status` - reports current progress
5. Add `architect-master-plan` label creation
6. Add batch labels (`batch/1`, `batch/2`, etc.)

**Acceptance Criteria**:
- [ ] Master Plan issue created with full batch overview
- [ ] Only Batch 1 issues created on initial run
- [ ] `next` action correctly identifies and creates next batch
- [ ] `status` action reports accurate progress

### Phase 3: Batch Monitor Workflow

**Goal**: Create the monitor workflow that detects batch completion.

**Files to create**:
- `.github/workflows/claude-architect-monitor.yml`

**Tasks**:
1. Create workflow with `issues: [closed]` trigger
2. Add condition: only for `claude-implement` labeled issues
3. Implement Master Plan issue finder
4. Implement batch completion checker
5. Add notification comment on batch complete
6. Implement Master Plan status updater (checkboxes)

**Acceptance Criteria**:
- [ ] Workflow triggers only for architect-created issues
- [ ] Correctly detects when all issues in a batch are closed
- [ ] Posts completion notification on Master Plan
- [ ] Updates Master Plan checkboxes automatically

### Phase 4: Amendment Support

**Goal**: Implement PRD amendment handling.

**Files to modify**:
- `.github/workflows/claude-architect.yml`

**Tasks**:
1. Add `amend` action handling
2. Add `original_prd_url` and `updated_prd_url` inputs
3. Implement PRD diff logic in prompt
4. Add amendment issue template
5. Add `amendment` label for amendment issues

**Acceptance Criteria**:
- [ ] Can provide original and updated PRD URLs
- [ ] Amendment issues reference original issues
- [ ] Amendment issues have proper labeling
- [ ] Amendments added to next available batch

### Phase 5: Post-Run Summary

**Goal**: Add summary comment after workflow completion.

**Files to modify**:
- `.github/workflows/claude-architect.yml`

**Tasks**:
1. Add `actions/github-script` step for summary
2. Post summary comment on Master Plan issue
3. Include link to workflow run
4. List created issues in summary

**Acceptance Criteria**:
- [ ] Summary posted after successful run
- [ ] Includes count of issues created
- [ ] Links to workflow logs for details

## File Structure

```
.github/
└── workflows/
    ├── claude-architect.yml        # Main architect workflow (Phase 1, 2, 4, 5)
    └── claude-architect-monitor.yml # Batch completion monitor (Phase 3)
```

## Dependencies

- **Spec 0002** (AI-Powered CI/CD Platform) must be committed first
  - Provides `claude-implement.yml` that will pick up created issues
  - Provides label and workflow patterns to follow

## Labels to Create

The workflows will create these labels if they don't exist:

| Label | Description | Color |
|-------|-------------|-------|
| `claude-implement` | Issues for Claude Builder to implement | `#7057ff` |
| `architect-master-plan` | Master Plan tracking issues | `#0e8a16` |
| `amendment` | PRD amendment issues | `#fbca04` |
| `batch/1`, `batch/2`, etc. | Batch grouping | `#c5def5` |
| `size/XS`, `size/S`, `size/M`, `size/L` | Scope estimates | `#bfdadc` |
| `priority/high`, `priority/medium`, `priority/low` | Priority levels | `#d93f0b`, `#fbca04`, `#0e8a16` |

## Testing Plan

### Manual Testing

1. **Preview Mode Test**
   - Trigger workflow with sample PRD content
   - Verify markdown output in logs
   - Confirm no issues created

2. **Create Mode Test**
   - Trigger with `mode: create`, small PRD
   - Verify Master Plan issue created
   - Verify Batch 1 issues created with correct labels
   - Verify dependency references in issue bodies

3. **Batch Completion Test**
   - Close all Batch 1 issues
   - Verify monitor workflow triggers
   - Verify notification comment posted
   - Verify Master Plan checkboxes updated

4. **Next Batch Test**
   - Comment `@claude-architect next` on Master Plan
   - Verify Batch 2 issues created
   - Verify Master Plan updated

5. **Amendment Test**
   - Trigger with `action: amend` and two PRD URLs
   - Verify amendment issues created
   - Verify references to original issues

### Integration Testing

1. **Full Pipeline Test**
   - Create issues via Architect
   - Verify Builder workflow picks them up
   - Verify PRs created and merged
   - Verify batch completion detected

## Security Considerations

1. **PRD URL Validation** - Only fetch from trusted domains (github.com, githubusercontent.com)
2. **Issue Limit** - Cap at 50 issues per run to prevent runaway creation
3. **Rate Limiting** - Add delays between issue creation calls
4. **Content Sanitization** - Escape any code blocks in PRD content

## Rollout Plan

1. Deploy to repository with `mode: preview` as default
2. Test with small PRD in preview mode
3. Test with small PRD in create mode (creates real issues)
4. Document usage in README or wiki
5. Announce availability to team

## Estimated Scope

| Phase | Scope | Notes |
|-------|-------|-------|
| Phase 1 | M (1-2 days) | Core workflow setup |
| Phase 2 | M (1-2 days) | Master Plan + batching logic |
| Phase 3 | S (4-8 hrs) | Monitor workflow |
| Phase 4 | S (4-8 hrs) | Amendment support |
| Phase 5 | XS (<4 hrs) | Summary posting |
| **Total** | **L (3-5 days)** | |

---

**Status**: planned
**Author**: Architect
**Date**: 2026-02-04
