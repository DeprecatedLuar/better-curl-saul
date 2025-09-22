# Universal Logging System Specification

## System Philosophy

This logging system implements **simplicity over complexity** with **visual efficiency** for development workflow. It provides centralized logging control while maintaining instant visual feedback through emoji-based categorization.

### Core Principles

1. **Single Logger Per Project** - All logging must use the same centralized system
2. **Visual Efficiency** - Emoji prefixes for instant problem identification
3. **Environment Awareness** - Automatic behavior based on NODE_ENV
4. **Zero Configuration** - Works out of the box with sensible defaults
5. **Console-First** - File logging is opt-in for specific use cases

## Complete System Architecture

```
src/modules/logging/
├── logger.js              # Core logger factory and implementation
├── formatters.js          # Message formatting utilities
├── file-logger.js         # Optional file logging (opt-in)
├── index.js               # Public API exports
├── .purpose.md            # Architectural context
└── doc.md                 # This detailed specification
```

## Logger Implementation Specification

### Core Logger Factory

```javascript
// modules/logging/logger.js

const EMOJI_MAP = {
  info: '🔵',
  success: '✅', 
  error: '❌',
  warn: '⚠️',
  debug: '🔍',
  config: '🔧',
  network: '🌐',
  security: '🔒',
  performance: '⚡'
};

const isDevelopment = process.env.NODE_ENV !== 'production';

const getLogger = (component) => ({
  info: (msg, ctx) => log('info', component, msg, ctx),
  success: (msg, ctx) => log('success', component, msg, ctx),
  error: (msg, ctx) => log('error', component, msg, ctx),
  warn: (msg, ctx) => log('warn', component, msg, ctx),
  debug: (msg, ctx) => log('debug', component, msg, ctx),
  config: (msg, ctx) => log('config', component, msg, ctx),
  network: (msg, ctx) => log('network', component, msg, ctx),
  security: (msg, ctx) => log('security', component, msg, ctx),
  performance: (msg, ctx) => log('performance', component, msg, ctx)
});

function log(level, component, message, context = null) {
  if (!isDevelopment) return; // Production silence
  
  const emoji = EMOJI_MAP[level];
  const timestamp = new Date().toLocaleTimeString('en-US', { 
    hour12: false, 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit' 
  });
  
  let logMessage = `${emoji} [${component}] ${message}`;
  
  if (context) {
    logMessage += ` | Context: ${JSON.stringify(context)}`;
  }
  
  // Optional timestamp at end (uncomment if needed)
  // logMessage += ` [${timestamp}]`;
  
  console.log(logMessage);
}

module.exports = { getLogger };
```

### Standard Output Format

```
🔵 [Template] Server config loaded successfully
🔍 [Template] CSS variable set: --app-name = Zhang A.I. | Context: {"key": "appName"}
❌ [JobProcessor] Database timeout | Context: {"error": "connection_lost", "jobId": "123"}
✅ [AuthService] User authenticated | Context: {"userId": "user_456", "method": "oauth"}
⚠️ [APIClient] Rate limit approaching | Context: {"remaining": 5, "reset": 300}
🔧 [ConfigLoader] Using development configuration
🌐 [HTTPClient] Request completed | Context: {"url": "/api/status", "status": 200, "duration": 150}
🔒 [SecurityMiddleware] Access denied | Context: {"ip": "192.168.1.1", "reason": "invalid_token"}
⚡ [DatabaseQuery] Slow query detected | Context: {"query": "SELECT * FROM users", "duration": 2500}
```

## Environment Control

### Simple Environment Switching

```javascript
// Controlled by NODE_ENV only - no additional variables needed
const isDevelopment = process.env.NODE_ENV !== 'production';

// Development: All logs shown
// Production: All logs hidden
// No complex log level hierarchies
```

### Environment-Specific Behavior

```bash
# Development (default)
NODE_ENV=development  # Shows all logs with full context

# Production
NODE_ENV=production   # Silences all logs for performance

# Local development
# (no NODE_ENV set)   # Defaults to development behavior
```

## Usage Patterns

### Basic Logger Usage

```javascript
// Import logger in any module
const { getLogger } = require('../modules/logging');
const logger = getLogger('ComponentName');

// Standard logging calls
logger.info('Operation started');
logger.success('Data saved successfully', { recordId: 123 });
logger.error('Connection failed', { error: error.message, retries: 3 });
logger.warn('Memory usage high', { usage: '85%', threshold: '80%' });
logger.debug('Processing step completed', { step: 'validation', data: payload });

// Specialized logging
logger.config('Loaded settings from environment');
logger.network('API call completed', { endpoint: '/users', status: 200 });
logger.security('Authentication attempt', { userId: 'user123', success: true });
logger.performance('Query executed', { query: 'getUserData', duration: 145 });
```

### Component-Specific Loggers

```javascript
// Each module gets its own tagged logger
const apiLogger = getLogger('APIClient');
const authLogger = getLogger('AuthService'); 
const dbLogger = getLogger('Database');
const configLogger = getLogger('ConfigLoader');

// Context automatically includes component identification
apiLogger.info('Request sent');     // 🔵 [APIClient] Request sent
authLogger.error('Login failed');   // ❌ [AuthService] Login failed
```

### Context Objects

```javascript
// Rich context for debugging
logger.error('Payment processing failed', {
  orderId: 'order_123',
  amount: 99.99,
  currency: 'USD',
  error: error.message,
  gateway: 'stripe',
  userId: 'user_456'
});

// Structured data for analysis
logger.performance('Database operation completed', {
  operation: 'SELECT',
  table: 'users',
  duration: 250,
  rowsReturned: 15,
  cacheHit: false
});
```

## File Logging (Optional)

