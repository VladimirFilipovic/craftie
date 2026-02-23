---
name: review-init
description: "Initialize the review system: detect project technologies, verify tools across all review domains, configure security file denial patterns, and inject technology-specific tooling into code-quality, security, and performance workflows."
allowed-tools: Glob, Bash(which:*), Bash(mkdir:*), Read, Write, AskUserQuestion
---

# Review Init

Initialize the full review system for this project. Detects technologies, checks tool availability, configures security settings, and injects tech-specific tooling into all three review domains.

> **Note**: This command supersedes `/security-init`. It does everything security-init does, plus configures code-quality and performance tooling.

**BEFORE DOING ANYTHING ELSE**: Begin with Phase 1 technology detection. Do NOT skip any phases.

---

## Phase 1: Technology Detection

Scan the project root using **Glob tool only** (no bash commands):

- **Go**: `go.mod`, `go.sum`
- **Node.js**: `package.json`, `yarn.lock`, `pnpm-lock.yaml`, `bun.lockb`
- **Python**: `requirements.txt`, `pyproject.toml`, `setup.py`, `Pipfile`, `poetry.lock`
- **Rust**: `Cargo.toml`, `Cargo.lock`
- **Java**: `pom.xml`, `build.gradle`, `build.gradle.kts`
- **Ruby**: `Gemfile`, `Gemfile.lock`
- **PHP**: `composer.json`, `composer.lock`
- **.NET**: `*.csproj`, `*.sln`, `global.json`
- **Docker**: `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`

Only check file existence — do NOT read file contents.

---

## Phase 2: Verify Tool Availability

For each detected technology, check tools across all three review domains using `which`:

### Go
```bash
which gofmt >/dev/null 2>&1        # Code Quality
which staticcheck >/dev/null 2>&1  # Code Quality
which govulncheck >/dev/null 2>&1  # Security
which gosec >/dev/null 2>&1        # Security
# pprof is built into Go toolchain — always available
```

### Python
```bash
which ruff >/dev/null 2>&1         # Code Quality
which mypy >/dev/null 2>&1         # Code Quality
which bandit >/dev/null 2>&1       # Security
which pip-audit >/dev/null 2>&1    # Security
which py-spy >/dev/null 2>&1       # Performance
```

### Node.js
```bash
which eslint >/dev/null 2>&1       # Code Quality
which tsc >/dev/null 2>&1          # Code Quality
which npm >/dev/null 2>&1          # Security
which clinic >/dev/null 2>&1       # Performance
```

### Rust
```bash
which cargo >/dev/null 2>&1        # Code Quality (clippy) + Security (audit)
```

### Java
```bash
which mvn >/dev/null 2>&1 || which gradle >/dev/null 2>&1  # All domains
```

### Ruby
```bash
which rubocop >/dev/null 2>&1      # Code Quality
which bundle >/dev/null 2>&1       # Security
which brakeman >/dev/null 2>&1     # Security
```

### PHP
```bash
which phpcs >/dev/null 2>&1        # Code Quality
which composer >/dev/null 2>&1     # Security
```

### .NET
```bash
which dotnet >/dev/null 2>&1       # All domains
```

Display results grouped by domain:

