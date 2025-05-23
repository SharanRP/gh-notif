# User Acceptance Testing Script

This document provides comprehensive user acceptance testing scenarios for gh-notif. Each scenario includes step-by-step instructions, expected outcomes, and validation criteria.

## Prerequisites

- [ ] gh-notif is installed on the test system
- [ ] GitHub account with notifications available
- [ ] GitHub personal access token with appropriate scopes
- [ ] Test environment is clean (no existing gh-notif configuration)

## Test Environment Setup

### Environment Information
- **Operating System**: ________________
- **gh-notif Version**: ________________
- **Test Date**: ________________
- **Tester**: ________________

### Initial Setup
1. Verify gh-notif installation:
   ```bash
   gh-notif --version
   ```
   **Expected**: Version information displayed
   **Result**: ☐ Pass ☐ Fail

2. Check initial state:
   ```bash
   gh-notif auth status
   ```
   **Expected**: "Not authenticated" message
   **Result**: ☐ Pass ☐ Fail

## Scenario 1: First-Time User Experience

### Objective
Verify that new users can successfully set up and use gh-notif.

### Steps

1. **Run first-time setup**
   ```bash
   gh-notif firstrun
   ```
   **Expected**: Welcome message and setup wizard
   **Result**: ☐ Pass ☐ Fail
   **Notes**: ________________

2. **Complete authentication**
   ```bash
   gh-notif auth login
   ```
   **Expected**: Device flow instructions or token prompt
   **Result**: ☐ Pass ☐ Fail
   **Notes**: ________________

3. **Verify authentication**
   ```bash
   gh-notif auth status
   ```
   **Expected**: "Authenticated as [username]" message
   **Result**: ☐ Pass ☐ Fail
   **Notes**: ________________

4. **Run tutorial**
   ```bash
   gh-notif tutorial
   ```
   **Expected**: Interactive tutorial starts
   **Result**: ☐ Pass ☐ Fail
   **Notes**: ________________

5. **List notifications**
   ```bash
   gh-notif list
   ```
   **Expected**: List of notifications displayed
   **Result**: ☐ Pass ☐ Fail
   **Notes**: ________________

### Validation Criteria
- [ ] Setup process is intuitive and clear
- [ ] Authentication works without errors
- [ ] Tutorial provides helpful guidance
- [ ] Initial notification list loads successfully
- [ ] Error messages (if any) are helpful and actionable

## Scenario 2: Core Functionality

### Objective
Verify all core features work as documented.

### Steps

1. **List all notifications**
   ```bash
   gh-notif list --all
   ```
   **Expected**: All notifications displayed with proper formatting
   **Result**: ☐ Pass ☐ Fail

2. **List with limit**
   ```bash
   gh-notif list --limit 10
   ```
   **Expected**: Exactly 10 notifications displayed
   **Result**: ☐ Pass ☐ Fail

3. **Filter unread notifications**
   ```bash
   gh-notif list --filter "is:unread"
   ```
   **Expected**: Only unread notifications displayed
   **Result**: ☐ Pass ☐ Fail

4. **Filter by repository**
   ```bash
   gh-notif list --filter "repo:owner/repo"
   ```
   **Expected**: Only notifications from specified repository
   **Result**: ☐ Pass ☐ Fail

5. **Filter by type**
   ```bash
   gh-notif list --filter "type:PullRequest"
   ```
   **Expected**: Only pull request notifications
   **Result**: ☐ Pass ☐ Fail

6. **Complex filter**
   ```bash
   gh-notif list --filter "is:unread AND type:PullRequest"
   ```
   **Expected**: Unread pull request notifications only
   **Result**: ☐ Pass ☐ Fail

7. **Group by repository**
   ```bash
   gh-notif group --by repository
   ```
   **Expected**: Notifications grouped by repository
   **Result**: ☐ Pass ☐ Fail

8. **Group by type**
   ```bash
   gh-notif group --by type
   ```
   **Expected**: Notifications grouped by type
   **Result**: ☐ Pass ☐ Fail

9. **Search notifications**
   ```bash
   gh-notif search "bug fix"
   ```
   **Expected**: Notifications containing "bug fix" in title/body
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] All commands execute without errors
- [ ] Filtering works accurately
- [ ] Grouping displays logical organization
- [ ] Search returns relevant results
- [ ] Output formatting is consistent and readable

