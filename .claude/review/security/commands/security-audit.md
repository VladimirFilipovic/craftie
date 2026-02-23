---
name: security-audit
description: "Comprehensive security audit to identify vulnerabilities, OWASP Top 10 issues, and security anti-patterns."
---

# Security Audit

## Orchestrator Integration

When called from `/full-review` via the Task tool, use the scope and report path provided in the prompt — skip the interactive flow below and go directly to the Security Analysis step.

## Standalone Flow

### Pre-Audit Check: Security Configuration

Read `.claude/settings.json` using the Read tool. If it has fewer than 4 deny rules:

```
⚠️  Security Configuration Warning

Your .claude/settings.json has fewer than 4 file denial rules.
Run /review-init first to configure proper security settings.

Would you like to:
1. Continue with audit anyway
2. Run /review-init first (recommended — restart Claude Code after)
```

### Scope Selection

Ask the user to choose the audit scope:

1. **Whole project** — full codebase analysis
2. **Current branch** — changes since diverging from main
3. **Pending changes** — staged and unstaged changes only

Pass the selected scope to the security-auditor agent.

## Security Analysis

Use the Task tool with subagent_type `security:security-auditor` to perform a thorough security analysis.

### Analysis Scope

1. **Code Pattern Analysis**: Scan for injection vulnerabilities, authentication bypasses, insecure configurations
2. **Architecture Review**: Analyze authentication, authorization, session management, and data protection patterns
3. **Dependency Security**: Review packages for known vulnerabilities and outdated versions
4. **OWASP Compliance**: Assess against OWASP Top 10 2021
5. **Configuration Security**: Check for hardcoded secrets, missing security headers, and misconfigurations

### Output Requirements

- Save the report to: `docs/security/YYYY-MM-DD-HHMMSS-security-audit.md`
- Include actual findings with exact file paths and line numbers
- Provide before/after code examples for remediation
- Prioritize findings: Critical, High, Medium, Low

### Important Notes

- Focus on **defensive security** — identifying vulnerabilities to help write secure code
- Provide actionable remediation with specific code examples
- Create a prioritized remediation roadmap based on risk severity