```
Tool Availability:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code Quality
  ✓ gofmt
  ✗ staticcheck  →  go install honnef.co/go/tools/cmd/staticcheck@latest

Security
  ✓ govulncheck
  ✗ gosec        →  go install github.com/securego/gosec/v2/cmd/gosec@latest

Performance
  ✓ pprof (built-in)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

If any tools are missing, ask:

```
Some tools are not installed. How would you like to proceed?
1. Continue anyway (missing tools will be skipped during reviews)
2. Stop and install missing tools first (recommended)
```

If user chooses to stop, show all install commands and exit.

---

## Phase 3: Build Security Denial Patterns

Compile deny patterns for `.claude/settings.json` to prevent reading sensitive files.

### Base Patterns (always included)

**Environment files**: `.env`, `**/.env`, `.env.*`, `**/.env.*`, `.env.local`, `.env.development`, `.env.production`, `.env.test`

**Credentials & Secrets**: `credentials.json`, `**/credentials.json`, `secrets.yml`, `**/secrets.yml`, `.secret`, `**/.secret`, `*.secret`

**SSH & Certificates**: `id_rsa`, `id_rsa.pub`, `id_ed25519`, `id_ed25519.pub`, `*.pem`, `*.key`, `*.p12`, `*.jks`, `*.pfx`, `*.keystore`, `*.cer`, `*.crt`

**Cloud Provider Credentials**: `.aws/credentials`, `.aws/config`, `.gcp/credentials.json`, `.azure/credentials`

**Database Files**: `*.db`, `*.sqlite`, `*.sqlite3`

**Version Control & IDE**: `.git/**`, `.vscode/**`, `.idea/**`, `.devcontainer/**`, `.github/workflows/**`

**Package Management**: `node_modules/**`, `package-lock.json`

### Technology-Specific Patterns

**Go**: `vendor/**`
**Python**: `.venv/**`, `venv/**`, `__pycache__/**`, `**/__pycache__/**`, `*.pyc`, `.pytest_cache/**`, `.mypy_cache/**`, `.ruff_cache/**`
**Node.js**: `node_modules/**`, `.next/**`, `.nuxt/**`, `dist/**`, `build/**`, `.cache/**`, `.turbo/**`
**Rust**: `target/**`
**Java**: `target/**`, `*.class`, `.gradle/**`, `build/**`
**Ruby**: `vendor/bundle/**`, `.bundle/**`
**PHP**: `vendor/**`
**.NET**: `bin/**`, `obj/**`, `*.user`, `*.suo`, `.vs/**`, `TestResults/**`
**Docker**: `docker-compose.override.yml`, `docker-compose.override.yaml`

---

## Phase 4: Check Existing Configuration

Use the **Read tool** to check if `.claude/settings.json` exists:

1. If it exists: parse JSON, check for existing `permissions.deny` section
2. Ask merge strategy using AskUserQuestion:
   - **Deduplicate** (default): keep existing, add only new patterns
   - **Append**: add all patterns, keep duplicates
   - **Replace**: replace entire deny section
3. If it doesn't exist: proceed to create it

---

## Phase 5: Preview & Confirm

Display:

```
🔍 Technologies Detected: Go, Docker

Security Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Base patterns:        25
Go patterns:           1 (vendor/**)
Docker patterns:       2
New patterns total:   28
After merge:          28

Review Workflows to Update:
  → code-review-expert.md    (Go static analysis tools)
  → security-audit.md        (govulncheck, gosec)
  → security-upgrading SKILL (Go remediation patterns)
  → performance-audit.md     (pprof, benchmarks)
  → performance-auditing SKILL (Go profiling guidance)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Proceed? (yes/no)
```

---

## Phase 6: Write settings.json

After confirmation:

1. `mkdir -p .claude` (Bash tool)
2. Write updated `.claude/settings.json` (Write tool) — preserve all non-permission settings, format with 2-space indentation

---

## Phase 7: Inject Code-Quality Tooling

Read `.claude/review/code-quality/agents/code-review-expert.md`. Find the `<!-- TECH-TOOLS-START -->` marker and replace the content between `<!-- TECH-TOOLS-START -->` and `<!-- TECH-TOOLS-END -->` with technology-specific static analysis tools. Only append entries for detected technologies (don't duplicate if already present).

**Go:**
```markdown
#### Go Static Analysis
- Run `gofmt -l .` to identify files with formatting issues
- Run `go vet ./...` to detect suspicious constructs and common mistakes
- Run `staticcheck ./...` for advanced static analysis (if installed)
- Check for excessive function length, deep nesting, and ignored errors
```

**Python:**
```markdown
#### Python Static Analysis
- Run `ruff check .` for fast linting and style enforcement (if installed)
- Run `mypy .` for type checking (if installed)
- Check for mutable default arguments, bare excepts, and missing type hints
```

**Node.js:**
```markdown
#### Node.js Static Analysis
- Run `eslint .` for linting (if installed)
- Run `tsc --noEmit` for type checking in TypeScript projects (if installed)
- Check for callback hell, missing error handling in promises, and prototype pollution
```

**Rust:**
```markdown
#### Rust Static Analysis
- Run `cargo clippy -- -D warnings` for linting
- Run `rustfmt --check src/**/*.rs` to verify formatting
- Check for unnecessary clones, unused Results, and missing error propagation
```

