# Repository Subscriptions

gh-notif provides powerful repository subscription management with wildcard patterns, activity filtering, and priority levels to help you stay informed about the projects that matter most.

## Overview

Repository subscriptions allow you to:
- **Subscribe to repositories** with flexible patterns
- **Filter by activity types** (issues, PRs, releases, etc.)
- **Set priority levels** for different repositories
- **Use wildcard patterns** for organization-wide subscriptions
- **Manage subscription lists** with import/export capabilities
- **Track subscription activity** and engagement metrics

## Quick Start

```bash
# Subscribe to a repository
gh-notif repo-subscribe owner/repo

# Subscribe with specific activity types
gh-notif repo-subscribe owner/repo --types issues,pulls,releases

# Subscribe to all repositories in an organization
gh-notif repo-subscribe "myorg/*" --priority high

# List your subscriptions
gh-notif subscriptions list

# View subscription activity
gh-notif subscriptions activity
```

## Commands

### Subscribe to Repository

Subscribe to repository notifications with advanced options:

```bash
# Basic subscription
gh-notif repo-subscribe owner/repo

# Subscribe with specific activity types
gh-notif repo-subscribe owner/repo --types issues,pulls,releases

# Set priority level
gh-notif repo-subscribe owner/repo --priority high

# Subscribe with custom filters
gh-notif repo-subscribe owner/repo --filter "author:@me OR mentions:@me"

# Subscribe to organization repositories
gh-notif repo-subscribe "myorg/*" --types releases

# Subscribe with notification preferences
gh-notif repo-subscribe owner/repo --notify-email --notify-web
```

#### Subscribe Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--types` | Activity types to subscribe to | all |
| `--priority` | Priority level (low, normal, high, critical) | normal |
| `--filter` | Custom filter expression | none |
| `--notify-email` | Enable email notifications | false |
| `--notify-web` | Enable web notifications | true |
| `--auto-watch` | Automatically watch new repositories | false |

#### Activity Types

- `issues` - Issue creation, updates, and comments
- `pulls` - Pull request creation, updates, and reviews
- `releases` - New releases and tags
- `commits` - New commits and pushes
- `discussions` - Discussion creation and updates
- `wiki` - Wiki page changes
- `security` - Security advisories and alerts
- `actions` - GitHub Actions workflow runs
- `projects` - Project board updates
- `all` - All activity types

### Unsubscribe from Repository

Remove repository subscriptions:

```bash
# Unsubscribe from specific repository
gh-notif repo-unsubscribe owner/repo

# Unsubscribe from organization pattern
gh-notif repo-unsubscribe "myorg/*"

# Unsubscribe from multiple repositories
gh-notif repo-unsubscribe owner/repo1 owner/repo2 owner/repo3

# Force unsubscribe (skip confirmation)
gh-notif repo-unsubscribe owner/repo --force
```

### Manage Subscriptions

List and manage your repository subscriptions:

```bash
# List all subscriptions
gh-notif subscriptions list

# List with details
gh-notif subscriptions list --detailed

# Filter by priority
gh-notif subscriptions list --priority high

# Filter by activity type
gh-notif subscriptions list --type releases

# Search subscriptions
gh-notif subscriptions list --search "react"

# Export subscriptions
gh-notif subscriptions export subscriptions.yaml

# Import subscriptions
gh-notif subscriptions import subscriptions.yaml
```

#### Subscriptions List Options

| Option | Description | Default |
|--------|-------------|---------|
| `--detailed` | Show detailed information | false |
| `--priority` | Filter by priority level | all |
| `--type` | Filter by activity type | all |
| `--search` | Search subscription names | none |
| `--format` | Output format (table, json, yaml) | table |
| `--sort` | Sort by (name, priority, created) | name |

### Subscription Activity

Monitor activity across your subscriptions:

```bash
# Show recent activity
gh-notif subscriptions activity

# Activity for specific repository
gh-notif subscriptions activity --repo owner/repo

# Activity by type
gh-notif subscriptions activity --type releases

# Activity in last 7 days
gh-notif subscriptions activity --since 7d

# Activity with statistics
gh-notif subscriptions activity --stats
```

#### Activity Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Filter by repository | all |
| `--type` | Filter by activity type | all |
| `--since` | Show activity since (1d, 1w, 1m) | 1d |
| `--limit` | Maximum number of items | 50 |
| `--stats` | Include statistics | false |
| `--format` | Output format | table |

## Wildcard Patterns

Use powerful wildcard patterns for flexible subscriptions:

### Organization Subscriptions

```bash
# Subscribe to all repositories in an organization
gh-notif repo-subscribe "myorg/*"

# Subscribe to specific patterns
gh-notif repo-subscribe "myorg/frontend-*"
gh-notif repo-subscribe "myorg/*-api"

# Subscribe to multiple organizations
gh-notif repo-subscribe "org1/*" "org2/*" "org3/*"
```

### Pattern Examples

| Pattern | Matches | Description |
|---------|---------|-------------|
| `myorg/*` | All repos in myorg | Organization-wide |
| `myorg/web-*` | myorg/web-app, myorg/web-api | Prefix matching |
| `myorg/*-api` | myorg/user-api, myorg/auth-api | Suffix matching |
| `*/react-*` | Any org's React projects | Cross-organization |
| `myorg/project-*-v2` | myorg/project-auth-v2 | Complex patterns |

## Priority Levels

