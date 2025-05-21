# Keyboard Shortcuts Guide

This document provides a comprehensive list of keyboard shortcuts for the gh-notif terminal UI.

## Global Shortcuts

These shortcuts work in all views of the terminal UI.

| Shortcut | Description |
|----------|-------------|
| `?` | Show help |
| `q`, `Ctrl+C`, `Esc` | Quit |
| `Ctrl+L` | Refresh the screen |
| `Ctrl+S` | Save current view as a filter |
| `Ctrl+R` | Reload notifications |
| `Ctrl+F` | Search within the current view |
| `Ctrl+G` | Go to notification by ID |
| `Ctrl+H` | Show command history |
| `Ctrl+T` | Toggle theme (dark/light) |
| `Ctrl+Z` | Undo last action |
| `F1` | Show tutorial |
| `F2` | Open configuration |
| `F5` | Refresh notifications |
| `F10` | Show menu |
| `Alt+1` to `Alt+9` | Switch to view 1-9 |

## Navigation Shortcuts

These shortcuts are for navigating within the UI.

| Shortcut | Description |
|----------|-------------|
| `↑`, `k` | Move up |
| `↓`, `j` | Move down |
| `←`, `h` | Move left |
| `→`, `l` | Move right |
| `PgUp`, `Ctrl+B` | Page up |
| `PgDn`, `Ctrl+F` | Page down |
| `Home`, `g` | Go to top |
| `End`, `G` | Go to bottom |
| `Tab` | Next section |
| `Shift+Tab` | Previous section |
| `Enter` | Select/Expand item |
| `Backspace` | Go back |
| `Space` | Toggle selection |

## List View Shortcuts

These shortcuts are specific to the notification list view.

| Shortcut | Description |
|----------|-------------|
| `r` | Mark as read |
| `u` | Mark as unread |
| `o` | Open in browser |
| `a` | Archive |
| `s` | Subscribe |
| `S` | Unsubscribe |
| `m` | Mute repository |
| `M` | Unmute repository |
| `f` | Filter |
| `F` | Clear filter |
| `g` | Group |
| `G` | Clear grouping |
| `t` | Sort |
| `T` | Reverse sort order |
| `v` | Change view mode |
| `c` | Copy notification URL |
| `y` | Copy notification ID |
| `/` | Search |
| `n` | Next search result |
| `N` | Previous search result |
| `1` | Switch to compact view |
| `2` | Switch to detailed view |
| `3` | Switch to split view |
| `4` | Switch to table view |

## Detail View Shortcuts

These shortcuts are specific to the notification detail view.

| Shortcut | Description |
|----------|-------------|
| `r` | Mark as read |
| `u` | Mark as unread |
| `o` | Open in browser |
| `a` | Archive |
| `s` | Subscribe |
| `S` | Unsubscribe |
| `c` | Copy URL |
| `y` | Copy ID |
| `b` | Back to list |
| `e` | Expand/collapse sections |
| `p` | Preview in Markdown |
| `d` | Show diff (for pull requests) |
| `i` | Show issue details |
| `t` | Show timeline |
| `m` | Show comments |
| `w` | Watch thread |

## Filter View Shortcuts

These shortcuts are specific to the filter view.

| Shortcut | Description |
|----------|-------------|
| `Enter` | Apply filter |
| `Esc` | Cancel |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Ctrl+Space` | Show autocomplete |
| `Alt+S` | Save filter |
| `Alt+L` | Load filter |
| `Alt+C` | Clear filter |
| `Alt+H` | Show filter history |
| `Alt+E` | Edit filter as text |
| `Alt+V` | Validate filter |

## Group View Shortcuts

These shortcuts are specific to the group view.

| Shortcut | Description |
|----------|-------------|
| `Enter` | Expand/collapse group |
| `Space` | Select/deselect group |
| `r` | Mark group as read |
| `a` | Archive group |
| `s` | Subscribe to group |
| `S` | Unsubscribe from group |
| `o` | Open all in group |
| `g` | Change grouping |
| `f` | Filter within group |
| `t` | Sort group |
| `c` | Collapse all groups |
| `e` | Expand all groups |

## Search View Shortcuts

These shortcuts are specific to the search view.

| Shortcut | Description |
|----------|-------------|
| `Enter` | Execute search |
| `Esc` | Cancel search |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Ctrl+Space` | Show search history |
| `Alt+R` | Toggle regex mode |
| `Alt+C` | Toggle case sensitivity |
| `Alt+W` | Toggle whole word |
| `Alt+S` | Save search |
| `Alt+H` | Show search history |
| `Alt+A` | Advanced search options |

## Watch Mode Shortcuts

These shortcuts are specific to the watch mode.

| Shortcut | Description |
|----------|-------------|
| `Space` | Pause/resume watching |
| `r` | Mark as read |
| `o` | Open in browser |
| `f` | Filter notifications |
| `i` | Show notification details |
| `d` | Toggle desktop notifications |
| `+` | Increase refresh interval |
| `-` | Decrease refresh interval |
| `c` | Clear notifications |
| `s` | Show statistics |

## Action Mode Shortcuts

These shortcuts are specific to the action mode.

| Shortcut | Description |
|----------|-------------|
| `Enter` | Execute action |
| `Esc` | Cancel action |
| `Tab` | Next option |
| `Shift+Tab` | Previous option |
| `Space` | Toggle option |
| `a` | Select all |
| `n` | Select none |
| `i` | Invert selection |
| `f` | Filter selection |
| `s` | Save selection |
| `l` | Load selection |

## Help View Shortcuts

These shortcuts are specific to the help view.

| Shortcut | Description |
|----------|-------------|
| `q`, `Esc` | Close help |
| `↑`, `k` | Scroll up |
| `↓`, `j` | Scroll down |
| `PgUp` | Page up |
| `PgDn` | Page down |
| `Home` | Go to top |
| `End` | Go to bottom |
| `Tab` | Next section |
| `Shift+Tab` | Previous section |
| `f` | Find in help |
| `n` | Next search result |
| `N` | Previous search result |
| `1`-`9` | Jump to section |

## Accessibility Shortcuts

These shortcuts are for accessibility features.

| Shortcut | Description |
|----------|-------------|
| `Alt+A` | Toggle screen reader mode |
| `Alt+C` | Toggle high contrast mode |
| `Alt+L` | Toggle large text mode |
| `Alt+U` | Toggle Unicode/ASCII mode |
| `Alt+M` | Toggle motion reduction |
| `Alt+F` | Toggle focus highlighting |
| `Alt+T` | Toggle text descriptions |
| `Alt+K` | Show keyboard shortcuts |

## Customizing Shortcuts

You can customize keyboard shortcuts in the configuration file:

```yaml
ui:
  shortcuts:
    quit: ["q", "ctrl+c", "esc"]
    help: ["?"]
    refresh: ["ctrl+r", "f5"]
    # Add your custom shortcuts here
```

Or using the configuration command:

```bash
gh-notif config set ui.shortcuts.quit '["q", "ctrl+c", "esc"]'
```

You can reset shortcuts to defaults with:

```bash
gh-notif config reset ui.shortcuts
```
