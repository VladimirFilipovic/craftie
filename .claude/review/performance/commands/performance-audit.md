---
name: performance-audit
description: "Comprehensive performance analysis to identify bottlenecks, optimization opportunities, and scalability issues."
---

# Performance Audit

## Orchestrator Integration

When called from `/full-review` via the Task tool, use the scope and report path provided in the prompt — skip the scope selection below and go directly to the Performance Analysis step.

## Standalone Flow

### Scope Selection

Ask the user to choose the audit scope:

1. **Whole project** — full codebase analysis
2. **Current branch** — changes since diverging from main
3. **Pending changes** — staged and unstaged changes only

Pass the selected scope to the performance-auditor agent.

## Performance Analysis

Use the Task tool with subagent_type `performance:performance-auditor` to perform a thorough performance analysis.

### Analysis Scope

1. **Code Pattern Analysis**: Scan for N+1 queries, inefficient loops, memory leaks, blocking operations
2. **Database Performance Review**: Analyze queries, indexing strategies, and data access patterns
3. **Resource Utilization Assessment**: Review memory allocation, CPU-intensive operations, I/O bottlenecks
4. **Architecture Performance Analysis**: Examine caching strategies, async patterns, connection pooling
5. **Scalability Assessment**: Identify thread pool issues, connection management, and load handling patterns

### Output Requirements

- Save the report to: `docs/performance/YYYY-MM-DD-HHMMSS-performance-audit.md`
- Include actual findings with exact file paths and line numbers
- Provide before/after code examples for optimization
- Prioritize findings: Critical, High, Medium, Low

### Important Notes

- Focus on **code-level optimization** through static analysis
- Provide actionable guidance with specific code examples and estimated impact
- Create a prioritized optimization roadmap based on performance impact