**Java:**
```markdown
#### Java Static Analysis
- Run `mvn checkstyle:check` or `gradle checkstyleMain` (if configured)
- Check for resource leaks (unclosed streams), null pointer risks, and raw types
```

**Ruby:**
```markdown
#### Ruby Static Analysis
- Run `rubocop` for style and quality checks (if installed)
- Check for N+1 queries, missing validations, and mass assignment
```

**PHP:**
```markdown
#### PHP Static Analysis
- Run `phpcs` for coding standard checks (if installed)
- Check for SQL concatenation, unescaped output, and global variable usage
```

**.NET:**
```markdown
#### .NET Static Analysis
- Run `dotnet format --verify-no-changes` to check formatting
- Use Roslyn analyzers (built into build) for common issues
- Check for async-over-sync, improper disposal, and missing null checks
```

---

## Phase 8: Inject Security Domain Tooling

### Update security-audit.md

Read `.claude/review/security/commands/security-audit.md`. Find the "Analysis Scope" section and append technology-specific scanning entries. Only add entries for detected technologies.

**Go:**
```markdown
6. **Go Security**: Run `govulncheck ./...` for known dependency vulnerabilities; run `gosec ./...` for static security analysis; check for unsafe package usage and swallowed errors
```

**Python:**
```markdown
6. **Python Security**: Run `bandit -r .` for static security analysis; run `pip-audit` for dependency vulnerabilities; check for pickle usage, eval/exec calls, SQL string formatting
```

**Node.js:**
```markdown
6. **Node.js Security**: Run `npm audit` for dependency vulnerabilities; check for prototype pollution patterns, unsafe RegEx (ReDoS), and eval/Function constructor usage
```

**Rust:**
```markdown
6. **Rust Security**: Run `cargo audit` for dependency vulnerabilities; review unsafe blocks and document justification; check for panics in library code
```

**Java:**
```markdown
6. **Java Security**: Run OWASP Dependency-Check or `mvn dependency:tree`; check for deserialization vulnerabilities, SQL injection in JDBC calls, and hardcoded secrets
```

**Ruby:**
```markdown
6. **Ruby Security**: Run `bundle audit` for dependency vulnerabilities; run `brakeman` for Rails security analysis; check for mass assignment vulnerabilities
```

**PHP:**
```markdown
6. **PHP Security**: Run `composer audit` for dependency vulnerabilities; check for eval/exec/system usage and SQL injection in raw queries
```

**.NET:**
```markdown
6. **.NET Security**: Run `dotnet list package --vulnerable`; check for SQL injection in raw EF queries, XXE in XML parsing, and missing anti-forgery tokens
```

### Update security-upgrading SKILL.md

Read `.claude/review/security/skills/security-upgrading/SKILL.md`. Find "Important Notes" section and append technology-specific fix patterns for detected technologies.

**Go:**
```markdown
### Go-Specific Fixes
- **Dependency vulnerabilities**: `go get -u <module>` or pin specific version in go.mod
- **Unsafe package**: replace with safe alternatives or add documented justification
- **Error swallowing**: ensure all errors are handled or explicitly ignored with `_ =`
- **Hardcoded secrets**: move to `os.Getenv()` or a config package
```

**Python:**
```markdown
### Python-Specific Fixes
- **Dependency vulnerabilities**: update in requirements.txt/pyproject.toml, run `pip install --upgrade`
- **Pickle usage**: replace with `json` for untrusted data
- **eval/exec**: replace with `ast.literal_eval()` or safer alternatives
- **SQL formatting**: use parameterized queries with `?` or `%s` placeholders
```

**Node.js:**
```markdown
### Node.js-Specific Fixes
- **Dependency vulnerabilities**: run `npm audit fix` or update package.json manually
- **Prototype pollution**: use `Object.create(null)` or Map for user-controlled keys
- **eval usage**: replace with `JSON.parse()` or structured alternatives
```

