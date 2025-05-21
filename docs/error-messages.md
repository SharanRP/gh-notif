# Error Messages Guide

This document provides a comprehensive list of error messages that you might encounter when using gh-notif, along with explanations and suggested solutions.

## Authentication Errors

### Error: Not authenticated

**Message:**
```
Error: Not authenticated. Please run 'gh-notif auth login' to authenticate with GitHub.
```

**Explanation:**
You are trying to use a command that requires authentication, but you are not currently authenticated with GitHub.

**Solution:**
Run `gh-notif auth login` to authenticate with GitHub.

### Error: Authentication token expired

**Message:**
```
Error: Authentication token has expired. Please run 'gh-notif auth refresh' to refresh your token.
```

**Explanation:**
Your authentication token has expired and needs to be refreshed.

**Solution:**
Run `gh-notif auth refresh` to refresh your authentication token.

### Error: Authentication failed

**Message:**
```
Error: Authentication failed: <reason>
```

**Explanation:**
The authentication process failed for the specified reason.

**Solution:**
- Check your internet connection
- Verify that you entered the correct device code
- Ensure that you have granted the necessary permissions
- Try again with `gh-notif auth login`

### Error: Token storage not available

**Message:**
```
Error: Token storage not available: <reason>
```

**Explanation:**
The system could not access the token storage mechanism.

**Solution:**
- If using keyring storage, ensure that your system's keyring is unlocked
- If using file storage, ensure that the file is accessible and not corrupted
- Try setting a different storage method with `gh-notif config set auth.token_storage file`

## API Errors

### Error: API rate limit exceeded

**Message:**
```
Error: GitHub API rate limit exceeded. Please try again later.
```

**Explanation:**
You have exceeded the GitHub API rate limit.

**Solution:**
- Wait until the rate limit resets (usually 1 hour)
- Reduce the frequency of API requests
- Use caching to minimize API requests with `gh-notif config set advanced.cache_ttl 3600`

### Error: API request failed

**Message:**
```
Error: GitHub API request failed: <status_code> <reason>
```

**Explanation:**
A request to the GitHub API failed with the specified status code and reason.

**Solution:**
- Check your internet connection
- Verify that the GitHub API is operational
- Ensure that you have the necessary permissions for the requested resource
- Try again later

### Error: Connection timeout

**Message:**
```
Error: Connection timeout while connecting to GitHub API.
```

**Explanation:**
The connection to the GitHub API timed out.

**Solution:**
- Check your internet connection
- Increase the timeout value with `gh-notif config set api.timeout 60`
- Try again later

## Configuration Errors

### Error: Configuration file not found

**Message:**
```
Error: Configuration file not found at <path>
```

**Explanation:**
The configuration file could not be found at the specified path.

**Solution:**
- Create a new configuration file with `gh-notif config init`
- Specify a different configuration file with `--config <path>`
- Run the setup wizard with `gh-notif wizard`

### Error: Invalid configuration

**Message:**
```
Error: Invalid configuration: <reason>
```

**Explanation:**
The configuration file contains invalid settings.

**Solution:**
- Edit the configuration file with `gh-notif config edit`
- Reset to default configuration with `gh-notif config reset`
- Run the setup wizard with `gh-notif wizard`

### Error: Configuration value not found

**Message:**
```
Error: Configuration value '<key>' not found
```

**Explanation:**
The specified configuration key does not exist.

**Solution:**
- Check the key name for typos
- List all available configuration keys with `gh-notif config list`
- Set the value with `gh-notif config set <key> <value>`

## Filter Errors

### Error: Invalid filter expression

**Message:**
```
Error: Invalid filter expression: <reason>
```

**Explanation:**
The filter expression is invalid for the specified reason.

**Solution:**
- Check the filter syntax
- Use simpler filter expressions
- Use the filter wizard with `gh-notif filter create --interactive`

### Error: Filter not found

**Message:**
```
Error: Filter '<name>' not found
```

**Explanation:**
The specified named filter does not exist.

**Solution:**
- Check the filter name for typos
- List all available filters with `gh-notif filter list`
- Create the filter with `gh-notif filter save <name> <expression>`

## Cache Errors

### Error: Cache initialization failed

**Message:**
```
Error: Cache initialization failed: <reason>
```

**Explanation:**
The cache could not be initialized for the specified reason.

**Solution:**
- Check that the cache directory is accessible
- Try a different cache type with `gh-notif config set advanced.cache_type memory`
- Clear the cache with `gh-notif cache clear`

### Error: Cache write failed

**Message:**
```
Error: Cache write failed: <reason>
```

**Explanation:**
Writing to the cache failed for the specified reason.

**Solution:**
- Check that the cache directory is writable
- Ensure that there is sufficient disk space
- Try a different cache type with `gh-notif config set advanced.cache_type memory`

## UI Errors

### Error: Terminal size too small

**Message:**
```
Error: Terminal size too small. Minimum required: <width>x<height>
```

**Explanation:**
The terminal window is too small to display the UI properly.

**Solution:**
- Resize your terminal window to be larger
- Use a different view mode with `gh-notif ui --view compact`
- Use the non-interactive mode with `--no-interactive`

### Error: Unsupported terminal

**Message:**
```
Error: Unsupported terminal. Please use a terminal that supports ANSI escape sequences.
```

**Explanation:**
Your terminal does not support the features required by the UI.

**Solution:**
- Use a terminal that supports ANSI escape sequences
- Use the non-interactive mode with `--no-interactive`
- Use the `--no-color` flag to disable colors

## Notification Errors

### Error: Notification not found

**Message:**
```
Error: Notification with ID '<id>' not found
```

**Explanation:**
The specified notification ID does not exist or is not accessible.

**Solution:**
- Check the notification ID for typos
- List available notifications with `gh-notif list`
- Ensure that you have permission to access the notification

### Error: No notifications found

**Message:**
```
No notifications found matching the specified criteria.
```

**Explanation:**
No notifications were found that match your filter criteria.

**Solution:**
- Use less restrictive filter criteria
- Check for typos in your filter expression
- Try listing all notifications with `gh-notif list --all`

## General Errors

### Error: Command failed

**Message:**
```
Error: Command failed: <reason>
```

**Explanation:**
The command failed for the specified reason.

**Solution:**
- Check the error message for specific details
- Try running the command with the `--debug` flag for more information
- Check the documentation for the correct usage of the command

### Error: Unexpected error

**Message:**
```
Error: Unexpected error: <error>
```

**Explanation:**
An unexpected error occurred.

**Solution:**
- Try running the command again
- Run the command with the `--debug` flag for more information
- Report the issue on GitHub if it persists

## Getting Help

If you encounter an error that is not listed here or if the suggested solutions do not resolve your issue, you can:

1. Run the command with the `--debug` flag to get more detailed information
2. Check the documentation with `gh-notif help <command>`
3. Run the interactive tutorial with `gh-notif tutorial`
4. Report the issue on GitHub at https://github.com/user/gh-notif/issues