### When to Enable File Logging

**Use file logging for:**
- Performance analysis and timing metrics
- Audit trails for compliance requirements
- Production debugging of specific issues
- Historical analysis of system behavior

**Don't use file logging for:**
- Standard development workflow
- Simple applications without performance requirements
- Systems already using external logging services

### File Logger Implementation

```javascript
// modules/logging/file-logger.js

const fs = require('fs');
const path = require('path');

class FileLogger {
  constructor(logDir = 'logs', maxFiles = 3) {
    this.logDir = logDir;
    this.maxFiles = maxFiles;
    this.ensureLogDirectory();
  }
  
  log(level, component, message, context, type = 'app') {
    const timestamp = new Date().toISOString();
    const logEntry = {
      timestamp,
      level,
      component,
      message,
      context
    };
    
    const fileName = `${type}-${new Date().getDate()}.log`;
    const filePath = path.join(this.logDir, fileName);
    
    fs.appendFileSync(filePath, JSON.stringify(logEntry) + '\n');
    this.rotateLogsIfNeeded();
  }
  
  // Implementation details for rotation...
}

// Optional file logger export
const getFileLogger = (component, options = {}) => {
  const fileLogger = new FileLogger(options.logDir, options.maxFiles);
  
  return {
    logToFile: (level, message, context, type) => 
      fileLogger.log(level, component, message, context, type)
  };
};

module.exports = { getFileLogger };
```

## Strict Implementation Requirements

### Mandatory Integration Checklist

**Every new module MUST:**
- [ ] Import logger from modules/logging
- [ ] Use component-tagged logger (not global console)
- [ ] Include contextual information in error logs
- [ ] Follow emoji-based logging levels
- [ ] Never create custom logging implementations

### Forbidden Patterns

```javascript
// ❌ FORBIDDEN - Direct console usage
console.log('User logged in');
console.error('Something went wrong');

// ❌ FORBIDDEN - Custom logging systems
import winston from 'winston';
const customLogger = winston.createLogger();

// ❌ FORBIDDEN - Multiple logging libraries
const log4js = require('log4js');
const bunyan = require('bunyan');

// ✅ CORRECT - Framework logging
const { getLogger } = require('../modules/logging');
const logger = getLogger('UserService');
logger.success('User logged in', { userId: 'user123' });
```

### Code Review Requirements

**Before any commit:**
- [ ] No new logging libraries in package.json
- [ ] All log statements use centralized logger
- [ ] Logger imports match framework patterns
- [ ] Component names are descriptive and consistent
- [ ] Context objects provided for errors and important events

## Migration Guide

### Converting Existing Projects

1. **Audit Current Logging:** Find all console.log, print(), etc.
2. **Install Framework Logger:** Copy modules/logging/ to project
3. **Replace Console Calls:** Convert to component-tagged loggers
4. **Add Context:** Enhance logs with structured context objects
5. **Test Environment Control:** Verify NODE_ENV switching works
6. **Remove Old Dependencies:** Clean up custom logging libraries

### Example Migration

```javascript
// Before: Scattered console logging
console.log('Processing user data...');
console.error('Database error:', error);
console.log('User created successfully');

// After: Framework logging
const { getLogger } = require('../modules/logging');
const logger = getLogger('UserProcessor');

logger.info('Processing user data started');
logger.error('Database operation failed', { 
  error: error.message, 
  operation: 'createUser',
  userId: userData.id 
});
logger.success('User created successfully', { 
  userId: newUser.id, 
  email: newUser.email 
});
```

## Performance Considerations

### Console Logging Performance
- **Development:** Minimal overhead, optimized for debugging workflow
- **Production:** Zero overhead (all calls short-circuit)
- **Context Serialization:** JSON.stringify only in development

### Memory Management
- **No Persistent Storage:** Console logs don't accumulate in memory
- **Context Objects:** Passed by reference, not copied
- **String Interpolation:** Performed only when logging is active

## Integration with Framework

### Path Management Integration

```javascript
// Use path manager for log file locations
const { pathManager } = require('../paths');
const logDir = pathManager.get('log_directory');
```

### Configuration Integration

```javascript
// settings.toml can control specialized logging
[logging]
enable_performance_logs = false
enable_security_logs = true
file_logging_enabled = false
```

### Error Handling Integration

```javascript
// Logging integrates with error handling system
const { getLogger } = require('../modules/logging');
const { createStandardError } = require('../modules/errors');

const logger = getLogger('PaymentService');

try {
  processPayment(data);
} catch (error) {
  const standardError = createStandardError('PAYMENT_ERROR', 'Payment processing failed', data, error);
  logger.error(standardError.error.message, standardError.error.context);
  throw standardError;
}
```

## API Reference

### getLogger(component)
**Purpose:** Creates component-tagged logger instance
**Parameters:** 
- `component` (string): Component name for log tagging
**Returns:** Logger object with level methods

### Logger Methods
- `logger.info(message, context?)` - General information
- `logger.success(message, context?)` - Successful operations  
- `logger.error(message, context?)` - Error conditions
- `logger.warn(message, context?)` - Warning conditions
- `logger.debug(message, context?)` - Debug information
- `logger.config(message, context?)` - Configuration events
- `logger.network(message, context?)` - Network operations
- `logger.security(message, context?)` - Security events
- `logger.performance(message, context?)` - Performance metrics

### Context Object Guidelines
- **Structure:** Plain JavaScript object
- **Content:** Relevant debugging information
- **Size:** Keep reasonable for console output
- **Sensitive Data:** Avoid logging passwords, tokens, or PII

This logging system provides the foundation for consistent, efficient debugging across all projects while maintaining the simplicity and anti-overengineering principles of the universal framework.