# Additional Commands

This document covers additional commands and features in gh-notif that provide enhanced functionality for notification management, system utilities, and developer tools.

## Archive Command

Archive notifications to keep your inbox clean while preserving access to important notifications.

### Usage

```bash
# Archive a specific notification
gh-notif archive <notification-id>

# Archive multiple notifications
gh-notif archive <id1> <id2> <id3>

# Archive all read notifications
gh-notif archive --read

# Archive notifications older than 30 days
gh-notif archive --older-than 30d

# Archive by repository
gh-notif archive --repo owner/repo

# Archive with confirmation
gh-notif archive <notification-id> --confirm

# Archive and show details
gh-notif archive <notification-id> --verbose
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--read` | Archive all read notifications | false |
| `--older-than` | Archive notifications older than duration | none |
| `--repo` | Archive notifications from specific repository | all |
| `--confirm` | Require confirmation before archiving | false |
| `--verbose` | Show detailed information | false |
| `--dry-run` | Show what would be archived without doing it | false |

## Undo Command

Undo the last action performed, providing a safety net for accidental operations.

### Usage

```bash
# Undo the last action
gh-notif undo

# Show what can be undone
gh-notif undo --show

# Undo specific action by ID
gh-notif undo --action-id <id>

# Clear undo history
gh-notif undo --clear

# Show undo history
gh-notif undo --history
```

### Supported Actions

- Mark as read/unread
- Archive/unarchive
- Subscribe/unsubscribe
- Mute/unmute
- Filter operations
- Bulk actions

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--show` | Show what can be undone | false |
| `--action-id` | Undo specific action | last action |
| `--clear` | Clear undo history | false |
| `--history` | Show undo history | false |
| `--force` | Force undo without confirmation | false |

## Actions Command

Perform batch actions on multiple notifications efficiently.

### Usage

```bash
# Mark multiple notifications as read
gh-notif actions mark-read --filter "repo:owner/repo"

# Archive notifications in bulk
gh-notif actions archive --older-than 7d

# Subscribe to multiple repositories
gh-notif actions subscribe --repos "org1/repo1,org2/repo2"

# Batch mute repositories
gh-notif actions mute --pattern "spam-org/*"

# Interactive batch actions
gh-notif actions --interactive

# Dry run to preview actions
gh-notif actions mark-read --filter "is:read" --dry-run
```

### Action Types

- `mark-read` - Mark notifications as read
- `mark-unread` - Mark notifications as unread
- `archive` - Archive notifications
- `subscribe` - Subscribe to repositories
- `unsubscribe` - Unsubscribe from repositories
- `mute` - Mute repositories
- `unmute` - Unmute repositories

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--filter` | Filter expression for target notifications | none |
| `--repos` | Comma-separated list of repositories | none |
| `--pattern` | Pattern for repository matching | none |
| `--older-than` | Target notifications older than duration | none |
| `--interactive` | Interactive mode for action selection | false |
| `--dry-run` | Preview actions without executing | false |
| `--confirm` | Require confirmation for each action | false |
| `--parallel` | Number of parallel operations | 5 |

## Man Pages Command

Generate and install manual pages for system-level documentation.

### Usage

```bash
# Generate man pages
gh-notif man generate

# Install man pages to system
gh-notif man install

# Install to custom directory
gh-notif man install --dir /usr/local/share/man

# Generate specific man page
gh-notif man generate --command list

# View man page
gh-notif man view gh-notif

# Remove installed man pages
gh-notif man uninstall
```

### Generated Man Pages

- `gh-notif(1)` - Main command overview
- `gh-notif-list(1)` - List command details
- `gh-notif-auth(1)` - Authentication commands
- `gh-notif-config(1)` - Configuration management
- `gh-notif-discussions(1)` - Discussions commands
- `gh-notif-ui(1)` - Terminal UI guide

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--dir` | Installation directory | `/usr/local/share/man` |
| `--command` | Generate specific command man page | all |
| `--format` | Output format (man, html, pdf) | man |
| `--force` | Overwrite existing files | false |

## Completion Command

Generate shell completion scripts for enhanced command-line experience.

### Usage

```bash
# Generate bash completion
gh-notif completion bash

# Generate zsh completion
gh-notif completion zsh

