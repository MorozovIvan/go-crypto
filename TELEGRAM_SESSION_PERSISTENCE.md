# Telegram Session Persistence Implementation

## Overview

The Telegram service now includes robust session persistence that maintains user authentication across server restarts. This implementation ensures users don't need to re-authenticate every time the server is restarted.

## Key Features

### 1. **Dual-Layer Session Storage**

- **Session File (`session.json`)**: Contains Telegram's low-level authentication keys and session data
- **Auth State File (`auth_state.json`)**: Contains high-level application state including user ID, phone, and authentication status

### 2. **Session Recovery Mechanism**

When `AUTH_KEY_UNREGISTERED` errors occur, the system attempts automatic session recovery:

- Reinitializes the Telegram client
- Reloads the session file
- Validates the recovered session
- Falls back to clearing auth state if recovery fails

### 3. **Session Validation on Startup**

- Checks for existing session files on service initialization
- Loads and validates persistent auth state
- Performs session validation with proper timing
- Handles edge cases like expired or corrupted sessions

### 4. **Enhanced Error Handling**

- Graceful handling of `AUTH_KEY_UNREGISTERED` errors
- Automatic retry mechanisms for temporary network issues
- Proper cleanup of invalid session data
- Comprehensive logging for debugging

## Session Persistence Flow

### Authentication Process

1. User initiates authentication with phone number
2. User verifies with SMS code or 2FA password
3. On successful authentication:
   - `session.json` is automatically created by gotd library
   - `auth_state.json` is created with user details and timestamp
   - Authentication state is marked as active

### Server Restart Process

1. Server starts and initializes Telegram service
2. Service checks for existing `session.json` file
3. If found, loads `auth_state.json` for application state
4. Validates session age (expires after 7 days)
5. Performs background session validation
6. If validation fails, attempts session recovery
7. User remains authenticated if session is valid

### Session Recovery Process

1. Detects `AUTH_KEY_UNREGISTERED` error
2. Attempts to recover by reinitializing client with existing session file
3. Tests recovered session by calling `GetCurrentUser`
4. If successful, user remains authenticated
5. If failed, clears auth state and requires re-authentication

## Configuration

### Session Expiry

- Auth state expires after 7 days of inactivity
- Session files are preserved for potential recovery
- Users need to re-authenticate after expiry

### File Locations

- `session.json`: Root directory (created by gotd library)
- `auth_state.json`: Root directory (created by application)
- `telegram_service.log`: Detailed logging of all operations

## Auth State Format

```json
{
  "user_auth": true,
  "user_id": 123456789,
  "phone": "+1234567890",
  "saved_at": "2025-06-17T23:11:46.000Z",
  "version": "1.0"
}
```

## Benefits

1. **User Experience**: Users stay logged in across server restarts
2. **Reliability**: Automatic recovery from temporary session issues
3. **Security**: Session expiry and validation mechanisms
4. **Debugging**: Comprehensive logging for troubleshooting
5. **Robustness**: Handles edge cases and error conditions

## Testing Session Persistence

To test session persistence:

1. Authenticate a user through the API
2. Restart the server
3. Check `/api/telegram/status` endpoint
4. User should remain authenticated
5. Check logs for session restoration messages

## Troubleshooting

### Common Issues

**Session not restored after restart:**

- Check if `auth_state.json` exists
- Verify session is not older than 7 days
- Check logs for session validation errors

**AUTH_KEY_UNREGISTERED errors:**

- Session recovery will be attempted automatically
- If recovery fails, user needs to re-authenticate
- Check network connectivity and Telegram API status

**Session validation timeouts:**

- Increase client ready timeout in service initialization
- Check network stability
- Verify Telegram API accessibility

### Log Messages

- `"Session restored successfully from persistent state"`: Session persistence working correctly
- `"Attempting session recovery"`: Automatic recovery in progress
- `"Session recovery successful"`: Recovery completed successfully
- `"Session recovery failed, clearing auth state"`: User needs to re-authenticate

## Security Considerations

1. Session files are created with restricted permissions (0600)
2. Auth state includes timestamp for age validation
3. Automatic cleanup of expired or invalid sessions
4. No sensitive data stored in auth state file
5. Session validation prevents unauthorized access

This implementation provides a robust foundation for maintaining Telegram authentication across server restarts while ensuring security and reliability.
