# Session: GitHub Quality & CI Workflows

## OSSF Scorecard

Target score: ≥7/10. Key checks: Branch-Protection, Code-Review, Maintained, SBOM, Pinned-Dependencies.

```yaml
# scorecard.yml - runs weekly, uploads SARIF to Security tab
uses: ossf/scorecard-action@<sha> # v2.4.3
```

## Branch Rulesets (New Approach)

Use rulesets instead of legacy branch protection:

```bash
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

## GitHub Security Features via API

```bash
# Enable secret scanning + push protection
gh api repos/OWNER/REPO -X PATCH \
  -f "security_and_analysis[secret_scanning][status]=enabled" \
  -f "security_and_analysis[secret_scanning_push_protection][status]=enabled"

# Enable dependabot security updates
gh api repos/OWNER/REPO/automated-security-fixes -X PUT
```

## Renovate: Version Sync Across Files

Use `customManagers:githubActionsVersions` preset to sync tool versions between `.mise.toml` and CI workflows.

**Workflow pattern** (env var with renovate comment):

```yaml
env:
  # renovate: datasource=github-releases depName=golangci/golangci-lint
  GOLANGCI_LINT_VERSION: "2.6.2"

jobs:
  lint:
    steps:
      - uses: golangci/golangci-lint-action@<sha>
        with:
          version: v${{ env.GOLANGCI_LINT_VERSION }}
```

**renovate.json** grouping rule:

```json
{
  "extends": ["customManagers:githubActionsVersions"],
  "packageRules": [
    {
      "description": "Group golangci-lint across mise and CI workflow",
      "matchPackageNames": ["golangci-lint", "golangci/golangci-lint"],
      "groupName": "golangci-lint"
    }
  ]
}
```

## Dependency Review Action

The `deny-licenses` option is deprecated in v4.x (removed in v5.x). Use allow-list approach or omit license checking:

```yaml
# Don't use: deny-licenses: GPL-3.0, AGPL-3.0
- uses: actions/dependency-review-action@<sha>
  with:
    fail-on-severity: high
    comment-summary-in-pr: always
```

## Smyklot Bot Workflows

Two workflows for PR automation:

1. **pr-commands.yml** - Handles bot commands from PR comments
2. **poll-reactions.yml** - Polls reactions every 5 minutes for auto-merge

Required secrets/vars: `SMYKLOT_APP_ID`, `SMYKLOT_PRIVATE_KEY`, `SMYKLOT_CONFIG`

## Workflow File Conventions

- This repo uses `.yml` extension (not `.yaml`)
- Pin actions to SHA digests with version comment: `@<sha> # v1.2.3`
- Use `ubuntu-24.04` runner
