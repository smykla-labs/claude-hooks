# GitHub Quality & CI Architecture

GitHub repository configuration, CI workflows, and automation patterns for quality and security.

## Core Design Philosophy

**OSSF Scorecard Driven**: Target score ≥7/10. Design decisions prioritize scorecard checks (Branch-Protection, Code-Review, Maintained, SBOM, Pinned-Dependencies).

**Branch Rulesets Over Legacy Protection**: Use new rulesets API instead of deprecated branch protection API. Rulesets provide better granularity and organization-level inheritance.

**Version Sync Across Files**: Single source of truth for tool versions via Renovate's `customManagers:githubActionsVersions`. Versions declared in workflow env vars, synced to `.mise.toml`.

**Bot-Driven Automation**: Smyklot bot handles PR commands and reaction-based auto-merge. Reduces manual PR management overhead.

## OSSF Scorecard Integration

Target score: ≥7/10. Runs weekly, uploads SARIF to Security tab.

```yaml
# .github/workflows/scorecard.yml
name: Scorecard analysis
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  push:
    branches: [main]

jobs:
  analysis:
    runs-on: ubuntu-24.04
    permissions:
      security-events: write
      id-token: write
    steps:
      - uses: ossf/scorecard-action@<sha>  # v2.4.3
        with:
          results_file: results.sarif
          publish_results: true
      - uses: github/codeql-action/upload-sarif@<sha>
        with:
          sarif_file: results.sarif
```

**Why Weekly**: Checks like Maintained and Pinned-Dependencies change over time. Weekly schedule catches regressions early.

## Branch Rulesets (Modern Approach)

Use rulesets API instead of legacy branch protection:

```bash
# Create ruleset via GitHub API
gh api repos/OWNER/REPO/rulesets -X POST --input - <<'EOF'
{
  "name": "main-protection",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["refs/heads/main"],
      "exclude": []
    }
  },
  "rules": [
    {
      "type": "pull_request",
      "parameters": {
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": true,
        "require_last_push_approval": false,
        "required_approving_review_count": 1,
        "required_review_thread_resolution": true
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "required_status_checks": [
          {"context": "Test"},
          {"context": "Lint"}
        ],
        "strict_required_status_checks_policy": true
      }
    },
    {"type": "deletion"},
    {"type": "non_fast_forward"}
  ]
}
EOF
```

**Why Rulesets vs Legacy Branch Protection**:

- Rulesets support inheritance across organization
- Better UI in repository settings
- More granular conditions (tag protection, etc.)
- Legacy API will be deprecated eventually

**Gotcha**: Status check context names must match workflow job names exactly. Check with `gh api repos/OWNER/REPO/commits/COMMIT_SHA/status`.

## GitHub Security Features

Enable secret scanning and dependabot via API:

```bash
# Enable secret scanning + push protection
gh api repos/OWNER/REPO -X PATCH \
  -f "security_and_analysis[secret_scanning][status]=enabled" \
  -f "security_and_analysis[secret_scanning_push_protection][status]=enabled"

# Enable dependabot security updates
gh api repos/OWNER/REPO/automated-security-fixes -X PUT

# Verify settings
gh api repos/OWNER/REPO --jq '.security_and_analysis'
```

**Why API Configuration**: Reproducible via scripts. Can be templated for multiple repositories.

## Renovate Version Synchronization

Problem: Tool versions duplicated across `.mise.toml` and GitHub Actions workflows. Renovate updates one but not the other.

Solution: Use `customManagers:githubActionsVersions` preset to extract versions from workflow env vars with renovate comments.

### Workflow Pattern

```yaml
# .github/workflows/lint.yml
env:
  # renovate: datasource=github-releases depName=golangci/golangci-lint
  GOLANGCI_LINT_VERSION: "2.6.2"

jobs:
  lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: golangci/golangci-lint-action@<sha>
        with:
          version: v${{ env.GOLANGCI_LINT_VERSION }}
```

### Renovate Configuration

```json
{
  "extends": [
    "config:recommended",
    "customManagers:githubActionsVersions"
  ],
  "packageRules": [
    {
      "description": "Group golangci-lint across mise and CI workflow",
      "matchPackageNames": ["golangci-lint", "golangci/golangci-lint"],
      "groupName": "golangci-lint"
    }
  ]
}
```

**How It Works**:

1. `customManagers:githubActionsVersions` preset extracts version from renovate comment in workflow file
2. Renovate detects version in both `.mise.toml` and workflow env var
3. `groupName` rule creates single PR updating both files

