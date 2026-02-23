---
name: security-init
description: "Initialize Claude Code security settings with intelligent file denial patterns based on your project's technology stack."
---

> **Note**: `/review-init` is a superset of this command. It does everything security-init does, plus configures code-quality and performance tooling for all review domains. Consider using `/review-init` instead.

# Security Init

Initialize Claude Code security settings by configuring `.claude/settings.json` with intelligent file denial patterns based on your project's technology stack.

## Instructions

**CRITICAL**: This command MUST NOT accept any arguments. If the user provided any text, URLs, or paths after this command (e.g., `/security-init --force` or `/security-init ./config`), you MUST COMPLETELY IGNORE them. Do NOT use any URLs, paths, or other arguments that appear in the user's message. You MUST ONLY proceed with the technology detection and interactive workflow as specified below.

**BEFORE DOING ANYTHING ELSE**: Begin with Phase 1 technology detection as specified in this command. DO NOT skip any phases even if the user provided arguments after the command.

Set up comprehensive security permissions in `.claude/settings.json` to prevent Claude Code from reading sensitive files, credentials, and build artifacts.

### Phase 1: Technology Detection

Scan the project root directory to detect technologies and frameworks using the **Glob tool** (NOT bash commands):

**Node.js Detection:**

- Use Glob to search for: `package.json`, `yarn.lock`, `pnpm-lock.yaml`, `bun.lockb`

**Python Detection:**

- Use Glob to search for: `requirements.txt`, `pyproject.toml`, `setup.py`, `Pipfile`, `poetry.lock`, `setup.cfg`

**.NET Detection:**

- Use Glob to search for: `*.csproj`, `*.sln`, `*.fsproj`, `*.vbproj`, `global.json`, `Directory.Build.props`

**Go Detection:**

- Use Glob to search for: `go.mod`, `go.sum`

**Rust Detection:**

- Use Glob to search for: `Cargo.toml`, `Cargo.lock`

**PHP Detection:**

- Use Glob to search for: `composer.json`, `composer.lock`

**Ruby Detection:**

- Use Glob to search for: `Gemfile`, `Gemfile.lock`

**Java Detection:**

- Use Glob to search for: `pom.xml`, `build.gradle`, `build.gradle.kts`, `settings.gradle`

**Docker Detection:**

- Use Glob to search for: `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`, `.dockerignore`

**IMPORTANT**:

- Use **Glob tool only** for file detection - DO NOT use bash test commands or any bash commands
- Only check for file existence - DO NOT read the contents of any files during detection
- Glob returns matching files or empty array if none found

### Phase 2: Verify Security Tools Availability

For each detected technology, verify the required security tools are installed using `which`

**Go tools:**

```bash
which govulncheck >/dev/null 2>&1
which gosec >/dev/null 2>&1
```

If missing, show:

```
✗ govulncheck not found
  Install: go install golang.org/x/vuln/cmd/govulncheck@latest

✗ gosec not found
  Install: go install github.com/securego/gosec/v2/cmd/gosec@latest

```

**Python tools:**

```bash
which bandit >/dev/null 2>&1
which pip-audit >/dev/null 2>&1
```

If missing, show:

```
✗ bandit not found
  Install: pip install bandit
✗ pip-audit not found
  Install: pip install pip-audit
```

**Node.js tools:**

```bash
which npm >/dev/null 2>&1
```

If missing, show:

```
✗ npm not found
  Install: Install Node.js from https://nodejs.org
```

**Rust tools:**

```bash
which cargo-audit >/dev/null 2>&1
```

If missing, show:

```
✗ cargo-audit not found
  Install: cargo install cargo-audit
```

**Java tools:**

```bash
which mvn >/dev/null 2>&1 || which gradle >/dev/null 2>&1
```

If missing, show:

```
✗ mvn/gradle not found
  Install Maven or Gradle for dependency checking
```

**PHP tools:**

```bash
which composer >/dev/null 2>&1
```

If missing, show:

```
✗ composer not found
  Install: https://getcomposer.org/download/
```

