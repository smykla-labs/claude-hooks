package git

import (
	"fmt"
	"sort"
	"strings"

	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
	"github.com/smykla-labs/claude-hooks/pkg/parser"
)

const (
	defaultRemote = "origin"
)

// PushValidator validates git push commands
type PushValidator struct {
	validator.BaseValidator
	gitRunner GitRunner
}

// NewPushValidator creates a new PushValidator instance
func NewPushValidator(log logger.Logger, gitRunner GitRunner) *PushValidator {
	if gitRunner == nil {
		gitRunner = NewRealGitRunner()
	}
	return &PushValidator{
		BaseValidator: *validator.NewBaseValidator("validate-git-push", log),
		gitRunner:     gitRunner,
	}
}

// Name returns the validator name
func (v *PushValidator) Name() string {
	return "validate-git-push"
}

// Validate validates git push commands
func (v *PushValidator) Validate(ctx *hook.Context) *validator.Result {
	log := v.Logger()
	command := ctx.GetCommand()
	if command == "" {
		return validator.Pass()
	}

	bashParser := parser.NewBashParser()
	parseResult, err := bashParser.Parse(command)
	if err != nil {
		log.Debug("failed to parse command", "error", err)
		return validator.Pass()
	}

	for _, cmd := range parseResult.Commands {
		if cmd.Name != "git" || len(cmd.Args) == 0 {
			continue
		}

		gitCmd, parseErr := parser.ParseGitCommand(cmd)
		if parseErr != nil {
			log.Debug("failed to parse git command", "error", parseErr)
			continue
		}

		if gitCmd.Subcommand != "push" {
			continue
		}

		result := v.validatePushCommand(gitCmd)
		if !result.Passed {
			return result
		}
	}

	return validator.Pass()
}

// validatePushCommand validates a single git push command
func (v *PushValidator) validatePushCommand(gitCmd *parser.GitCommand) *validator.Result {
	log := v.Logger()

	if !v.gitRunner.IsInRepo() {
		log.Debug("not in a git repository, skipping validation")
		return validator.Pass()
	}

	remote := v.extractRemote(gitCmd)
	if remote == "" {
		log.Debug("no remote specified, skipping validation")
		return validator.Pass()
	}

	if result := v.validateRemoteExists(remote); !result.Passed {
		return result
	}

	repoRoot, err := v.gitRunner.GetRepoRoot()
	if err != nil {
		log.Debug("failed to get repo root", "error", err)
		return validator.Pass()
	}

	projectType := detectProjectType(repoRoot)
	return v.validateProjectSpecificRules(projectType, remote)
}

// extractRemote extracts the remote name from a git push command
func (v *PushValidator) extractRemote(gitCmd *parser.GitCommand) string {
	if len(gitCmd.Args) == 0 {
		branch, err := v.gitRunner.GetCurrentBranch()
		if err != nil {
			return defaultRemote
		}

		remote, err := v.gitRunner.GetBranchRemote(branch)
		if err != nil {
			return defaultRemote
		}

		return remote
	}

	for _, arg := range gitCmd.Args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}

	return ""
}

// validateRemoteExists checks if the remote exists
func (v *PushValidator) validateRemoteExists(remote string) *validator.Result {
	_, err := v.gitRunner.GetRemoteURL(remote)
	if err != nil {
		remotes, remoteErr := v.gitRunner.GetRemotes()
		if remoteErr != nil {
			return validator.Fail(
				fmt.Sprintf("🚫 Git push validation failed:\n\n❌ Remote '%s' does not exist", remote),
			)
		}

		return validator.Fail(
			"🚫 Git push validation failed:\n\n" + v.formatRemoteNotFoundError(remote, remotes),
		)
	}

	return validator.Pass()
}

// formatRemoteNotFoundError formats the error message for a missing remote
func (v *PushValidator) formatRemoteNotFoundError(remote string, remotes map[string]string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("❌ Remote '%s' does not exist\n\n", remote))
	msg.WriteString("Available remotes:\n")

	names := make([]string, 0, len(remotes))
	for name := range remotes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		url := remotes[name]
		msg.WriteString(fmt.Sprintf("  %s  %s\n", name, url))
	}

	msg.WriteString("\nUse 'git remote -v' to list all configured remotes.")

	return msg.String()
}

// detectProjectType detects the project type based on the repo root path
func detectProjectType(repoRoot string) string {
	if strings.Contains(repoRoot, "/kumahq/kuma") {
		return "kumahq/kuma"
	}
	if strings.Contains(repoRoot, "/kong/") || strings.Contains(repoRoot, "/Kong/") {
		return "kong-org"
	}
	return ""
}

// validateProjectSpecificRules validates project-specific push rules
func (v *PushValidator) validateProjectSpecificRules(projectType, remote string) *validator.Result {
	switch projectType {
	case "kong-org":
		return v.validateKongOrgPush(remote)
	case "kumahq/kuma":
		return v.validateKumaPush(remote)
	default:
		return validator.Pass()
	}
}

// validateKongOrgPush validates Kong organization push rules
func (v *PushValidator) validateKongOrgPush(remote string) *validator.Result {
	if remote == "origin" {
		return validator.Fail(
			`🚫 Git push validation failed:

❌ Kong org projects should push to 'upstream' remote (main repo)
   Note: 'origin' is your fork, use 'upstream' for Kong repos

Expected: git push upstream branch-name`,
		)
	}

	return validator.Pass()
}

// validateKumaPush validates kumahq/kuma push rules
func (v *PushValidator) validateKumaPush(remote string) *validator.Result {
	if remote == "upstream" {
		return &validator.Result{
			Passed: false,
			Message: `⚠️  Warning: Pushing to 'upstream' remote in kumahq/kuma
   This should only be done when explicitly intended
   Normal workflow: push to 'origin' (your fork)`,
			ShouldBlock: false,
		}
	}

	return validator.Pass()
}
