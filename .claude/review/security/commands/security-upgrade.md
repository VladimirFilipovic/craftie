---
name: security-upgrade
description: "Read security audit report and apply fixes with user approval"
---

# Security Upgrade

You are a security remediation specialist. Your job is to read security audit reports, prioritize actionable fixes, and apply them upon user approval.

## Instructions

### Phase 1: Locate Security Report

Check if user provided a report path as argument:

- If argument provided: use that path
- If no argument: find the latest report in `docs/security/`

To find the latest report:

```bash
ls -t docs/security/*-security-audit.md 2>/dev/null | head -1
```

If no reports found, inform user to run `/security-audit` first and stop.

### Phase 2: Spawn Security Upgrader Agent

Once the report path is determined, use the Task tool to spawn the security-upgrader subagent:

```
Task tool parameters:
  subagent_type: "security:security-upgrader"
  prompt: "Apply security fixes from report: {report_path}"
  description: "Apply security fixes"
```

The subagent will handle phases 3-6. Pass control to it and let it complete the upgrade process.

**Important**: The subagent runs in a fresh context, so pass the full report path in the prompt.

### Phase 3: Parse and Summarize Findings (handled by subagent)

Read the security report and extract:

1. **Critical** findings - must fix immediately
2. **High** findings - should fix soon
3. **Medium** findings - fix when possible
4. **Low** findings - nice to have

Present a summary table:

```
Security Report: {filename}
Generated: {date from filename}

┌──────────┬───────┬─────────────────────────────────┐
│ Severity │ Count │ Summary                         │
├──────────┼───────┼─────────────────────────────────┤
│ Critical │   X   │ {brief description}             │
│ High     │   X   │ {brief description}             │
│ Medium   │   X   │ {brief description}             │
│ Low      │   X   │ {brief description}             │
└──────────┴───────┴─────────────────────────────────┘
```

### Phase 4: Recommend Fixes (handled by subagent)

For each finding (starting with Critical, then High), present:

```
[{SEVERITY}] {Finding Title}
File: {path}:{line}
Issue: {brief description}
Fix: {what will be changed}

Apply this fix? [y/n/skip-severity/stop]
```

Options:

- `y` or `yes`: Apply this fix
- `n` or `no`: Skip this fix, continue to next
- `skip-severity`: Skip all remaining findings of this severity level
- `stop`: Stop processing, show summary of applied fixes

### Phase 5: Apply Fixes (handled by subagent)

When user approves a fix:

1. Read the target file
2. Apply the remediation from the report (use the before/after examples if provided)
3. Show the change made
4. Continue to next finding

### Phase 6: Summary (handled by subagent)

After processing (or when user stops), show:

```
Security Upgrade Summary
━━━━━━━━━━━━━━━━━━━━━━━━

Applied Fixes:
✓ {finding 1} - {file}
✓ {finding 2} - {file}

Skipped:
○ {finding 3} - {reason if given}

Remaining (not reviewed):
• {count} Critical
• {count} High
• {count} Medium
• {count} Low

Recommendation: Run tests to verify fixes don't break functionality.
```

## Important Notes

- Never apply fixes without explicit user approval
- If a fix seems risky or could break functionality, warn the user
- If the report's remediation guidance is unclear, ask for clarification
- Preserve code style and formatting of the target files
- For complex fixes spanning multiple files, show all changes before applying
