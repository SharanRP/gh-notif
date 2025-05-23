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
- **GitHub Discussions Monitoring**: Full-featured discussions tracking with analytics and engagement metrics
- **Discussion Analytics**: Trending topics, community insights, and engagement analysis
- **Smart Discussion Search**: Full-text search with indexing and relevance scoring
- **Discussion Thread Visualization**: Interactive terminal UI for viewing discussions and comments
- **Platform-specific Secure Storage**: Secure credential storage for each platform
- **Interactive Tutorial**: Built-in tutorial for learning the tool
- **Setup Wizard**: Guided configuration for optimal settings
- **Progressive Discovery**: Learn features as you need them
- **Comprehensive Documentation**: Detailed help, examples, and man pages

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

### GitHub Discussions Monitoring

- **Comprehensive Discussion Tracking**: Monitor discussions across repositories
  ```
  gh-notif discussions list
  gh-notif discussions list --repo owner/repo
  gh-notif discussions list --category "Q&A" --state open
  ```

- **Discussion Analytics**: Get insights into community engagement
  ```
  gh-notif discussions analytics
  gh-notif discussions analytics --repo owner/repo
  gh-notif discussions trending
  ```

- **Smart Search**: Full-text search across discussions with relevance scoring
  ```
  gh-notif discussions search "bug report"
  gh-notif discussions search "feature request" --category "Ideas"
  ```

- **Interactive Discussion Viewer**: View discussions with threading and markdown rendering
  ```
  gh-notif discussions view owner/repo 123
  ```

- **Unanswered Questions**: Find questions that need attention
  ```
  gh-notif discussions unanswered
  gh-notif discussions unanswered --repo owner/repo
  ```

- **Discussion Filtering**: Advanced filtering by category, state, author, and engagement
  ```
  gh-notif discussions list --category "Q&A" --author username
  gh-notif discussions list --state open --min-upvotes 5
  ```

## Installation

### Quick Install (Recommended)

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/user/gh-notif/main/scripts/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/user/gh-notif/main/scripts/install.ps1 | iex
```

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew install user/tap/gh-notif
```

#### Scoop (Windows)
```powershell
scoop bucket add user https://github.com/user/scoop-bucket
scoop install gh-notif
```

#### Snap (Linux)
```bash
sudo snap install gh-notif
```

#### Flatpak (Linux)
```bash
flatpak install flathub com.github.user.gh-notif
```

#### APT (Debian/Ubuntu)
```bash
# Download and install DEB package
curl -fsSL https://github.com/user/gh-notif/releases/latest/download/gh-notif_amd64.deb -o gh-notif.deb
sudo dpkg -i gh-notif.deb
```

#### YUM/DNF (RHEL/Fedora)
```bash
# Download and install RPM package
curl -fsSL https://github.com/user/gh-notif/releases/latest/download/gh-notif-1.0.0-1.x86_64.rpm -o gh-notif.rpm
sudo rpm -i gh-notif.rpm
```

### Docker

```bash
# Run directly
docker run --rm -it ghcr.io/user/gh-notif:latest --help

# Create alias for easier use
echo 'alias gh-notif="docker run --rm -it -v ~/.gh-notif:/root/.gh-notif ghcr.io/user/gh-notif:latest"' >> ~/.bashrc
```

### Manual Download

