# gh-notif

A high-performance CLI tool for managing GitHub notifications in the terminal with advanced filtering, grouping, and search capabilities.

[![Go Report Card](https://goreportcard.com/badge/github.com/user/gh-notif)](https://goreportcard.com/report/github.com/user/gh-notif)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`gh-notif` helps you efficiently manage GitHub notifications with powerful filtering, sorting, and grouping capabilities. It provides a modern terminal UI and supports concurrent operations for improved performance.

## Features

- **Secure Authentication**: OAuth2 device flow with secure token storage
- **Advanced Filtering**: Complex filters with boolean expressions and named filters
- **Smart Grouping**: Group notifications by repository, owner, type, or using smart algorithms
- **Notification Scoring**: Automatically prioritize important notifications
- **High Performance**: Concurrent operations with optimized caching for lightning-fast response
- **Modern Terminal UI**: Interactive interface with responsive layouts
- **Powerful Search**: Full-text search across all notification content
- **Watch Mode**: Real-time updates with desktop notifications
- **Comprehensive Actions**: Mark as read, archive, subscribe/unsubscribe with batch operations
- **Platform-specific Secure Storage**: Secure credential storage for each platform

### Performance Optimizations

- **Persistent Caching**: BadgerDB/BoltDB-backed persistent cache with smart invalidation
- **Conditional Requests**: ETag/If-Modified-Since support to minimize API usage
- **Concurrent Operations**: Optimized worker pools with intelligent backpressure
- **Memory Efficiency**: Object pooling and streaming responses for reduced memory usage
- **Request Batching**: Smart batching of API requests for optimal throughput
- **Background Prefetching**: Predictive loading of likely-needed data
- **Profiling Tools**: Built-in profiling and benchmarking capabilities

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
git clone https://github.com/user/gh-notif.git
cd gh-notif

# Build the binary
go build -o gh-notif

# Install to your GOPATH (optional)
go install
```

### Using Go Install

```bash
go install github.com/user/gh-notif@latest
```

### Windows

```powershell
# Clone the repository
git clone https://github.com/user/gh-notif.git
cd gh-notif

# Build the binary
go build -o gh-notif.exe
```

### Prerequisites

- Go 1.18 or higher
- GitHub account with personal access token (for authentication)

## Usage

### Authentication

Before using gh-notif, you need to authenticate with GitHub:

```bash
# Start the authentication process
gh-notif auth login

# Check authentication status
gh-notif auth status

# Refresh your token if it has expired
gh-notif auth refresh

# Log out and remove stored credentials
gh-notif auth logout
```

The login command will guide you through the OAuth2 device flow:
1. A device code will be displayed
2. A browser will open automatically to GitHub's device activation page
3. Enter the code on the GitHub page
4. Authorize the application
5. The token will be securely stored on your system

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

gh-notif uses a configuration file located at `~/.gh-notif.yaml` by default on Unix systems and `%USERPROFILE%\.gh-notif.yaml` on Windows. You can specify a different configuration file using the `--config` flag.

### Configuration Management

You can manage configuration using the `config` command:

```bash
# Get a configuration value
gh-notif config get auth.client_id

# Set a configuration value
gh-notif config set scoring.age_weight 0.5

# List all configuration values
gh-notif config list

# Edit the configuration file in your default editor
gh-notif config edit

# Export configuration to a file
gh-notif config export config-backup.yaml

# Import configuration from a file
gh-notif config import config-backup.yaml
```

### Example Configuration

```yaml
# API settings
api:
  base_url: https://api.github.com
  retry_count: 3
  retry_delay: 1
  timeout: 30
  upload_url: https://uploads.github.com

# Authentication settings
auth:
  client_id: your_client_id
  client_secret: your_client_secret
  scopes:
    - notifications
    - repo
    - user
  token_storage: file  # Options: file, keyring, auto

# Display settings
display:
  compact_mode: false
  date_format: relative  # Options: relative, absolute, iso
  output_format: table   # Options: table, json, yaml, csv
  show_emojis: true
  theme: dark            # Options: dark, light

# Advanced settings
advanced:
  cache_dir: ""          # Default: ~/.gh-notif-cache
  cache_ttl: 3600        # Cache time-to-live in seconds
  cache_type: "badger"   # Options: memory, badger, bolt, null
  cache_max_size: 1073741824  # 1GB max cache size
  cache_memory_limit: 104857600  # 100MB memory limit
  debug: false
  editor: notepad        # Default editor for config edit
  max_concurrent: 5      # Maximum concurrent operations
  batch_size: 5          # Size of batches for concurrent requests
  use_etag: true         # Use ETags for conditional requests
  background_refresh: true  # Refresh cache in the background
  prefetch_concurrency: 2  # Number of concurrent prefetch operations
  stream_response: false  # Stream API responses

# Notification settings
notifications:
  auto_refresh: false
  default_filter: unread
  exclude_orgs: []       # Organizations to exclude
  exclude_repos: []      # Repositories to exclude
  exclude_types: []      # Notification types to exclude
  include_orgs: []       # Organizations to include (if empty, include all)
  include_repos: []      # Repositories to include (if empty, include all)
  include_types: []      # Notification types to include (if empty, include all)
  refresh_interval: 60   # Refresh interval in seconds
```

## Development

### Prerequisites

- Go 1.18 or higher
- Git

### Building

```bash
# Build for your current platform
go build -o gh-notif

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o gh-notif.exe

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o gh-notif-macos

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o gh-notif-linux
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Project Structure

```
gh-notif/
├── cmd/                  # Command-line interface
│   └── gh-notif/         # Main command and subcommands
├── internal/             # Internal packages
│   ├── auth/             # Authentication
│   ├── config/           # Configuration management
│   ├── filter/           # Notification filtering
│   ├── github/           # GitHub API client
│   ├── grouping/         # Notification grouping
│   ├── output/           # Output formatting
│   ├── scoring/          # Notification scoring
│   ├── search/           # Search functionality
│   ├── ui/               # Terminal UI
│   └── watch/            # Watch mode
├── main.go               # Application entry point
└── README.md             # Documentation
```

## Command Reference

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with GitHub |
| `auth status` | Check authentication status |
| `auth logout` | Log out from GitHub |
| `auth refresh` | Refresh authentication token |
| `config get` | Get a configuration value |
| `config set` | Set a configuration value |
| `config list` | List all configuration values |
| `config edit` | Edit the configuration file |
| `config export` | Export configuration to a file |
| `config import` | Import configuration from a file |
| `list` | List notifications with filtering |
| `read` | Mark a notification as read |
| `mark-read` | Mark multiple notifications as read |
| `open` | Open a notification in the browser |
| `group` | Group notifications |
| `search` | Search notifications |
| `watch` | Watch for new notifications |
| `ui` | Interactive terminal UI |
| `filter save` | Save a filter |
| `filter list` | List saved filters |
| `filter get` | Get a filter |
| `filter delete` | Delete a filter |
| `subscribe` | Subscribe to notification threads |
| `unsubscribe` | Unsubscribe from notification threads |
| `mute` | Mute notifications from repositories |
| `unmute` | Unmute notifications from repositories |
| `profile` | Profile and benchmark the application |

### UI Command

The `ui` command provides an interactive terminal interface for managing notifications:

```bash
# Show all notifications in the UI
gh-notif ui

# Show only unread notifications
gh-notif ui --unread

# Show notifications for a specific repository
gh-notif ui --repo owner/repo

# Use a specific view mode
gh-notif ui --view split

# Use a specific color scheme
gh-notif ui --theme light

# Enable high contrast mode for accessibility
gh-notif ui --high-contrast
```

#### UI Command Options

| Option | Description |
|--------|-------------|
| `--all`, `-a` | Show all notifications, including read ones |
| `--unread`, `-u` | Show only unread notifications (default: true) |
| `--repo`, `-r` | Show notifications for a specific repository |
| `--org`, `-o` | Show notifications for a specific organization |
| `--view`, `-v` | Initial view mode (compact, detailed, split, table) |
| `--theme`, `-t` | Color theme (dark, light, high-contrast) |
| `--high-contrast` | Enable high contrast mode |
| `--large-text` | Enable large text mode |
| `--screen-reader` | Optimize for screen readers |
| `--no-unicode` | Use ASCII characters instead of Unicode |
| `--no-animations` | Disable animations |

### Profile Command

The `profile` command provides tools for profiling and benchmarking the application:

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

#### Profile Command Options

| Option | Description |
|--------|-------------|
| `--cpu` | Enable CPU profiling (default: true) |
| `--memory` | Enable memory profiling (default: true) |
| `--duration` | Duration of profiling in seconds (default: 30) |
| `--http` | Enable HTTP profiling server |
| `--http-addr` | HTTP profiling server address (default: localhost:6060) |
| `--benchmark-runs` | Number of benchmark runs (default: 5) |
| `--benchmark-size` | Size of benchmark data (default: 100) |

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
