# tomv - TOML Variables Library

**Import:** `github.com/DeprecatedLuar/toml-vars-letsgooo`

## Core API

```go
import "github.com/DeprecatedLuar/toml-vars-letsgooo"

// Required values (panics if missing)
port := tomv.GetInt("server.port")
host := tomv.Get("database.host")
enabled := tomv.GetBool("features.login")

// Safe values with defaults (never panics)
timeout := tomv.GetIntOr("api.timeout", 30)
debug := tomv.GetBoolOr("app.debug", false)
name := tomv.GetOr("app.name", "myapp")

// Utility
exists := tomv.Exists("config.key")
```

## TOML Variable Substitution

```toml
# config.toml
[database]
host = "localhost"
port = 5432
name = "myapp"

[connection]
url = "postgres://{{database.host}}:{{database.port}}/{{database.name}}"
backup = "{{connection.url}}_backup"

[paths]
base = "/app"
uploads = "{{paths.base}}/uploads"
```

## Environment Variable Integration

```toml
# Explicit env var references with defaults
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"

[database]
url = "{{ENV.DATABASE_URL:-postgres://localhost:5432/app}}"
password = "{{ENV.DB_PASSWORD}}"  # Required env var (no default)
```

## File Discovery

- **Auto-discovery**: Finds all `*.toml` files in project recursively
- **Zero config**: No setup required - just import and use
- **Smart caching**: Reloads files only when changed

## Type Functions

| Function | Returns | Example |
|----------|---------|---------|
| `Get(key)` | string | `tomv.Get("app.name")` |
| `GetInt(key)` | int | `tomv.GetInt("server.port")` |
| `GetBool(key)` | bool | `tomv.GetBool("features.debug")` |
| `GetOr(key, def)` | string | `tomv.GetOr("app.env", "dev")` |
| `GetIntOr(key, def)` | int | `tomv.GetIntOr("timeout", 30)` |
| `GetBoolOr(key, def)` | bool | `tomv.GetBoolOr("cache", true)` |
| `Exists(key)` | bool | `tomv.Exists("optional.config")` |

## Key Syntax

- **Basic**: `section.key` → `[section] key = "value"`
- **Nested**: `app.database.host` → `[app.database] host = "localhost"`
- **Variables**: `{{section.key}}` → References other TOML values
- **Environment**: `{{ENV.VAR_NAME:-default}}` → Environment variables with defaults

## Behavior

- **Required config**: Use `Get*()` - panics if missing (fail-fast)
- **Optional config**: Use `Get*Or()` - returns default if missing
- **Environment override**: Env vars always override TOML values
- **Variable resolution**: Internal `{{}}` references resolved automatically
- **File monitoring**: Configuration always current (auto-reload on change)

## Error Patterns

```go
// Fail-fast for critical config
dbURL := tomv.Get("database.url")  // Must exist

// Safe defaults for optional config
maxRetries := tomv.GetIntOr("api.retries", 3)

// Check before using
if tomv.Exists("features.experimental") {
    experimental := tomv.GetBool("features.experimental")
}
```