Download the latest release for your platform from the [releases page](https://github.com/user/gh-notif/releases/latest).

### From Source

#### Using Go Install
```bash
go install github.com/user/gh-notif@latest
```

#### Build from source
```bash
# Clone the repository
git clone https://github.com/user/gh-notif.git
cd gh-notif

# Build the binary
make build

# Or build manually
go build -o gh-notif
```

### Shell Completions

After installation, enable shell completions:

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

#### PowerShell
```powershell
# For current session
gh-notif completion powershell | Out-String | Invoke-Expression

# For all sessions
gh-notif completion powershell > gh-notif.ps1
# Add ". path\to\gh-notif.ps1" to your PowerShell profile
```

### Prerequisites

- GitHub account for authentication
- For building from source: Go 1.20 or higher

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

### Managing Discussions

To work with GitHub discussions:

```bash
# List discussions from subscribed repositories
gh-notif discussions list

# List discussions from a specific repository
gh-notif discussions list --repo owner/repo

# Filter discussions by category and state
gh-notif discussions list --category "Q&A" --state open

# Search discussions
gh-notif discussions search "bug report"
gh-notif discussions search "feature request" --repo owner/repo

# View a specific discussion with comments
gh-notif discussions view owner/repo 123

# Show trending discussions
gh-notif discussions trending
gh-notif discussions trending --repo owner/repo --limit 10

# Find unanswered questions
gh-notif discussions unanswered
gh-notif discussions unanswered --repo owner/repo

# Generate discussion analytics
gh-notif discussions analytics
gh-notif discussions analytics --repo owner/repo

# Interactive discussion browser
gh-notif discussions list --interactive
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
| `tutorial` | Interactive tutorial for learning gh-notif |
| `wizard` | Setup wizard for configuring gh-notif |

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

### Tutorial Command

The `tutorial` command provides an interactive tutorial for learning gh-notif:

```bash
# Start the interactive tutorial
gh-notif tutorial

# Skip the authentication section
gh-notif tutorial --skip-auth

# Skip the basics section
gh-notif tutorial --skip-basics

# Skip the advanced section
gh-notif tutorial --skip-advanced

# Use a specific theme
gh-notif tutorial --theme light

# Disable colors
gh-notif tutorial --no-color
```

#### Tutorial Command Options

| Option | Description |
|--------|-------------|
| `--interactive` | Run in interactive mode (default: true) |
| `--skip-auth` | Skip the authentication section |
| `--skip-basics` | Skip the basics section |
| `--skip-advanced` | Skip the advanced section |
| `--theme` | Color theme (dark, light, auto) |
| `--no-color` | Disable colors |
| `--width` | Width of the tutorial |
| `--height` | Height of the tutorial |

### Wizard Command

The `wizard` command provides an interactive setup wizard for configuring gh-notif:

```bash
# Start the interactive wizard
gh-notif wizard

# Skip the authentication section
gh-notif wizard --skip-auth

# Skip the display section
gh-notif wizard --skip-display

# Skip the notifications section
gh-notif wizard --skip-notifications

# Skip the advanced section
gh-notif wizard --skip-advanced

# Use a specific theme
gh-notif wizard --theme light

# Disable colors
gh-notif wizard --no-color
```

#### Wizard Command Options

| Option | Description |
|--------|-------------|
| `--interactive` | Run in interactive mode (default: true) |
| `--skip-auth` | Skip the authentication section |
| `--skip-display` | Skip the display section |
| `--skip-notifications` | Skip the notifications section |
| `--skip-advanced` | Skip the advanced section |
| `--theme` | Color theme (dark, light, auto) |
| `--no-color` | Disable colors |
| `--width` | Width of the wizard |
| `--height` | Height of the wizard |
| `--config` | Path to the configuration file |

## License

MIT

## Documentation

gh-notif provides comprehensive documentation to help you get the most out of the tool:

### In-tool Documentation

- **Command Help**: Detailed help for all commands with `--help` flag
- **Examples**: Practical examples for common use cases
- **Interactive Tutorial**: Step-by-step guide with `gh-notif tutorial`
- **Contextual Help**: Keyboard shortcut `?` in the TUI for context-specific help

### Project Documentation

- **README.md**: Overview, installation, and usage instructions
- **CONTRIBUTING.md**: Guidelines for contributors
- **Man Pages**: System-level documentation with `man gh-notif`
- **API Documentation**: Generated from code comments
- **Tutorials**: Step-by-step guides in the `docs/tutorials` directory
- **Examples**: Sample configurations and usage in the `docs/examples` directory

### Visual Documentation

- **Terminal Recordings**: Demonstrations of features
- **Screenshots**: Annotated UI examples
- **Quick Reference**: Visual guide to common commands
- **Diagrams**: Explanations of complex features

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.