## Scenario 3: Notification Actions

### Objective
Verify notification action commands work correctly.

### Prerequisites
- At least one unread notification available
- Note notification ID for testing: ________________

### Steps

1. **Mark notification as read (dry run)**
   ```bash
   gh-notif read [NOTIFICATION_ID] --dry-run
   ```
   **Expected**: Shows what would be done without executing
   **Result**: ☐ Pass ☐ Fail

2. **Mark notification as read**
   ```bash
   gh-notif read [NOTIFICATION_ID]
   ```
   **Expected**: Confirmation message, notification marked as read
   **Result**: ☐ Pass ☐ Fail

3. **Open notification in browser (dry run)**
   ```bash
   gh-notif open [NOTIFICATION_ID] --dry-run
   ```
   **Expected**: Shows URL that would be opened
   **Result**: ☐ Pass ☐ Fail

4. **Subscribe to thread (dry run)**
   ```bash
   gh-notif subscribe [NOTIFICATION_ID] --dry-run
   ```
   **Expected**: Shows subscription action preview
   **Result**: ☐ Pass ☐ Fail

5. **Archive notification (dry run)**
   ```bash
   gh-notif archive [NOTIFICATION_ID] --dry-run
   ```
   **Expected**: Shows archive action preview
   **Result**: ☐ Pass ☐ Fail

6. **Batch mark as read**
   ```bash
   gh-notif list --filter "is:unread" --limit 5 | gh-notif read --batch
   ```
   **Expected**: Multiple notifications marked as read
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] Dry run mode works correctly
- [ ] Actions execute successfully
- [ ] Batch operations work efficiently
- [ ] Confirmation messages are clear
- [ ] State changes are reflected in subsequent listings

## Scenario 4: Configuration Management

### Objective
Verify configuration system works properly.

### Steps

1. **List current configuration**
   ```bash
   gh-notif config list
   ```
   **Expected**: Current configuration displayed
   **Result**: ☐ Pass ☐ Fail

2. **Set configuration value**
   ```bash
   gh-notif config set display.limit 25
   ```
   **Expected**: Configuration updated successfully
   **Result**: ☐ Pass ☐ Fail

3. **Get configuration value**
   ```bash
   gh-notif config get display.limit
   ```
   **Expected**: Returns "25"
   **Result**: ☐ Pass ☐ Fail

4. **Test configuration effect**
   ```bash
   gh-notif list
   ```
   **Expected**: Shows 25 notifications (or available count)
   **Result**: ☐ Pass ☐ Fail

5. **Set invalid configuration**
   ```bash
   gh-notif config set display.limit invalid
   ```
   **Expected**: Error message about invalid value
   **Result**: ☐ Pass ☐ Fail

6. **Reset configuration**
   ```bash
   gh-notif config reset display.limit
   ```
   **Expected**: Configuration reset to default
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] Configuration changes persist
- [ ] Invalid values are rejected with helpful errors
- [ ] Configuration affects application behavior
- [ ] Reset functionality works correctly

## Scenario 5: Output Formats and Export

### Objective
Verify different output formats and export functionality.

### Steps

1. **JSON output**
   ```bash
   gh-notif list --format json --limit 5
   ```
   **Expected**: Valid JSON output
   **Result**: ☐ Pass ☐ Fail

2. **CSV output**
   ```bash
   gh-notif list --format csv --limit 5
   ```
   **Expected**: Valid CSV output with headers
   **Result**: ☐ Pass ☐ Fail

3. **Table output**
   ```bash
   gh-notif list --format table --limit 5
   ```
   **Expected**: Formatted table output
   **Result**: ☐ Pass ☐ Fail

4. **Export to file**
   ```bash
   gh-notif list --format json --output notifications.json --limit 10
   ```
   **Expected**: File created with JSON content
   **Result**: ☐ Pass ☐ Fail

5. **Verify export file**
   ```bash
   cat notifications.json
   ```
   **Expected**: Valid JSON content in file
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] All output formats are valid and well-formatted
- [ ] Export functionality creates correct files
- [ ] File content matches expected format
- [ ] Large exports complete successfully

## Scenario 6: Error Handling and Edge Cases

