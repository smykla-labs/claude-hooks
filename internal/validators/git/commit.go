package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/smykla-labs/klaudiush/internal/templates"
	"github.com/smykla-labs/klaudiush/internal/validator"
	"github.com/smykla-labs/klaudiush/pkg/hook"
	"github.com/smykla-labs/klaudiush/pkg/logger"
	"github.com/smykla-labs/klaudiush/pkg/parser"
)

const (
	gitCommand       = "git"
	commitSubcommand = "commit"
	addSubcommand    = "add"
)

var (
	// Commit message flags for inline messages.
	commitMessageFlags = []string{"-m", "--message"}

	// Commit file flags for message from file.
	commitFileFlags = []string{"-F", "--file"}
)

// CommitValidator validates git commit commands and messages
type CommitValidator struct {
	validator.BaseValidator
	gitRunner GitRunner
}

// NewCommitValidator creates a new CommitValidator instance
func NewCommitValidator(log logger.Logger, gitRunner GitRunner) *CommitValidator {
	if gitRunner == nil {
		gitRunner = NewGitRunner()
	}

	return &CommitValidator{
		BaseValidator: *validator.NewBaseValidator("validate-commit", log),
		gitRunner:     gitRunner,
	}
}

// Validate checks git commit command and message
func (v *CommitValidator) Validate(_ context.Context, hookCtx *hook.Context) *validator.Result {
	log := v.Logger()
	log.Debug("Running git commit validation")

	// Parse the command
	bashParser := parser.NewBashParser()

	result, err := bashParser.Parse(hookCtx.GetCommand())
	if err != nil {
		log.Error("Failed to parse command", "error", err)
		return validator.Warn(fmt.Sprintf("Failed to parse command: %v", err))
	}

	// Check if there's a git add in the same command chain
	hasGitAdd := v.hasGitAddInChain(result.Commands)

	// Find git commit commands
	for _, cmd := range result.Commands {
		if cmd.Name != gitCommand || len(cmd.Args) == 0 || cmd.Args[0] != commitSubcommand {
			continue
		}

		// Parse git command for flags and message
		gitCmd, err := parser.ParseGitCommand(cmd)
		if err != nil {
			log.Error("Failed to parse git command", "error", err)
			return validator.Warn(fmt.Sprintf("Failed to parse git command: %v", err))
		}

		// Check -sS flags
		if res := v.checkFlags(gitCmd); !res.Passed {
			return res
		}

		// Check staging area (skip for --amend, --allow-empty, or if git add is in the chain)
		if !gitCmd.HasFlag("--amend") && !gitCmd.HasFlag("--allow-empty") && !hasGitAdd {
			if res := v.checkStagingArea(gitCmd); !res.Passed {
				return res
			}
		}

		// Extract and validate commit message
		commitMsg, err := v.extractCommitMessage(gitCmd)
		if err != nil {
			log.Error("Failed to extract commit message", "error", err)
			return validator.Fail(fmt.Sprintf("Failed to read commit message: %v", err))
		}

		if commitMsg == "" {
			// No message flag, message will come from editor
			log.Debug("No message flag, message will come from editor")
			return validator.Pass()
		}

		// Validate the commit message
		return v.validateMessage(commitMsg)
	}

	log.Debug("No git commit commands found")

	return validator.Pass()
}

// checkFlags validates that the commit command has -sS flags
func (*CommitValidator) checkFlags(gitCmd *parser.GitCommand) *validator.Result {
	// Check for -s (signoff) and -S (GPG sign)
	hasSignoff := gitCmd.HasFlag("-s") || gitCmd.HasFlag("--signoff")
	hasGPGSign := gitCmd.HasFlag("-S") || gitCmd.HasFlag("--gpg-sign")

	if !hasSignoff || !hasGPGSign {
		message := templates.MustExecute(
			templates.GitCommitFlagsTemplate,
			templates.GitCommitFlagsData{
				ArgsStr: strings.Join(gitCmd.Args, " "),
			},
		)

		return validator.Fail(
			"Git commit must use -sS flags",
		).AddDetail("help", message)
	}

	return validator.Pass()
}