**Gotcha**: Package names must match exactly. Use `matchPackageNames` array to include both `golangci-lint` (mise) and `golangci/golangci-lint` (GitHub releases).

## Dependency Review Action

The `deny-licenses` option is deprecated in v4.x (removed in v5.x):

```yaml
# DEPRECATED - Don't use deny-licenses
- uses: actions/dependency-review-action@<sha>
  with:
    deny-licenses: GPL-3.0, AGPL-3.0  # Deprecated

# RECOMMENDED - Use allow-licenses or omit
- uses: actions/dependency-review-action@<sha>
  with:
    fail-on-severity: high
    comment-summary-in-pr: always
    # allow-licenses: MIT, Apache-2.0, BSD-3-Clause  # Optional
```

**Why Deprecated**: Maintaining deny-list is error-prone. Allow-list (or severity-based blocking) is more maintainable.

## Smyklot Bot Workflows

Two workflows for PR automation via Smyklot app:

### PR Commands (`pr-commands.yml`)

Handles slash commands in PR comments:

```yaml
# Triggered on issue_comment created
on:
  issue_comment:
    types: [created]

jobs:
  handle-command:
    if: github.event.issue.pull_request && startsWith(github.event.comment.body, '/')
    runs-on: ubuntu-24.04
    steps:
      - uses: smykla/smyklot-action@<sha>
        with:
          app-id: ${{ secrets.SMYKLOT_APP_ID }}
          private-key: ${{ secrets.SMYKLOT_PRIVATE_KEY }}
          config: ${{ vars.SMYKLOT_CONFIG }}
```

### Reaction Polling (`poll-reactions.yml`)

Polls PR reactions every 5 minutes for auto-merge:

```yaml
# Scheduled polling
on:
  schedule:
    - cron: '*/5 * * * *'  # Every 5 minutes

jobs:
  poll:
    runs-on: ubuntu-24.04
    steps:
      - uses: smykla/smyklot-action@<sha>
        with:
          mode: poll-reactions
          app-id: ${{ secrets.SMYKLOT_APP_ID }}
          private-key: ${{ secrets.SMYKLOT_PRIVATE_KEY }}
```

**Why Two Workflows**: Commands need instant response (triggered on comment). Auto-merge based on reactions uses polling (no webhook for reaction changes).

**Required Secrets**:

- `SMYKLOT_APP_ID`: GitHub App ID
- `SMYKLOT_PRIVATE_KEY`: GitHub App private key (PEM format)
- `SMYKLOT_CONFIG`: Bot configuration JSON (repository variable)

## Workflow File Conventions

Repository-specific conventions:

```yaml
# File extension: .yml (not .yaml)
# File: .github/workflows/test.yml

# Action pinning: SHA + version comment
- uses: actions/checkout@<sha>  # v4.1.0

# Runner: Use latest Ubuntu LTS
runs-on: ubuntu-24.04

# Permissions: Explicit, minimal scope
permissions:
  contents: read
  pull-requests: write
```

**Why SHA Pinning**: Security best practice. Prevents supply chain attacks via compromised action tags.

**Why ubuntu-24.04**: Latest LTS. Older images (20.04, 22.04) may have outdated tools.

## Common Pitfalls

1. **Using legacy branch protection API**: New rulesets API provides better features and organization inheritance. Use rulesets for new repositories.

2. **Mismatched status check names**: Ruleset `required_status_checks` context must match workflow job name exactly. Check actual contexts with commits API.

3. **deny-licenses in dependency-review v4+**: Option deprecated and removed. Use `allow-licenses` or omit license checking.

4. **Not grouping renovate updates**: Without `packageRules.groupName`, Renovate creates separate PRs for `.mise.toml` and workflow updates. Group by tool name.

5. **Missing renovate comment in workflow**: `customManagers:githubActionsVersions` requires special comment format. Without it, Renovate won't detect version.

6. **Hardcoding versions in workflow**: Use env vars at workflow level, not inline in steps. Allows Renovate to update single location.

7. **Wrong smyklot workflow triggers**: Commands need `issue_comment.created`, reactions need scheduled polling. Mixing triggers breaks functionality.

8. **Not pinning actions to SHA**: Pinning to tags (`v4`) allows malicious updates. Always pin to SHA with version comment.

9. **Insufficient workflow permissions**: Explicitly set minimal permissions. GITHUB_TOKEN has broad default permissions in older workflows.
