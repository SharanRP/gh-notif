# gh-notif

A high-performance CLI tool for managing GitHub notifications in the terminal.

## Features

- View and filter GitHub notifications
- Mark notifications as read
- Interactive terminal UI
- Concurrent processing for improved performance
- OAuth2 authentication with GitHub

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
```

### Marking Notifications as Read

To mark a notification as read:

```bash
gh-notif read <notification-id>
```

## Configuration

gh-notif uses a configuration file located at `~/.gh-notif.yaml`. You can specify a different configuration file using the `--config` flag.

Example configuration:

```yaml
github_client_id: your_client_id
github_client_secret: your_client_secret
refresh_interval: 60
max_concurrent: 5
debug: false
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