# Generate fish completion
gh-notif completion fish

# Generate PowerShell completion
gh-notif completion powershell

# Install completion for current shell
gh-notif completion install

# Install for specific shell
gh-notif completion install --shell bash
```

### Installation Examples

#### Bash
```bash
# For current session
source <(gh-notif completion bash)

# For all sessions
gh-notif completion bash | sudo tee /etc/bash_completion.d/gh-notif
```

#### Zsh
```bash
# For current session
source <(gh-notif completion zsh)

# For all sessions
gh-notif completion zsh > "${fpath[1]}/_gh-notif"
```

#### Fish
```bash
# For current session
gh-notif completion fish | source

# For all sessions
gh-notif completion fish > ~/.config/fish/completions/gh-notif.fish
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--shell` | Target shell (bash, zsh, fish, powershell) | auto-detect |
| `--install` | Install completion script | false |
| `--global` | Install globally for all users | false |

## Version Command

Display version information and build details.

### Usage

```bash
# Show version information
gh-notif version

# Show detailed build information
gh-notif version --detailed

# Show version in JSON format
gh-notif version --json

# Check for updates
gh-notif version --check-updates

# Show changelog
gh-notif version --changelog
```

### Output Information

- Version number
- Git commit hash
- Build date
- Go version used
- Platform and architecture
- Build flags and tags

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--detailed` | Show detailed build information | false |
| `--json` | Output in JSON format | false |
| `--check-updates` | Check for available updates | false |
| `--changelog` | Show changelog for current version | false |

## First Run Command

Interactive first-run experience for new users.

### Usage

```bash
# Start first-run setup
gh-notif firstrun

# Skip specific sections
gh-notif firstrun --skip-auth --skip-tutorial

# Run in non-interactive mode
gh-notif firstrun --non-interactive

# Reset first-run status
gh-notif firstrun --reset
```

### First Run Process

1. **Welcome Screen**: Introduction to gh-notif
2. **Authentication Setup**: GitHub OAuth configuration
3. **Basic Configuration**: Essential settings
4. **Repository Selection**: Choose repositories to monitor
5. **Tutorial Offer**: Optional interactive tutorial
6. **Completion**: Summary and next steps

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--skip-auth` | Skip authentication setup | false |
| `--skip-config` | Skip configuration setup | false |
| `--skip-tutorial` | Skip tutorial offer | false |
| `--non-interactive` | Run without user interaction | false |
| `--reset` | Reset first-run status | false |

## Integration Examples

### Automation Scripts

```bash
#!/bin/bash
# Daily notification cleanup

# Archive old read notifications
gh-notif archive --read --older-than 7d

# Mark low-priority notifications as read
gh-notif actions mark-read --filter "priority:low AND is:unread"

# Generate daily report
gh-notif list --format json > daily-notifications.json
```

### CI/CD Integration

```yaml
# GitHub Actions workflow
- name: Check notifications
  run: |
    gh-notif auth login --token ${{ secrets.GITHUB_TOKEN }}
    gh-notif list --repo ${{ github.repository }} --format json
```

### Shell Aliases

```bash
# Useful aliases
alias gn='gh-notif'
alias gnl='gh-notif list'
alias gns='gh-notif search'
alias gnd='gh-notif discussions list'
alias gnu='gh-notif undo'
```

## Best Practices

### Batch Operations

1. **Use Dry Run**: Always test batch operations with `--dry-run`
2. **Confirm Actions**: Use `--confirm` for destructive operations
3. **Parallel Processing**: Adjust `--parallel` based on system capabilities
4. **Filter Carefully**: Use specific filters to avoid unintended actions

### System Integration

1. **Install Completions**: Set up shell completions for better UX
2. **Install Man Pages**: Install man pages for offline documentation
3. **Regular Updates**: Check for updates regularly
4. **Backup Configuration**: Export configuration before major changes

### Automation

1. **Script Common Tasks**: Automate repetitive notification management
2. **Use JSON Output**: Parse JSON output for integration with other tools
3. **Error Handling**: Implement proper error handling in scripts
4. **Logging**: Log automation activities for debugging

These additional commands provide powerful utilities for managing notifications, system integration, and developer productivity, making gh-notif a comprehensive solution for GitHub notification management.