### Objective
Verify proper error handling and edge case behavior.

### Steps

1. **Invalid command**
   ```bash
   gh-notif invalidcommand
   ```
   **Expected**: Helpful error message with suggestions
   **Result**: ☐ Pass ☐ Fail

2. **Invalid filter syntax**
   ```bash
   gh-notif list --filter "invalid:syntax:here"
   ```
   **Expected**: Clear error message about filter syntax
   **Result**: ☐ Pass ☐ Fail

3. **Network connectivity test**
   - Disconnect network
   ```bash
   gh-notif list
   ```
   **Expected**: Network error message with retry suggestion
   **Result**: ☐ Pass ☐ Fail

4. **Invalid notification ID**
   ```bash
   gh-notif read invalid-id
   ```
   **Expected**: Error message about invalid ID
   **Result**: ☐ Pass ☐ Fail

5. **Permission denied scenario**
   ```bash
   gh-notif open [PRIVATE_REPO_NOTIFICATION_ID]
   ```
   **Expected**: Appropriate permission error message
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] Error messages are clear and actionable
- [ ] Application handles network issues gracefully
- [ ] Invalid inputs are rejected with helpful feedback
- [ ] No crashes or unexpected behavior

## Scenario 7: Performance and Usability

### Objective
Verify performance meets user expectations and usability is good.

### Steps

1. **Startup time test**
   - Measure time from command execution to output
   ```bash
   time gh-notif --version
   ```
   **Expected**: < 2 seconds
   **Result**: ☐ Pass ☐ Fail
   **Time**: ________________

2. **Large list performance**
   ```bash
   time gh-notif list --limit 100
   ```
   **Expected**: < 10 seconds
   **Result**: ☐ Pass ☐ Fail
   **Time**: ________________

3. **Complex filter performance**
   ```bash
   time gh-notif list --filter "is:unread AND (type:PullRequest OR type:Issue)"
   ```
   **Expected**: < 5 seconds
   **Result**: ☐ Pass ☐ Fail
   **Time**: ________________

4. **Help system usability**
   ```bash
   gh-notif --help
   gh-notif list --help
   gh-notif filter --help
   ```
   **Expected**: Comprehensive, clear help text
   **Result**: ☐ Pass ☐ Fail

5. **Tab completion test** (if available)
   ```bash
   gh-notif [TAB][TAB]
   ```
   **Expected**: Command completion suggestions
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] Commands execute within acceptable time limits
- [ ] Help text is comprehensive and helpful
- [ ] Tab completion works (where supported)
- [ ] User interface is intuitive

## Scenario 8: Integration and Compatibility

### Objective
Verify integration with system and other tools.

### Steps

1. **Shell integration**
   ```bash
   gh-notif list | head -5
   gh-notif list | grep "unread"
   ```
   **Expected**: Proper pipe behavior
   **Result**: ☐ Pass ☐ Fail

2. **Exit codes**
   ```bash
   gh-notif list; echo $?
   gh-notif invalidcommand; echo $?
   ```
   **Expected**: 0 for success, non-zero for errors
   **Result**: ☐ Pass ☐ Fail

3. **Environment variable support**
   ```bash
   GH_NOTIF_CONFIG=/tmp/test-config.yaml gh-notif config list
   ```
   **Expected**: Uses specified config file
   **Result**: ☐ Pass ☐ Fail

4. **Signal handling**
   - Start long-running command and press Ctrl+C
   ```bash
   gh-notif watch
   ```
   **Expected**: Graceful shutdown
   **Result**: ☐ Pass ☐ Fail

### Validation Criteria
- [ ] Proper shell integration
- [ ] Correct exit codes
- [ ] Environment variables work
- [ ] Signal handling is graceful

## Test Summary

### Overall Results
- **Total Test Cases**: ________________
- **Passed**: ________________
- **Failed**: ________________
- **Skipped**: ________________

### Critical Issues Found
1. ________________
2. ________________
3. ________________

### Minor Issues Found
1. ________________
2. ________________
3. ________________

### Recommendations
1. ________________
2. ________________
3. ________________

### Sign-off
- **Tester**: ________________
- **Date**: ________________
- **Approval**: ☐ Approved ☐ Approved with conditions ☐ Rejected

### Notes
________________
________________
________________
