package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// GitRunner defines the interface for git operations
type GitRunner interface {
	// IsInRepo checks if we're in a git repository
	IsInRepo() bool

	// GetStagedFiles returns the list of staged files
	GetStagedFiles() ([]string, error)

	// GetModifiedFiles returns the list of modified but unstaged files
	GetModifiedFiles() ([]string, error)

	// GetUntrackedFiles returns the list of untracked files
	GetUntrackedFiles() ([]string, error)

	// GetRepoRoot returns the git repository root directory
	GetRepoRoot() (string, error)

	// GetRemoteURL returns the URL for the given remote
	GetRemoteURL(remote string) (string, error)

	// GetCurrentBranch returns the current branch name
	GetCurrentBranch() (string, error)

	// GetBranchRemote returns the tracking remote for the given branch
	GetBranchRemote(branch string) (string, error)

	// GetRemotes returns the list of all remotes with their URLs
	GetRemotes() (map[string]string, error)
}

// RealGitRunner implements GitRunner using actual git commands
type RealGitRunner struct {
	timeout time.Duration
}

// NewRealGitRunner creates a new RealGitRunner instance
func NewRealGitRunner() *RealGitRunner {
	return &RealGitRunner{
		timeout: gitCommandTimeout,
	}
}

// IsInRepo checks if we're in a git repository
func (r *RealGitRunner) IsInRepo() bool {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetStagedFiles returns the list of staged files
func (r *RealGitRunner) GetStagedFiles() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.TrimSpace(string(output))
	if files == "" {
		return []string{}, nil
	}

	return strings.Split(files, "\n"), nil
}

// GetModifiedFiles returns the list of modified but unstaged files
func (r *RealGitRunner) GetModifiedFiles() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.TrimSpace(string(output))
	if files == "" {
		return []string{}, nil
	}

	return strings.Split(files, "\n"), nil
}

// GetUntrackedFiles returns the list of untracked files
func (r *RealGitRunner) GetUntrackedFiles() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.TrimSpace(string(output))
	if files == "" {
		return []string{}, nil
	}

	return strings.Split(files, "\n"), nil
}

// GetRepoRoot returns the git repository root directory
func (r *RealGitRunner) GetRepoRoot() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the URL for the given remote
func (r *RealGitRunner) GetRemoteURL(remote string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func (r *RealGitRunner) GetCurrentBranch() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetBranchRemote returns the tracking remote for the given branch
func (r *RealGitRunner) GetBranchRemote(branch string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	configKey := "branch." + branch + ".remote"
	//nolint:gosec // configKey is constructed from trusted input, not user-provided
	cmd := exec.CommandContext(ctx, "git", "config", configKey)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemotes returns the list of all remotes with their URLs
func (r *RealGitRunner) GetRemotes() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		const minFieldsRequired = 2
		if len(fields) >= minFieldsRequired {
			remoteName := fields[0]
			remoteURL := fields[1]
			// Only add each remote once (git remote -v shows fetch and push separately)
			if _, exists := remotes[remoteName]; !exists {
				remotes[remoteName] = remoteURL
			}
		}
	}

	return remotes, nil
}
