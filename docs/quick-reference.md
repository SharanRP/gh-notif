# gh-notif Quick Reference Guide

This quick reference guide provides a concise overview of the most commonly used commands and features in gh-notif.

## Authentication

```bash
# Login to GitHub
gh-notif auth login

# Check authentication status
gh-notif auth status

# Refresh authentication token
gh-notif auth refresh

# Logout from GitHub
gh-notif auth logout
```

## Listing Notifications

```bash
# List unread notifications (default)
gh-notif list

# List all notifications
gh-notif list --all

# List notifications for a specific repository
gh-notif list --repo="owner/repo"

# List notifications of a specific type
gh-notif list --type="PullRequest"

# List notifications with a specific reason
gh-notif list --reason="mention"

# List notifications since a specific date
gh-notif list --since="2023-01-01"

# List notifications with a limit
gh-notif list --limit=10

# List notifications with sorting
gh-notif list --sort="updated"

# List notifications with a complex filter
gh-notif list --filter="repo:owner/repo AND type:PullRequest AND is:unread"
```

## Managing Notifications

```bash
# Mark a notification as read
gh-notif read <notification-id>

# Mark all notifications as read
gh-notif mark-read

# Mark notifications matching a filter as read
gh-notif mark-read --filter="repo:owner/repo"

# Open a notification in the browser
gh-notif open <notification-id>

# Archive a notification
gh-notif archive <notification-id>

# Unarchive a notification
gh-notif unarchive <notification-id>

# Subscribe to a notification thread
gh-notif subscribe <notification-id>

# Unsubscribe from a notification thread
gh-notif unsubscribe <notification-id>

# Mute a repository
gh-notif mute <owner/repo>

# Unmute a repository
gh-notif unmute <owner/repo>
```

## Filtering Notifications

```bash
# Save a filter
gh-notif filter save my-prs "repo:owner/repo type:PullRequest is:unread"

# List saved filters
gh-notif filter list

# Get a filter
gh-notif filter get my-prs

# Delete a filter
gh-notif filter delete my-prs

# Use a saved filter
gh-notif list --filter="@my-prs"
```

## Grouping Notifications

```bash
# Group by repository
gh-notif group --by repository

# Group by type
gh-notif group --by type

# Group by reason
gh-notif group --by reason

# Group by author
gh-notif group --by author

# Group with smart grouping
gh-notif group --by smart

# Group with secondary grouping
gh-notif group --by repository --secondary-by type

# Group with a filter
gh-notif group --filter "is:unread" --by repository
```

## Searching Notifications

```bash
# Search for text
gh-notif search "bug fix"

# Search with a filter
gh-notif search "bug fix" --filter "repo:owner/repo is:unread"

# Search with regex
gh-notif search "bug.*fix" --regex

# Search with case sensitivity
gh-notif search "Bug" --case-sensitive

# Search in interactive mode
gh-notif search --interactive
```

## Watching Notifications

```bash
# Watch all notifications
gh-notif watch

# Watch with a filter
gh-notif watch --filter "repo:owner/repo is:unread"

# Watch with desktop notifications
gh-notif watch --desktop-notification

# Watch with a custom refresh interval
gh-notif watch --interval 30

# Watch with a timeout
gh-notif watch --timeout 3600

# Watch in interactive mode
gh-notif watch --interactive
```

## Terminal UI

```bash
# Start the terminal UI
gh-notif ui

# Start with a filter
gh-notif ui --filter="repo:owner/repo"

# Start with a specific view mode
gh-notif ui --view split

# Start with a specific theme
gh-notif ui --theme light

# Start with high contrast mode
gh-notif ui --high-contrast

# Start with large text mode
gh-notif ui --large-text
```

## GitHub Discussions

```bash
# List discussions
gh-notif discussions list

# List from specific repository
gh-notif discussions list --repo owner/repo

# Search discussions
gh-notif discussions search "bug report"

# Show analytics
gh-notif discussions analytics

# Show trending discussions
gh-notif discussions trending

# Find unanswered questions
gh-notif discussions unanswered

# Interactive mode
gh-notif discussions list --interactive
```

## Repository Subscriptions

```bash
# Subscribe to repository
gh-notif repo-subscribe owner/repo

# Subscribe with activity types
gh-notif repo-subscribe owner/repo --types issues,pulls

# Subscribe to organization
gh-notif repo-subscribe "myorg/*"

# Unsubscribe
gh-notif repo-unsubscribe owner/repo

# List subscriptions
gh-notif subscriptions list

# View activity
gh-notif subscriptions activity

# Export/import
gh-notif subscriptions export subs.yaml
gh-notif subscriptions import subs.yaml
```