Set priority levels to organize your subscriptions:

### Priority Levels

- **Critical**: Mission-critical repositories requiring immediate attention
- **High**: Important repositories for daily work
- **Normal**: Regular repositories of interest (default)
- **Low**: Repositories for occasional monitoring

### Priority-based Filtering

```bash
# List high-priority subscriptions
gh-notif subscriptions list --priority high

# Show only critical notifications
gh-notif list --subscription-priority critical

# Activity from high-priority repos only
gh-notif subscriptions activity --priority high
```

## Configuration

Configure repository subscriptions in your `~/.gh-notif.yaml`:

```yaml
subscriptions:
  # Default priority for new subscriptions
  default_priority: normal
  
  # Default activity types
  default_types:
    - issues
    - pulls
    - releases
  
  # Auto-subscribe to repositories you contribute to
  auto_subscribe_contributions: true
  
  # Auto-subscribe to repositories you star
  auto_subscribe_starred: false
  
  # Notification preferences
  notifications:
    email: false
    web: true
    desktop: true
  
  # Storage settings
  storage:
    file: ~/.gh-notif-subscriptions.yaml
    encrypt: true
    backup: true
  
  # Predefined subscription groups
  groups:
    work:
      pattern: "mycompany/*"
      priority: high
      types: [issues, pulls, releases]
    
    opensource:
      pattern: ["facebook/react", "vercel/next.js", "microsoft/vscode"]
      priority: normal
      types: [releases, discussions]
    
    security:
      pattern: "*"
      priority: critical
      types: [security]
      filter: "type:security"
```

## Import/Export

Manage subscriptions with import/export functionality:

### Export Subscriptions

```bash
# Export to YAML
gh-notif subscriptions export subscriptions.yaml

# Export to JSON
gh-notif subscriptions export subscriptions.json --format json

# Export specific priorities
gh-notif subscriptions export high-priority.yaml --priority high

# Export with encryption
gh-notif subscriptions export subscriptions.yaml --encrypt
```

### Import Subscriptions

```bash
# Import from file
gh-notif subscriptions import subscriptions.yaml

# Import with merge strategy
gh-notif subscriptions import subscriptions.yaml --merge

# Import and overwrite existing
gh-notif subscriptions import subscriptions.yaml --overwrite

# Import from URL
gh-notif subscriptions import https://example.com/subscriptions.yaml
```

### Subscription File Format

```yaml
version: "1.0"
subscriptions:
  - repository: "facebook/react"
    priority: high
    types: [releases, discussions]
    created: "2024-01-15T10:30:00Z"
    
  - repository: "myorg/*"
    priority: normal
    types: [issues, pulls]
    filter: "author:@me OR mentions:@me"
    created: "2024-01-16T09:15:00Z"
    
  - repository: "security/*"
    priority: critical
    types: [security]
    notifications:
      email: true
      web: true
    created: "2024-01-17T14:20:00Z"
```

## Examples

### Development Team Setup

```bash
# Subscribe to team repositories
gh-notif repo-subscribe "mycompany/frontend-*" --priority high --types issues,pulls
gh-notif repo-subscribe "mycompany/backend-*" --priority high --types issues,pulls,releases
gh-notif repo-subscribe "mycompany/infrastructure-*" --priority critical --types all

# Subscribe to dependencies
gh-notif repo-subscribe "facebook/react" --priority normal --types releases
gh-notif repo-subscribe "vercel/next.js" --priority normal --types releases,discussions
```

### Open Source Maintainer

```bash
# Subscribe to your projects
gh-notif repo-subscribe "myusername/*" --priority critical --types all

# Subscribe to dependencies
gh-notif repo-subscribe "nodejs/node" --priority high --types releases,security
gh-notif repo-subscribe "npm/cli" --priority normal --types releases

# Subscribe to community projects
gh-notif repo-subscribe "awesome-lists/*" --priority low --types releases
```

### Security Monitoring

```bash
# Subscribe to security advisories
gh-notif repo-subscribe "*" --priority critical --types security

# Subscribe to security-focused repositories
gh-notif repo-subscribe "github/advisory-database" --priority high
gh-notif repo-subscribe "OWASP/*" --priority normal --types releases,discussions
```

## Best Practices

### Subscription Management

1. **Use Priority Levels**: Organize subscriptions by importance
2. **Leverage Wildcards**: Efficiently subscribe to organization repositories
3. **Filter Activity Types**: Focus on relevant notifications
4. **Regular Cleanup**: Remove unused subscriptions periodically
5. **Export Backups**: Keep backups of your subscription configuration

### Performance Optimization

1. **Specific Patterns**: Use specific wildcards to reduce noise
2. **Activity Filtering**: Subscribe only to relevant activity types
3. **Priority-based Processing**: Process high-priority notifications first
4. **Batch Operations**: Use bulk subscribe/unsubscribe operations
5. **Cache Management**: Enable caching for frequently accessed data

### Team Collaboration

1. **Shared Configurations**: Use import/export for team setups
2. **Standardized Priorities**: Agree on priority level meanings
3. **Pattern Conventions**: Use consistent wildcard patterns
4. **Documentation**: Document subscription strategies
5. **Regular Reviews**: Periodically review and update subscriptions

Repository subscriptions provide a powerful way to stay informed about the projects and communities that matter most to your work and interests, with the flexibility to scale from individual repositories to entire organizations.
