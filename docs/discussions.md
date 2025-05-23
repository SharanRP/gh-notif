# GitHub Discussions Management

gh-notif provides comprehensive GitHub Discussions monitoring and management capabilities with advanced analytics, search, and interactive features.

## Overview

The discussions feature allows you to:
- **Monitor discussions** across multiple repositories
- **Search and filter** discussions with advanced criteria
- **View detailed analytics** and community insights
- **Track trending topics** and engagement metrics
- **Find unanswered questions** that need attention
- **Interactive browsing** with enhanced terminal UI

## Quick Start

```bash
# List recent discussions
gh-notif discussions list

# List discussions from a specific repository
gh-notif discussions list --repo vercel/next.js

# Search for discussions about bugs
gh-notif discussions search "bug report"

# Show analytics for a repository
gh-notif discussions analytics --repo owner/repo
```

## Commands

### List Discussions

List discussions with powerful filtering and sorting options:

```bash
# Basic listing
gh-notif discussions list

# Repository-specific discussions
gh-notif discussions list --repo owner/repo

# Filter by state
gh-notif discussions list --state open
gh-notif discussions list --state closed

# Filter by author
gh-notif discussions list --author username

# Limit results
gh-notif discussions list --limit 20

# Sort by different criteria
gh-notif discussions list --sort created --direction asc
gh-notif discussions list --sort updated --direction desc

# Interactive mode with enhanced UI
gh-notif discussions list --interactive

# JSON output for scripting
gh-notif discussions list --format json

# Enhanced table format
gh-notif discussions list --format table
```

#### List Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository to filter by (owner/repo) | All subscribed |
| `--category` | Discussion category ID to filter by | All categories |
| `--state` | Discussion state (open, closed, all) | All states |
| `--author` | Filter by discussion author | All authors |
| `--limit` | Maximum number of results | 50 |
| `--sort` | Sort by (created, updated, comments) | updated |
| `--direction` | Sort direction (asc, desc) | desc |
| `--format` | Output format (table, json, csv) | table |
| `--interactive` | Use enhanced interactive mode | false |

### Search Discussions

Powerful full-text search across discussions:

```bash
# Basic search
gh-notif discussions search "bug report"

# Search in specific repository
gh-notif discussions search "feature request" --repo owner/repo

# Search with filters
gh-notif discussions search "help" --state open --limit 10

# Search with sorting
gh-notif discussions search "performance" --sort created --direction desc
```

#### Search Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository to search in | All subscribed |
| `--category` | Category to search in | All categories |
| `--state` | State to search in | All states |
| `--author` | Author to search for | All authors |
| `--limit` | Maximum results | 50 |
| `--sort` | Sort results by | updated |
| `--direction` | Sort direction | desc |
| `--format` | Output format | table |

### View Discussion

View a specific discussion with comments and details:

```bash
# View discussion by repository and number
gh-notif discussions view owner/repo 123

# Include comments in the view
gh-notif discussions view owner/repo 123 --include-comments

# Interactive viewer with enhanced UI
gh-notif discussions view owner/repo 123 --interactive
```

*Note: The view command is currently under development for full implementation.*

### Analytics

Get comprehensive analytics and insights:

```bash
# Repository analytics
gh-notif discussions analytics --repo owner/repo

# Global analytics across all subscribed repositories
gh-notif discussions analytics

# Analytics with custom time range
gh-notif discussions analytics --repo owner/repo --days 30
```

#### Analytics Features

- **Engagement Metrics**: Total discussions, open/closed counts, answered discussions
- **Average Statistics**: Comments, reactions, upvotes per discussion
- **Top Authors**: Most active community members
- **Trending Topics**: Popular keywords and topics
- **Time-based Analysis**: Activity over time periods

### Trending Discussions

Find discussions with high engagement:

```bash
# Show trending discussions
gh-notif discussions trending

# Trending in specific repository
gh-notif discussions trending --repo owner/repo

# Limit trending results
gh-notif discussions trending --limit 10
```

### Unanswered Questions

Find questions that need attention:

```bash
# Show unanswered questions
gh-notif discussions unanswered

# Unanswered in specific repository
gh-notif discussions unanswered --repo owner/repo

# Filter by age
gh-notif discussions unanswered --older-than 7d
```

## Enhanced UI Features

### Beautiful Visual Design

