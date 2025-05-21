# gh-notif

A high-performance CLI tool for managing GitHub notifications in the terminal with advanced filtering, grouping, and search capabilities.

## Features

- View and filter GitHub notifications
- Mark notifications as read
- Interactive terminal UI
- Concurrent processing for improved performance
- OAuth2 authentication with GitHub

### Advanced Filtering

- **Named Filters**: Save and reuse complex filters
  ```
  gh-notif filter save my-prs "repo:owner/repo type:PullRequest is:unread"
  gh-notif list --filter @my-prs
  ```

- **Boolean Expressions**: Combine filters with AND, OR, NOT operators
  ```
  gh-notif list --filter "repo:owner/repo AND (type:PullRequest OR type:Issue) AND is:unread"
  ```

- **Filter Composition**: Build complex filters from simpler ones
  ```
  gh-notif filter save urgent-prs "repo:owner/repo type:PullRequest is:unread reason:mention" --parent my-prs
  ```

- **Shortcuts**: Use predefined shortcuts for common filters
  ```
  gh-notif list --filter @unread
  gh-notif list --filter @prs
  ```

### Notification Scoring

- **Importance Scoring**: Automatically score notifications based on relevance
  ```
  gh-notif list --sort score
  gh-notif list --min-score 75
  ```

- **Configurable Factors**: Score based on user involvement, age, activity level, etc.
  ```
  gh-notif config set scoring.age_weight 0.5
  gh-notif config set scoring.involvement_weight 0.8
  ```

### Smart Grouping

- **Algorithmic Grouping**: Group related notifications automatically
  ```
  gh-notif group --by repository
  gh-notif group --by owner
  gh-notif group --by type
  gh-notif group --by smart
  ```

- **Hierarchical Views**: View notifications in a hierarchical structure
  ```
  gh-notif group --by repository --secondary-by type
  ```

### Watch Mode

- **Real-time Updates**: Monitor notifications in real-time
  ```
  gh-notif watch
  ```

- **Desktop Notifications**: Get desktop notifications for new items
  ```
  gh-notif watch --desktop-notification
  ```

### Powerful Search

- **Full-text Search**: Search across all notification content
  ```
  gh-notif search "bug fix"
  ```

- **Advanced Query Syntax**: Use complex search queries
  ```
  gh-notif search "bug fix" --regex --case-sensitive
  ```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/gh-notif.git
cd gh-notif

# Build the binary
make build

# Install to your GOPATH
make install
```

### Using Go

```bash
go install github.com/yourusername/gh-notif@latest
```

## Usage

### Authentication

Before using gh-notif, you need to authenticate with GitHub:

```bash
gh-notif auth
```

This will guide you through the authentication process.

### Listing Notifications

To list your notifications:

```bash
# List all unread notifications (default)
gh-notif list

# List all notifications (including read)
gh-notif list --all

# Filter notifications by repository
gh-notif list --repo="username/repository"

# Use advanced filtering
gh-notif list --filter "repo:owner/repo AND is:unread AND type:PullRequest"

# Sort by score
gh-notif list --sort score

# Group notifications
gh-notif list --group-by repository
```

### Managing Filters

To manage your filters:

```bash
# Save a filter
gh-notif filter save my-prs "repo:owner/repo type:PullRequest is:unread"

# List saved filters
gh-notif filter list

# Get a filter
gh-notif filter get my-prs

# Delete a filter
gh-notif filter delete my-prs
```

### Grouping Notifications

To group your notifications:

```bash
# Group by repository
gh-notif group --by repository

# Group with secondary grouping
gh-notif group --by repository --secondary-by type

# Group with a filter
gh-notif group --filter "is:unread" --by repository

# Group in interactive mode
gh-notif group --interactive
```

### Searching Notifications

To search your notifications:

```bash
# Search for text
gh-notif search "bug fix"

# Search with a filter
gh-notif search "bug fix" --filter "repo:owner/repo is:unread"

# Search with regex
gh-notif search "bug.*fix" --regex

# Search in interactive mode
gh-notif search --interactive
```

### Watching Notifications

To watch for new notifications:

```bash
# Watch all notifications
gh-notif watch

# Watch with a filter
gh-notif watch --filter "repo:owner/repo is:unread"

# Watch with desktop notifications
gh-notif watch --desktop-notification

# Watch in interactive mode
gh-notif watch --interactive
```

### Marking Notifications as Read

To mark a notification as read:

```bash
gh-notif read <notification-id>
```

## Configuration

gh-notif uses a configuration file located at `~/.config/gh-notif/config.yaml` by default. You can specify a different configuration file using the `--config` flag.

Example configuration:

```yaml
# GitHub settings
github:
  client_id: your_client_id
  client_secret: your_client_secret
  enterprise_url: https://github.example.com  # Optional, for GitHub Enterprise

# General settings
general:
  refresh_interval: 60
  max_concurrent: 5
  debug: false

# Scoring settings
scoring:
  age_weight: 0.3
  activity_weight: 0.2
  involvement_weight: 0.3
  type_weight: 0.1
  reason_weight: 0.1
  repo_weight: 0.1
  custom_repo_weights:
    owner/important-repo: 0.9

# Watch settings
watch:
  refresh_interval: 30
  max_refresh_interval: 300
  backoff_factor: 1.5
  backoff_threshold: 3
  desktop_notifications: false

# UI settings
ui:
  theme: dark
  highlight_color: "205"
  show_status_bar: true
```

You can manage configuration using the `config` command:

```bash
# Get a configuration value
gh-notif config get scoring.age_weight

# Set a configuration value
gh-notif config set scoring.age_weight 0.5

# List all configuration values
gh-notif config list
```

## Development

### Prerequisites

- Go 1.18 or higher
- Make (optional, for using the Makefile)

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