## Additional Commands

```bash
# Archive notifications
gh-notif archive --older-than 7d

# Undo last action
gh-notif undo

# Batch actions
gh-notif actions mark-read --filter "repo:owner/repo"

# Generate man pages
gh-notif man generate

# Install completions
gh-notif completion install

# Show version
gh-notif version

# First run setup
gh-notif firstrun
```

## Configuration Management

```bash
# Get a configuration value
gh-notif config get auth.client_id

# Set a configuration value
gh-notif config set scoring.age_weight 0.5

# List all configuration values
gh-notif config list

# Edit the configuration file
gh-notif config edit

# Export configuration to a file
gh-notif config export config-backup.yaml

# Import configuration from a file
gh-notif config import config-backup.yaml
```

## Learning and Setup

```bash
# Run the interactive tutorial
gh-notif tutorial

# Skip specific sections of the tutorial
gh-notif tutorial --skip-auth --skip-basics

# Run the setup wizard
gh-notif wizard

# Skip specific sections of the wizard
gh-notif wizard --skip-auth --skip-display
```

## Profiling and Benchmarking

```bash
# Run basic profiling
gh-notif profile

# Enable CPU profiling
gh-notif profile --cpu

# Enable memory profiling
gh-notif profile --memory

# Set profiling duration
gh-notif profile --duration 60

# Enable HTTP profiling server
gh-notif profile --http

# Run benchmarks
gh-notif profile --benchmark-runs 10 --benchmark-size 200
```

## Common Flags

These flags can be used with most commands:

| Flag | Description |
|------|-------------|
| `--config` | Specify the configuration file |
| `--debug` | Enable debug output |
| `--no-color` | Disable color output |
| `--help` | Show help for a command |
| `--version` | Show version information |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GH_NOTIF_AUTH_TOKEN` | GitHub authentication token |
| `GH_NOTIF_CONFIG` | Path to configuration file |
| `GH_NOTIF_CACHE_DIR` | Path to cache directory |
| `GH_NOTIF_DEBUG` | Enable debug mode |
| `GH_NOTIF_NO_COLOR` | Disable color output |

## Common Filter Expressions

| Expression | Description |
|------------|-------------|
| `is:unread` | Unread notifications |
| `is:read` | Read notifications |
| `repo:owner/repo` | Notifications from a specific repository |
| `org:organization` | Notifications from a specific organization |
| `type:PullRequest` | Pull request notifications |
| `type:Issue` | Issue notifications |
| `reason:mention` | Notifications where you were mentioned |
| `reason:assign` | Notifications where you were assigned |
| `reason:review` | Notifications for review requests |
| `author:username` | Notifications from a specific author |
| `involves:username` | Notifications involving a specific user |
| `label:bug` | Notifications with a specific label |
| `state:open` | Notifications for open issues/PRs |
| `state:closed` | Notifications for closed issues/PRs |
| `created:>2023-01-01` | Notifications created after a date |
| `updated:<2023-01-01` | Notifications updated before a date |

## Combining Filters

| Operator | Description | Example |
|----------|-------------|---------|
| `AND` | Both conditions must be true | `repo:owner/repo AND type:PullRequest` |
| `OR` | Either condition can be true | `type:PullRequest OR type:Issue` |
| `NOT` | Negates a condition | `NOT type:Issue` |
| `()` | Groups conditions | `(type:PullRequest OR type:Issue) AND is:unread` |

## Keyboard Shortcuts (Terminal UI)

| Shortcut | Description |
|----------|-------------|
| `?` | Show help |
| `q` | Quit |
| `r` | Mark as read |
| `u` | Mark as unread |
| `o` | Open in browser |
| `a` | Archive |
| `s` | Subscribe |
| `S` | Unsubscribe |
| `f` | Filter |
| `g` | Group |
| `t` | Sort |
| `v` | Change view mode |
| `/` | Search |
| `Space` | Select notification |
| `Enter` | View details |
| `Esc` | Go back |
| `j`, `↓` | Next notification |
| `k`, `↑` | Previous notification |
| `h`, `←` | Go back |
| `l`, `→` | Go forward |
| `1`-`4` | Switch view modes |
| `Ctrl+R` | Refresh |
| `Ctrl+F` | Search |
| `Ctrl+S` | Save filter |
| `Ctrl+Z` | Undo |