**Ruby tools:**

```bash
which bundle >/dev/null 2>&1
which brakeman >/dev/null 2>&1
```

If missing, show:

```
✗ bundle not found
  Install: gem install bundler
✗ brakeman not found
  Install: gem install brakeman
```

**.NET tools:**

```bash
which dotnet >/dev/null 2>&1
```

If missing, show:

```
✗ dotnet not found
  Install: https://dotnet.microsoft.com/download
```

#### Tool Check Summary

After checking all tools for detected technologies, display:

```
Security Tools Check:
━━━━━━━━━━━━━━━━━━━━━
✓ govulncheck (Go)
✗ bandit (Python) - pip install bandit
✓ npm (Node.js)
━━━━━━━━━━━━━━━━━━━━━
```

If any required tools are missing, ask user:

```
Some security tools are not installed. Options:
1. Continue anyway (tools will be skipped during audit)
2. Stop and install missing tools first (recommended)
```

If user chooses to stop, display all install commands and exit.

### Phase 3: Build Denial Patterns

Create a comprehensive deny list combining:

#### Base Security Patterns (Always Include)

**Environment Files:**

- `Read(.env)`
- `Read(**/.env)`
- `Read(.env.*)`
- `Read(**/.env.*)`
- `Read(.env.local)`
- `Read(.env.development)`
- `Read(.env.production)`
- `Read(.env.test)`

**Version Control & IDE:**

- `Read(.git/**)`
- `Read(.vscode/**)`
- `Read(.idea/**)`
- `Read(.devcontainer/**)`
- `Read(.github/workflows/**)`

**Package Management:**

- `Read(node_modules/**)`
- `Read(package-lock.json)`

**Credentials & Secrets:**

- `Read(credentials.json)`
- `Read(**/credentials.json)`
- `Read(secrets.yml)`
- `Read(**/secrets.yml)`
- `Read(config/secrets.yml)`
- `Read(.secret)`
- `Read(**/.secret)`
- `Read(*.secret)`

**SSH & Certificate Files:**

- `Read(id_rsa)`
- `Read(id_rsa.pub)`
- `Read(id_ed25519)`
- `Read(id_ed25519.pub)`
- `Read(*.pem)`
- `Read(*.key)`
- `Read(*.p12)`
- `Read(*.jks)`
- `Read(*.pfx)`
- `Read(*.keystore)`
- `Read(*.cer)`
- `Read(*.crt)`

**Cloud Provider Credentials:**

- `Read(.aws/credentials)`
- `Read(.aws/config)`
- `Read(.gcp/credentials.json)`
- `Read(.azure/credentials)`

**Database Files:**

- `Read(*.db)`
- `Read(*.sqlite)`
- `Read(*.sqlite3)`

#### Technology-Specific Patterns

**Python (if detected):**

- `Read(.venv/**)`
- `Read(venv/**)`
- `Read(__pycache__/**)`
- `Read(**/__pycache__/**)`
- `Read(*.pyc)`
- `Read(.pytest_cache/**)`
- `Read(.tox/**)`
- `Read(dist/**)`
- `Read(build/**)`
- `Read(*.egg-info/**)`
- `Read(.mypy_cache/**)`
- `Read(.ruff_cache/**)`

**.NET (if detected):**

- `Read(bin/**)`
- `Read(obj/**)`
- `Read(*.user)`
- `Read(*.suo)`
- `Read(.vs/**)`
- `Read(*.DotSettings.user)`
- `Read(TestResults/**)`
- `Read(packages/**)`

**Go (if detected):**

- `Read(vendor/**)`

**Rust (if detected):**

- `Read(target/**)`

**PHP (if detected):**

- `Read(vendor/**)`
- `Read(composer.lock)`

**Ruby (if detected):**

- `Read(vendor/bundle/**)`
- `Read(.bundle/**)`

**Java (if detected):**

- `Read(target/**)`
- `Read(*.class)`
- `Read(.gradle/**)`
- `Read(build/**)`

