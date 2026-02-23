---
description: "Orchestrate a full review across code quality, security, and performance domains. Spawns specialized agents in parallel, consolidates findings into a unified report, and guides you through applying fixes."
allowed-tools: Task, Bash(git status:*), Bash(git diff:*), Bash(git log:*), Bash(date:*), Bash(mkdir:*), AskUserQuestion, Write
argument-hint: '[scope] - e.g., "whole project", "current branch", or leave blank to be asked'
---

# Full Review

## Step 1: Repository Snapshot

!`git status --short && echo "---" && git diff --stat && echo "---" && git log --oneline -5`

## Step 2: Select Domains

Use AskUserQuestion to ask which review domains to run (multiSelect: true):

- **Code Quality** — architecture, patterns, complexity, testing, documentation
- **Security** — vulnerabilities, OWASP compliance, dependency scanning
- **Performance** — bottlenecks, memory usage, profiling opportunities

If no argument was provided, show this selection. If `$ARGUMENTS` contains a clear scope like "all", run all three domains without asking.

## Step 3: Select Scope

Use AskUserQuestion to ask the review scope (single-select):

- **Whole project** — full codebase analysis
- **Current branch** — changes since branch diverged from main (`git diff main...HEAD`)
- **Pending changes** — staged and unstaged changes only (`git diff HEAD`)

Skip this question if the user already specified scope in `$ARGUMENTS`.

## Step 4: Generate Timestamp & Prepare Directories

```bash
date '+%Y-%m-%d-%H%M%S'
```

```bash
mkdir -p docs/reviews docs/security docs/performance
```

Store the timestamp — pass it to every agent so all reports share the same timestamp prefix.

## Step 5: Launch Agents in Parallel

Use the Task tool to spawn all selected domain agents **simultaneously** in a single response (multiple Task calls in one message). Each agent receives the scope and the shared timestamp.

Pass this context to every agent:
- The review scope selected in Step 3
- The shared timestamp (e.g., `2026-02-21-143022`)
- Instruction to return a summary section in their response (executive summary + all Critical and High findings) for consolidation, in addition to saving their full report to disk

### Code Quality Agent (if selected)

```
subagent_type: code-quality:code-review-expert
description: Full code quality review

Prompt:
Perform a comprehensive code quality review.

Scope: [scope from Step 3]

Review all six aspects: Architecture & Design, Code Quality, Security & Dependencies,
Performance & Scalability, Testing Quality, Documentation & API.

At the end of your response, include a section titled "## SUMMARY FOR CONSOLIDATION"
containing:
- A score out of 10 for each aspect reviewed
- A count of Critical, High, Medium, Low issues found
- All Critical and High issue titles with file:line references

Do NOT save a file — return your full report as your response.
```

### Security Agent (if selected)

```
subagent_type: security:security-auditor
description: Full security audit

Prompt:
Perform a comprehensive security audit following the security-auditing skill methodology.

Scope: [scope from Step 3]

Save the full report to: docs/security/[timestamp]-security-audit.md

At the end of your response, include a section titled "## SUMMARY FOR CONSOLIDATION"
containing:
- Overall security score out of 100
- Risk Assessment table (Critical/High/Medium/Low counts)
- All Critical and High finding titles (C-001, H-001 format) with file:line references
- OWASP compliance score (X/10)
```

### Performance Agent (if selected)

```
subagent_type: performance:performance-auditor
description: Full performance audit

Prompt:
Perform a comprehensive performance audit following the performance-auditing skill methodology.

Scope: [scope from Step 3]

Save the full report to: docs/performance/[timestamp]-performance-audit.md

At the end of your response, include a section titled "## SUMMARY FOR CONSOLIDATION"
containing:
- Overall performance score out of 100
- Issue count table (Critical/High/Medium/Low)
- All Critical and High finding titles (P-001, P-002 format) with file:line references
- Top 3 bottlenecks with estimated impact
```

Wait for all agents to complete before proceeding.

## Step 6: Consolidate Report

After all agents return, assemble and write the consolidated report to `docs/reviews/[timestamp]-full-review.md`.

Use the Write tool with this structure:

```markdown
# Full Review Report

**Date**: [timestamp formatted as YYYY-MM-DD HH:MM:SS]
**Scope**: [scope description]
**Domains**: [list of reviewed domains]

---

## Executive Summary

| Domain         | Score  | Critical | High | Medium | Low |
|----------------|--------|----------|------|--------|-----|
| Code Quality   | X/10   | X        | X    | X      | X   |
| Security       | X/100  | X        | X    | X      | X   |
| Performance    | X/100  | X        | X    | X      | X   |

[2-3 sentence narrative synthesizing the overall state across all domains]

---

## Priority Action Plan

### Critical — Fix Immediately

[Merged list from all domains, prefixed with domain tag]
- [CODE] Issue title — `file:line`
- [SEC] C-001: Issue title — `file:line`
- [PERF] P-001: Issue title — `file:line`

### High — Fix Before Next Release

[Same merged format]

### Medium — Fix Soon

[Same merged format]

---

## Domain Summaries

### Code Quality
[Full code-review-expert report content]

---

### Security
[Executive summary + Critical/High findings from security-auditor]
Full report: `docs/security/[timestamp]-security-audit.md`

---

### Performance
[Executive summary + Critical/High findings from performance-auditor]
Full report: `docs/performance/[timestamp]-performance-audit.md`

---

## Cross-Domain Patterns

[Identify files or components that appear across multiple domain reports — these are the highest-priority refactoring targets. E.g., "src/session/session.go appears in both security (missing validation) and performance (blocking I/O) findings — coordinate fixes."]

---

## Recommended Fix Sequence

1. Fix Critical security issues first (run /security-upgrade)
2. Address Critical performance bottlenecks
3. Resolve Critical and High code quality issues
4. Tackle remaining High issues by domain
5. Schedule Medium issues in next sprint

---

## Report Links

- Full review: `docs/reviews/[timestamp]-full-review.md`
[- Security report: `docs/security/[timestamp]-security-audit.md`]
[- Performance report: `docs/performance/[timestamp]-performance-audit.md`]
```

After writing, print the report path and a brief summary of findings.

## Step 7: Fix Workflow

Use AskUserQuestion to ask what to do next (single-select):

- **Fix security issues** — launch security-upgrader with the security report path
- **Walk through code quality fixes** — guide through Critical/High code issues inline
- **Walk through performance fixes** — guide through Critical/High performance issues inline
- **Save context and fix later** — show commands to resume later, then stop
- **Done for now** — stop

If user selects **Fix security issues**:
- Spawn the security-upgrader agent with prompt: `Fix findings from docs/security/[timestamp]-security-audit.md. Start with Critical findings.`

If user selects **Walk through code quality fixes** or **Walk through performance fixes**:
- Work through the Critical and High findings from the respective domain report inline, one at a time, applying fixes with user approval using the same approve/skip/stop pattern from security-upgrade.

If user selects **Save context and fix later**:
```
To resume after clearing context:
  Security fixes:    /security-upgrade docs/security/[timestamp]-security-audit.md
  Full report:       docs/reviews/[timestamp]-full-review.md
```