// checkStagingArea validates that there are files staged or -a/-A/--all flag is present
func (v *CommitValidator) checkStagingArea(gitCmd *parser.GitCommand) *validator.Result {
	// Check if -a, -A, or --all flags are present
	hasStageFlag := gitCmd.HasFlag("-a") || gitCmd.HasFlag("-A") || gitCmd.HasFlag("--all")
	if hasStageFlag {
		return validator.Pass()
	}

	// Check if we're in a git repository first
	if !v.gitRunner.IsInRepo() {
		// Not in a git repo or git not available, skip check
		v.Logger().Debug("Not in git repository, skipping staging check")
		return validator.Pass()
	}

	// Check if staging area has files
	stagedFiles, err := v.gitRunner.GetStagedFiles()
	if err != nil {
		v.Logger().Debug("Failed to check staging area", "error", err)
		return validator.Pass() // Don't block if we can't check
	}

	if len(stagedFiles) == 0 {
		// No files staged, get status info
		modifiedCount, untrackedCount := v.getStatusCounts()

		message := templates.MustExecute(
			templates.GitCommitNoStagedTemplate,
			templates.GitCommitNoStagedData{
				ModifiedCount:  modifiedCount,
				UntrackedCount: untrackedCount,
			},
		)

		return validator.Fail(
			"No files staged for commit",
		).AddDetail("help", message)
	}

	return validator.Pass()
}

// getStatusCounts returns the count of modified and untracked files
func (v *CommitValidator) getStatusCounts() (modified, untracked int) {
	// Get modified files
	modifiedFiles, err := v.gitRunner.GetModifiedFiles()
	if err == nil {
		modified = len(modifiedFiles)
	}

	// Get untracked files
	untrackedFiles, err2 := v.gitRunner.GetUntrackedFiles()
	if err2 == nil {
		untracked = len(untrackedFiles)
	}

	return modified, untracked
}

// hasGitAddInChain checks if there's a git add command in the command chain
// This is important because in PreToolUse hooks, the add hasn't executed yet,
// so we shouldn't check the staging area.
func (*CommitValidator) hasGitAddInChain(commands []parser.Command) bool {
	for _, cmd := range commands {
		if cmd.Name == gitCommand && len(cmd.Args) > 0 && cmd.Args[0] == addSubcommand {
			return true
		}
	}

	return false
}

// extractCommitMessage extracts commit message from -m/--message or -F/--file flags.
func (v *CommitValidator) extractCommitMessage(gitCmd *parser.GitCommand) (string, error) {
	// Check for file flags first (-F/--file)
	if filePath := v.getFlagValue(gitCmd, commitFileFlags); filePath != "" {
		v.Logger().Debug("Reading commit message from file", "path", filePath)

		content, err := os.ReadFile(
			filePath,
		) //#nosec G304 -- file path is user-provided from git commit -F flag
		if err != nil {
			return "", fmt.Errorf("failed to read commit message file %s: %w", filePath, err)
		}

		return strings.TrimSpace(string(content)), nil
	}

	// Check for inline message flags (-m/--message)
	if msg := v.getFlagValue(gitCmd, commitMessageFlags); msg != "" {
		return msg, nil
	}

	return "", nil
}

// getFlagValue returns the value for any of the provided flags, or empty string if not found.
func (*CommitValidator) getFlagValue(gitCmd *parser.GitCommand, flags []string) string {
	for _, flag := range flags {
		if value := gitCmd.GetFlagValue(flag); value != "" {
			return value
		}
	}

	return ""
}

// Ensure CommitValidator implements validator.Validator
var _ validator.Validator = (*CommitValidator)(nil)