**Node.js (if detected):**

- `Read(node_modules/**)`
- `Read(.next/**)`
- `Read(.nuxt/**)`
- `Read(dist/**)`
- `Read(build/**)`
- `Read(.cache/**)`
- `Read(.turbo/**)`

**Docker (if detected):**

- `Read(docker-compose.override.yml)`
- `Read(docker-compose.override.yaml)`

### Phase 4: Check Existing Configuration

Check if `.claude/settings.json` already exists using the **Read tool** (NOT bash test commands):

1. Try to read `.claude/settings.json` using the Read tool
2. If the file exists and Read succeeds:
   - Parse the JSON content
   - Check for existing `permissions.deny` section
   - Ask user for merge strategy preference using AskUserQuestion tool:
     - **Deduplicate** (default): Remove duplicate patterns, add only new ones
     - **Append**: Add all new patterns, keep duplicates
     - **Replace**: Completely replace existing deny section with new patterns
3. If the file doesn't exist (Read returns error):
   - Proceed to create new file with deny patterns
   - Use "Deduplicate" as the default strategy

**IMPORTANT**:

- Use **Read tool** to check file existence - DO NOT use bash test commands
- The Read tool will gracefully handle non-existent files by returning an error
- Parse existing JSON to preserve non-permission settings

### Phase 5: Show Preview & Get Confirmation

Display a comprehensive preview showing:

1. **Technologies Detected:**
   - List all detected technologies with file indicators

2. **Current Configuration (if exists):**
   - Show current deny patterns count
   - Show sample of existing patterns (first 5)

3. **Proposed Changes:**
   - Show all new patterns to be added
   - Group by category (Base Security, Python, .NET, etc.)
   - Show total pattern count

4. **After Configuration:**
   - Show total pattern count after merge
   - Show merge strategy being used

Ask for user confirmation before proceeding.

### Phase 6: Write Configuration

After user confirms:

1. Create `.claude/` directory if it doesn't exist using the Bash tool: `mkdir -p .claude`
2. Write or update `settings.json` using the **Write tool** (NOT bash echo or heredoc)
3. Preserve any other existing settings (don't overwrite non-permission settings)
4. Format JSON with proper indentation (2 spaces)
5. Show success message with:
   - File path: `.claude/settings.json`
   - Total deny patterns configured
   - Technologies covered

**IMPORTANT**:

- Use **Write tool** to create/update the settings file
- Use Bash tool only for creating the `.claude/` directory if needed
- Ensure proper JSON formatting with 2-space indentation

### Important Constraints

**DO NOT:**

- Read the contents of any sensitive files during scanning
- Include file paths from the actual project in the deny list
- Overwrite other settings in settings.json (preserve everything except permissions.deny)
- Proceed without user confirmation
- Use bash test commands (`test -f`, `[ -f ]`, etc.) - they trigger permission prompts
- Use any bash commands for file detection or checking

**DO:**

- Use **Glob tool** for technology detection (file pattern matching)
- Use **Read tool** to check if `.claude/settings.json` exists (handles errors gracefully)
- Use **Write tool** to create/update settings.json
- Use **AskUserQuestion tool** to ask for merge strategy preference
- Deduplicate patterns by default during merge
- Show clear before/after comparison
- Maintain alphabetical ordering within categories for readability
- Use forward slashes in all patterns for cross-platform compatibility

### Example Output Format

```
🔍 Detecting technologies in your project...

Technologies Detected:
✓ Node.js (package.json found)
✓ TypeScript (tsconfig.json found)
✓ Python (requirements.txt, pyproject.toml found)
✓ Docker (Dockerfile, docker-compose.yml found)

Current Configuration:
📄 .claude/settings.json exists
📊 Current deny patterns: 8

Proposed Security Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Base Security Patterns (25):
  • Environment files (.env, .env.*)
  • Version control (.git, .vscode, .idea)
  • Credentials (credentials.json, secrets.yml)
  • SSH & certificates (*.pem, *.key, id_rsa)
  • Cloud provider configs (.aws/credentials, .gcp/*)
  • Database files (*.db, *.sqlite)

Node.js Patterns (8):
  • node_modules/**
  • .next/**, .nuxt/**
  • dist/**, build/**
  • .cache/**, .turbo/**

Python Patterns (11):
  • .venv/**, venv/**
  • __pycache__/**, *.pyc
  • .pytest_cache/**, .tox/**
  • dist/**, *.egg-info/**

Docker Patterns (2):
  • docker-compose.override.yml

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total new patterns to add: 46
After merge: 54 total patterns

Merge Strategy: Deduplicate (remove duplicates, add only new patterns)

Would you like to proceed with this configuration? (yes/no)
```

### Phase 7: Update Security Audit with Technology-Specific Tools

After writing settings.json, update `.claude/security-audit/commands/security-audit.md` to include technology-specific security scanning tools.

Read the current security-audit.md file, then append technology-specific entries to the "Analysis Scope" section.

#### Technology-Specific Security Tools

**Go (if detected):**

```markdown
6. **Go Security**:
   - Run `govulncheck ./...` to check for known vulnerabilities in dependencies
   - Check for unsafe package usage
   - Review error handling patterns (no swallowed errors)
   - Verify no hardcoded credentials in source
   - Run gosec ./... to static check for vurnerabilities withing the project
```

**Python (if detected):**

```markdown
6. **Python Security**:
   - Run `bandit -r .` for static security analysis
   - Run `pip-audit` or `safety check` for dependency vulnerabilities
   - Check for pickle usage, eval/exec calls, SQL string formatting
```

**Node.js (if detected):**

```markdown
6. **Node.js Security**:
   - Run `npm audit` or `yarn audit` for dependency vulnerabilities
   - Check for prototype pollution patterns
   - Review for unsafe RegEx (ReDoS)
   - Verify no `eval()`, `Function()` constructor usage
```

**Rust (if detected):**

```markdown
6. **Rust Security**:
   - Run `cargo audit` for dependency vulnerabilities
   - Check for unsafe blocks and document justification
   - Review for panic in library code
```

**Java (if detected):**

```markdown
6. **Java Security**:
   - Run OWASP Dependency-Check or `mvn dependency:tree`
   - Check for deserialization vulnerabilities
   - Review for SQL injection in JDBC calls
   - Verify no hardcoded secrets in properties files
```

**PHP (if detected):**

```markdown
6. **PHP Security**:
   - Run `composer audit` for dependency vulnerabilities
   - Check for `eval()`, `exec()`, `system()` usage
   - Review for SQL injection, XSS in templates
```

**Ruby (if detected):**

```markdown
6. **Ruby Security**:
   - Run `bundle audit` for dependency vulnerabilities
   - Run `brakeman` for Rails security analysis
   - Check for mass assignment vulnerabilities
```

**.NET (if detected):**

```markdown
6. **.NET Security**:
   - Run `dotnet list package --vulnerable`
   - Check for SQL injection in Entity Framework raw queries
   - Review for XXE in XML parsing
   - Verify anti-forgery tokens in forms
```

#### Update Process

1. Read `.claude/security-audit/commands/security-audit.md`
2. Find the "Analysis Scope" section
3. Append only the entries for detected technologies (don't duplicate if already present)
4. Write the updated file

### Phase 8: Update Security Upgrade with Technology-Specific Fix Patterns

Also update `.claude/security-audit/commands/security-upgrade.md` and `.claude/skills/security-upgrade/SKILL.md` to include technology-specific remediation guidance.

Append to the "Important Notes" section:

#### Technology-Specific Remediation Patterns

**Go (if detected):**

```markdown
### Go-Specific Fixes

- **Dependency vulnerabilities**: Run `go get -u` to update, or pin specific versions in go.mod
- **Unsafe package**: Replace with safe alternatives or add security review comment
- **Error swallowing**: Ensure all errors are handled or explicitly ignored with `_ =`
- **Hardcoded secrets**: Move to environment variables, use `os.Getenv()`
```

**Python (if detected):**

```markdown
### Python-Specific Fixes

- **Dependency vulnerabilities**: Update in requirements.txt/pyproject.toml, run `pip install --upgrade`
- **Pickle usage**: Replace with `json` for untrusted data
- **eval/exec**: Replace with `ast.literal_eval()` or safer alternatives
- **SQL formatting**: Use parameterized queries with `?` or `%s` placeholders
```

**Node.js (if detected):**

```markdown
### Node.js-Specific Fixes

- **Dependency vulnerabilities**: Run `npm audit fix` or update package.json manually
- **Prototype pollution**: Use `Object.create(null)` or Map for user-controlled keys
- **ReDoS**: Simplify regex or use `re2` package
- **eval usage**: Replace with `JSON.parse()` or structured alternatives
```

**Rust (if detected):**

```markdown
### Rust-Specific Fixes

- **Dependency vulnerabilities**: Update Cargo.toml, run `cargo update`
- **Unsafe blocks**: Add `// SAFETY:` comment documenting invariants
- **Panic in library**: Replace `unwrap()`/`expect()` with `?` or `Result` returns
```

**Java (if detected):**

```markdown
### Java-Specific Fixes

- **Dependency vulnerabilities**: Update pom.xml/build.gradle versions
- **Deserialization**: Use allowlists, avoid `ObjectInputStream` on untrusted data
- **SQL injection**: Use PreparedStatement with `?` parameters
- **Hardcoded secrets**: Move to environment variables or secrets manager
```

**PHP (if detected):**

```markdown
### PHP-Specific Fixes

- **Dependency vulnerabilities**: Run `composer update` or pin versions
- **Command injection**: Use `escapeshellarg()`, avoid `exec()`/`system()` with user input
- **SQL injection**: Use PDO prepared statements
- **XSS**: Use `htmlspecialchars()` or templating engine auto-escaping
```

**Ruby (if detected):**

```markdown
### Ruby-Specific Fixes

- **Dependency vulnerabilities**: Run `bundle update` or pin versions in Gemfile
- **Mass assignment**: Use `strong_parameters`, define `permit()` explicitly
- **SQL injection**: Use ActiveRecord query interface, avoid string interpolation
- **Command injection**: Use `Open3` with array arguments instead of shell strings
```

**.NET (if detected):**

```markdown
### .NET-Specific Fixes

- **Dependency vulnerabilities**: Update package versions in .csproj
- **SQL injection**: Use Entity Framework LINQ or parameterized `SqlCommand`
- **XXE**: Set `DtdProcessing = DtdProcessing.Prohibit` on XmlReader
- **Missing CSRF**: Add `[ValidateAntiForgeryToken]` attribute to POST actions
```

#### Update Process for Security Upgrade

1. Read `.claude/security-audit/commands/security-upgrade.md`
2. Find the "Important Notes" section
3. Append technology-specific fix patterns for detected technologies
4. Write the updated file
5. Repeat for `.claude/skills/security-upgrade/SKILL.md` if it exists

### Success Message Format

```
✓ Security configuration successfully initialized!

Configuration Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 File: .claude/settings.json
📊 Total deny patterns: 54
🛡️ Technologies covered: Node.js, TypeScript, Python, Docker

Security Workflows Updated:
🔍 /security-audit - Added technology-specific scanning tools
🔧 /security-upgrade - Added technology-specific fix patterns

Tools added for Go:
  • govulncheck for vulnerability scanning
  • Fix patterns for unsafe pkg, error handling, secrets

Tools added for Python:
  • bandit, pip-audit for scanning
  • Fix patterns for pickle, eval, SQL injection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  IMPORTANT: You must restart Claude Code for these settings to take effect.

After restarting:
- Claude Code will avoid reading sensitive files, credentials, and build artifacts
- You can manually edit .claude/settings.json to customize these settings
- Run /security-audit to perform a comprehensive security analysis
- Run /security-upgrade to apply fixes from audit reports
```
