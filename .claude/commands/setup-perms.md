---
description: "Set up safe Claude Code permissions for the current project — scopes Edit/Write to the project dir, allows read tools everywhere, and allows safe git/go commands."
allowed-tools: Bash(pwd), Write
---

# Setup Permissions

## Step 1: Get project root

```bash
pwd
```

Use the output as `$PROJECT_ROOT`.

## Step 2: Write .claude/settings.json

Merge the following into `.claude/settings.json`, preserving any existing keys (like `enabledPlugins`). Read the file first if it exists, then write the merged result.

The permissions block to set:

```json
{
  "permissions": {
    "allow": [
      "Read(*)",
      "Edit($PROJECT_ROOT/*)",
      "Write($PROJECT_ROOT/*)",
      "Glob(*)",
      "Grep(*)",
      "Bash(go build *)",
      "Bash(go test *)",
      "Bash(go run *)",
      "Bash(go vet *)",
      "Bash(go fmt *)",
      "Bash(go mod *)",
      "Bash(git status)",
      "Bash(git diff *)",
      "Bash(git log *)",
      "Bash(git show *)",
      "Bash(git add *)",
      "Bash(git commit *)"
    ]
  }
}
```

Replace `$PROJECT_ROOT` with the actual path from Step 1.

## Step 3: Confirm

Tell the user the permissions have been set up, listing what is auto-approved and what still requires confirmation (push, reset, arbitrary bash).
