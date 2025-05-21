# Getting Started with gh-notif

This tutorial will guide you through the process of setting up and using gh-notif, a high-performance CLI tool for managing GitHub notifications.

## Table of Contents

- [Installation](#installation)
- [Authentication](#authentication)
- [Basic Usage](#basic-usage)
- [Filtering Notifications](#filtering-notifications)
- [Grouping Notifications](#grouping-notifications)
- [Searching Notifications](#searching-notifications)
- [Watching Notifications](#watching-notifications)
- [Terminal UI](#terminal-ui)
- [Configuration](#configuration)
- [Next Steps](#next-steps)

## Installation

### Prerequisites

- Go 1.18 or higher
- Git
- A GitHub account

### Installing from Source

```bash
# Clone the repository
git clone https://github.com/user/gh-notif.git
cd gh-notif

# Build and install
go install
```

### Installing from Binary

Download the latest release from the [releases page](https://github.com/user/gh-notif/releases) and extract it to a directory in your PATH.

## Authentication

Before using gh-notif, you need to authenticate with GitHub:

```bash
# Start the authentication process
gh-notif auth login
```

This will start the OAuth device flow:

1. A device code will be displayed in your terminal
2. You'll be prompted to visit a URL in your browser
3. Enter the device code on the GitHub website
4. Grant permission to gh-notif
5. Return to the terminal

Your authentication token will be securely stored using your system's keyring (or in an encrypted file if a keyring is not available).

## Basic Usage

### Listing Notifications

The most basic command is `list`, which shows your unread notifications:

```bash
# List all unread notifications
gh-notif list
```

You can specify various options to filter the results:

```bash
# List all notifications (including read ones)
gh-notif list --all

# List only participating notifications
gh-notif list --participating

# Limit the number of notifications
gh-notif list --limit 10

# Sort notifications
gh-notif list --sort updated
```

### Reading Notifications

To mark a notification as read:

```bash
# Mark a notification as read
gh-notif read <notification-id>

# Mark all notifications as read
gh-notif mark-read

# Mark notifications matching a filter as read
gh-notif mark-read --filter="repo:owner/repo"
```

### Opening Notifications

To open a notification in your browser:

```bash
# Open a notification in the browser
gh-notif open <notification-id>
```

## Filtering Notifications

gh-notif provides powerful filtering capabilities:

### Simple Filtering

```bash
# Filter by repository
gh-notif list --repo="owner/repo"

# Filter by type
gh-notif list --type="PullRequest"

# Filter by reason
gh-notif list --reason="mention"

# Filter by state
gh-notif list --state="unread"

# Filter by time
gh-notif list --since="2023-01-01"
```

### Advanced Filtering

You can use complex filter expressions:

```bash
# Use a complex filter expression
gh-notif list --filter="repo:owner/repo AND type:PullRequest AND is:unread"

# Use OR conditions
gh-notif list --filter="repo:owner/repo OR repo:owner/another-repo"

# Use NOT conditions
gh-notif list --filter="NOT type:Issue"

# Use parentheses for grouping
gh-notif list --filter="(repo:owner/repo OR repo:owner/another-repo) AND is:unread"
```

### Saving Filters

You can save filters for later use:

```bash
# Save a filter
gh-notif filter save my-prs "repo:owner/repo type:PullRequest is:unread"

# Use a saved filter
gh-notif list --filter="@my-prs"

# List saved filters
gh-notif filter list

# Get a saved filter
gh-notif filter get my-prs

# Delete a saved filter
gh-notif filter delete my-prs
```

## Grouping Notifications

You can group notifications by various criteria:

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
```

## Searching Notifications

You can search notifications using full-text search:

```bash
# Search for text
gh-notif search "bug fix"

# Search with regex
gh-notif search "bug.*fix" --regex

# Search in interactive mode
gh-notif search --interactive
```

## Watching Notifications

You can watch for new notifications in real-time:

```bash
# Watch all notifications
gh-notif watch

# Watch with a filter
gh-notif watch --filter="repo:owner/repo"

# Watch with desktop notifications
gh-notif watch --desktop-notification

# Watch with a custom refresh interval
gh-notif watch --interval 30
```

## Terminal UI

gh-notif provides an interactive terminal UI:

```bash
# Start the terminal UI
gh-notif ui

# Start with a filter
gh-notif ui --filter="repo:owner/repo"
```

In the terminal UI, you can:

- Navigate notifications with arrow keys
- Mark notifications as read with `r`
- Open notifications in the browser with `o`
- Filter notifications with `/`
- Group notifications with `g`
- Sort notifications with `s`
- View notification details with `Enter`
- Perform batch actions with `Space` to select and `a` for actions
- Get help with `?`

## Configuration

gh-notif is highly configurable:

```bash
# List all configuration values
gh-notif config list

# Get a configuration value
gh-notif config get display.theme

# Set a configuration value
gh-notif config set display.theme dark

# Edit the configuration file
gh-notif config edit
```

The configuration file is located at `~/.gh-notif.yaml` by default.

## Next Steps

Now that you're familiar with the basics of gh-notif, you can:

- Explore advanced features like notification scoring and smart grouping
- Customize your configuration for optimal workflow
- Set up aliases for common commands
- Integrate gh-notif with your shell for enhanced productivity

For more information, check out the [documentation](https://github.com/user/gh-notif) or run `gh-notif help` to see all available commands.
