# Development Environment Setup

## Prerequisites

This project uses [mise](https://mise.jdx.dev/) for managing development tool versions.

### Install mise

```bash
# macOS
brew install mise

# Linux & Windows (WSL)
curl https://mise.run | sh
```

### Activate mise in your shell

Add this to your shell configuration (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
eval "$(mise activate bash)"  # or zsh, fish, etc.
```

Or for the current session:

```bash
eval "$(mise activate bash --shims)"
```

## Quick Start

1. Install project tools:
   
   ```bash
   mise install
   ```
   
   This will install:

   - Go 1.25.4
   - golangci-lint 2.6.2
   - task 3.45.5
   - markdownlint-cli 0.46.0

2. Verify installation:

   ```bash
   mise list
   ```

3. Run development tasks:

   ```bash
   task          # Show available tasks
   task build    # Build the binary
   task test     # Run tests
   task lint     # Run linters
   ```

## Tool Versions

Tool versions are pinned in `.mise.toml`:

```toml
[tools]
go = "1.25.4"
golangci-lint = "2.6.2"
task = "3.45.5"
"npm:markdownlint-cli" = "0.46.0"
```

To update versions:

1. Check available versions:

   ```bash
   mise ls-remote go
   mise ls-remote golangci-lint
   mise ls-remote task
   npm view markdownlint-cli version  # Check latest npm package version
   ```

2. Update `.mise.toml` with new versions
3. Run `mise install` to apply changes

## Benefits of mise

- **Version pinning**: Ensures all developers use the same tool versions
- **Automatic activation**: Tools are available when you `cd` into the project
- **Per-project isolation**: Different projects can use different versions
- **Fast**: Tools are downloaded and cached locally
- **No system pollution**: Tools are installed in `~/.local/share/mise`

## Without mise

If you prefer not to use mise, you can install tools manually:

```bash
# Install Go 1.25.4
# See https://go.dev/dl/

# Install golangci-lint 2.6.2
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.6.2

# Install task 3.45.5
go install github.com/go-task/task/v3/cmd/task@v3.45.5

# Install markdownlint-cli 0.46.0
npm install -g markdownlint-cli@0.46.0
```

Then update `Taskfile.yaml` to remove `mise exec --` prefixes from commands.