- **Status Indicators**: 🟢 OPEN, ✅ ANSWERED with color coding
- **Rich Emojis**: 📝 titles, 📁 repositories, 📂 categories, 👤 authors
- **Engagement Metrics**: 👍 upvotes, 💬 comments, 🔥 reactions
- **Smart Time Display**: Relative times (2h ago, 3d ago, etc.)
- **Category Display**: Emoji and name combinations

### Interactive Features

- **Enhanced Table Format**: Clean, structured display with borders
- **Interactive Mode**: Box-style layout with enhanced navigation
- **JSON Export**: Complete data structures for automation
- **Real-time Updates**: Live refresh capabilities
- **Keyboard Navigation**: Full keyboard support

### Output Formats

#### Table Format (Default)
```
🗣️  Found 3 discussions:

▶ 🟢 OPEN #55272
  📝 Docs: Bun runtime support
  📁 vercel/next.js • 📂 :bulb: Ideas
  👤 @zigang93 • 🕒 Sep 11, 2023
  👍 105 • 💬 15 • 🔥 133
  🔗 https://github.com/vercel/next.js/discussions/55272
```

#### Interactive Format
```
🗣️  Enhanced Discussion List (3 discussions)

┌─ 🟢 OPEN #55272
│  📝 Docs: Bun runtime support
│  📁 vercel/next.js • 📂 :bulb: Ideas
│  👤 @zigang93 • 🕒 Sep 11, 2023
│  👍 105 • 💬 15 • 🔥 133
│  🔗 https://github.com/vercel/next.js/discussions/55272
└─
```

#### JSON Format
Complete structured data for automation and integration.

## Configuration

Configure discussions monitoring in your `~/.gh-notif.yaml`:

```yaml
discussions:
  # Default repository for discussions commands
  default_repo: ""
  
  # Default number of results to show
  default_limit: 50
  
  # Default sort order
  default_sort: "updated"
  default_direction: "desc"
  
  # Default output format
  default_format: "table"
  
  # Enable interactive mode by default
  interactive_mode: false
  
  # Cache settings
  cache_ttl: 300  # 5 minutes
  use_cache: true
  
  # Repositories to monitor
  subscribed_repos:
    - "vercel/next.js"
    - "facebook/react"
    - "microsoft/vscode"
  
  # Categories to focus on
  focus_categories:
    - "Q&A"
    - "Help"
    - "Ideas"
  
  # Notification settings
  notifications:
    new_discussions: true
    answered_questions: true
    trending_topics: true
```

## Examples

### Daily Workflow

```bash
# Morning: Check new discussions
gh-notif discussions list --limit 10

# Search for specific topics
gh-notif discussions search "performance issues"

# Check unanswered questions in your repositories
gh-notif discussions unanswered --repo myorg/myrepo

# Weekly: Review analytics
gh-notif discussions analytics --repo myorg/myrepo
```

### Community Management

```bash
# Find questions needing attention
gh-notif discussions unanswered --older-than 3d

# Monitor trending topics
gh-notif discussions trending --limit 5

# Track specific categories
gh-notif discussions list --category "Q&A" --state open

# Export data for analysis
gh-notif discussions analytics --format json > analytics.json
```

### Integration with Scripts

```bash
#!/bin/bash
# Daily discussion report

echo "📊 Daily Discussion Report"
echo "=========================="

# Get unanswered questions
UNANSWERED=$(gh-notif discussions unanswered --format json | jq length)
echo "Unanswered questions: $UNANSWERED"

# Get trending discussions
gh-notif discussions trending --limit 5

# Get analytics
gh-notif discussions analytics
```

## Tips and Best Practices

### Efficient Monitoring

1. **Use Repository Filters**: Focus on specific repositories with `--repo`
2. **Set Reasonable Limits**: Use `--limit` to avoid information overload
3. **Leverage Caching**: Enable caching for faster repeated queries
4. **Use Interactive Mode**: For browsing and exploration

### Community Engagement

1. **Monitor Unanswered Questions**: Regular checks help community health
2. **Track Trending Topics**: Stay informed about community interests
3. **Use Analytics**: Understand engagement patterns and community growth
4. **Export Data**: Use JSON format for custom analysis and reporting

### Performance Optimization

1. **Enable Caching**: Reduces API calls and improves response times
2. **Use Specific Filters**: Narrow down results to reduce processing
3. **Batch Operations**: Process multiple repositories efficiently
4. **Background Refresh**: Keep data fresh without blocking operations

The discussions feature provides a comprehensive toolkit for monitoring and engaging with GitHub's discussion communities, helping you stay connected with your projects and contribute effectively to open source conversations.