**Rust:**
```markdown
### Rust-Specific Fixes
- **Dependency vulnerabilities**: update Cargo.toml, run `cargo update`
- **Unsafe blocks**: add `// SAFETY:` comment documenting invariants
- **Panic in library**: replace `unwrap()`/`expect()` with `?` or `Result` returns
```

**Java:**
```markdown
### Java-Specific Fixes
- **SQL injection**: use PreparedStatement with `?` parameters
- **Deserialization**: use allowlists, avoid `ObjectInputStream` on untrusted data
- **Hardcoded secrets**: move to environment variables or secrets manager
```

**Ruby:**
```markdown
### Ruby-Specific Fixes
- **Mass assignment**: use `strong_parameters`, define `permit()` explicitly
- **SQL injection**: use ActiveRecord query interface, avoid string interpolation
```

**PHP:**
```markdown
### PHP-Specific Fixes
- **SQL injection**: use PDO prepared statements
- **Command injection**: use `escapeshellarg()`, avoid exec/system with user input
- **XSS**: use `htmlspecialchars()` or templating engine auto-escaping
```

**.NET:**
```markdown
### .NET-Specific Fixes
- **SQL injection**: use Entity Framework LINQ or parameterized `SqlCommand`
- **XXE**: set `DtdProcessing = DtdProcessing.Prohibit` on XmlReader
- **Missing CSRF**: add `[ValidateAntiForgeryToken]` to POST actions
```

---

## Phase 9: Inject Performance Domain Tooling

### Update performance-audit.md

Read `.claude/review/performance/commands/performance-audit.md`. Find the "Analysis Scope" section and append technology-specific profiling entries.

**Go:**
```markdown
6. **Go Performance**: Run `go test -bench=. -benchmem ./...` for micro-benchmarks; use `go tool pprof` to analyze CPU/memory profiles from `go test -cpuprofile`; run `go build -gcflags="-m" ./...` to review escape analysis; check for goroutine leaks and unbuffered channels in hot paths
```

**Python:**
```markdown
6. **Python Performance**: Use `cProfile` + `pstats` for function-level profiling; run `py-spy top` for live sampling (if installed); review generator vs list usage for large data; check for GIL contention in threaded code
```

**Node.js:**
```markdown
6. **Node.js Performance**: Use `clinic flame` or `0x` for flame graphs (if installed); analyze with `node --prof` and `node --prof-process`; check for synchronous fs operations, blocking event loop, and missing stream usage
```

**Rust:**
```markdown
6. **Rust Performance**: Run `cargo bench` for micro-benchmarks; use `cargo flamegraph` for profiling (if installed); check for excessive cloning, unnecessary heap allocations, and missing `#[inline]` on hot paths
```

**Java:**
```markdown
6. **Java Performance**: Use JProfiler or async-profiler for profiling; check for excessive object creation, missing connection pooling, and synchronization bottlenecks; review GC pressure from large object allocations
```

**Ruby:**
```markdown
6. **Ruby Performance**: Use `rack-mini-profiler` for web request profiling; check for N+1 queries with bullet gem patterns; review memory allocation with derailed_benchmarks patterns
```

**.NET:**
```markdown
6. **.NET Performance**: Use `BenchmarkDotNet` for micro-benchmarks; profile with dotTrace or PerfView; check for boxing allocations, async-over-sync, and missing `ConfigureAwait(false)`
```

### Update performance-auditing SKILL.md

Read `.claude/review/performance/skills/performance-auditing/SKILL.md`. Find `<!-- TECH-TOOLS-START -->` marker and replace content between it and `<!-- TECH-TOOLS-END -->` with technology-specific profiling guidance for detected technologies. Follow same pattern as Phase 7.

---

## Phase 10: Summary

```
✓ Review system initialized!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Files Updated:
  .claude/settings.json                              (XX deny patterns)
  .claude/review/code-quality/agents/code-review-expert.md
  .claude/review/security/commands/security-audit.md
  .claude/review/security/skills/security-upgrading/SKILL.md
  .claude/review/performance/commands/performance-audit.md
  .claude/review/performance/skills/performance-auditing/SKILL.md

Technologies Covered: [list]
Tools Configured: [list of available tools]

⚠️  Restart Claude Code for settings.json changes to take effect.

Next:
  /full-review    — comprehensive review across all domains
  /security-audit — standalone security analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```
