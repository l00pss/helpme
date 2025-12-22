# HelpMe - Go Utility Library Collection

A comprehensive collection of Go utility libraries designed to simplify common development tasks. This project provides well-tested, production-ready modules for configuration management, logging, functional programming patterns, and query operations.

## 📦 Modules

### 🔧 HAConfig - Configuration Management
A powerful configuration library that supports both YAML files and environment variables with automatic type conversion and validation.

**Features:**
- Load from YAML files and environment variables
- Environment variables override YAML values
- Support for nested structs and pointers
- Custom environment variable prefixes
- Field validation with required tags
- Automatic CamelCase to snake_case conversion
- Support for duration, slices, and all basic types

**Usage:**
```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    *RedisConfig   `yaml:"redis,omitempty"`
}

type ServerConfig struct {
    Host string `yaml:"host" required:"true"`
    Port int    `yaml:"port" required:"true"`
    TLS  bool   `yaml:"tls"`
}

// Load from environment variables only
config := haconfig.New()
var cfg Config
err := config.Load(&cfg)

// Load from YAML with env override
config := haconfig.New(
    haconfig.WithYAMLFile("config.yaml"),
    haconfig.WithEnvPrefix("MYAPP"),
)
err := config.Load(&cfg)

// Quick load functions
err := haconfig.LoadFromFile("config.yaml", &cfg)
err := haconfig.LoadFromEnv(&cfg, haconfig.WithEnvPrefix("MYAPP"))
```

### 📝 O4G Logger - Structured Logging
A feature-rich logging library built on top of logrus with colored output and structured logging capabilities.

**Features:**
- Multiple log levels (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- Colored console output with customizable formatting
- File and console output support
- Structured logging with fields
- Global logger instance for convenience

### 🎯 Option - Optional Values
Rust-inspired Option type for handling nullable values in a type-safe manner.

**Features:**
- `Some[T]` and `None[T]` constructors
- `IsSome()` and `IsNone()` checks
- Safe value extraction methods
- Functional programming patterns

### ✅ Result - Error Handling
Rust-inspired Result type for explicit error handling without exceptions.

**Features:**
- `Ok[T]` and `Err[T]` constructors
- `IsOk()` and `IsErr()` checks
- Safe value and error extraction
- Chainable operations

### 🔍 Wrapper - Query Operations
A query builder and wrapper system for database operations with support for projections, pagination, sorting, and filtering.

**Features:**
- Type-safe query building
- Projection support
- Pagination capabilities
- Sorting and filtering
- Context-aware operations

## 🚀 Installation

Since this is a multi-module project, you can install specific modules:

```bash
# Install the configuration module
go get github.com/l00pss/helpme/haconfig

# Install the logger module
go get github.com/l00pss/helpme/o4g_logger

# Install all modules (if using go.work)
go get github.com/l00pss/helpme
```

## 📋 Requirements

- Go 1.25 or later
- Dependencies are managed per module

## 🧪 Testing

Each module includes comprehensive tests. Run tests for all modules:

```bash
# Test all modules
go test ./...

# Test specific module
cd haconfig && go test
cd o4g_logger && go test
cd option && go test
cd result && go test
cd wrapper && go test

# Run tests with coverage
go test -cover ./...
```

## 📁 Project Structure

```
helpme/
├── haconfig/          # Configuration management
├── o4g_logger/        # Structured logging
├── option/            # Optional value types
├── result/            # Result/Error types
├── wrapper/           # Query operations
├── go.mod             # Main module
├── go.work            # Workspace configuration
└── README.md          # This file
```

## 🔧 Environment Variable Configuration (HAConfig)

The HAConfig module automatically converts struct field names to environment variable names:

- `ServerHost` → `SERVER_HOST`
- `DatabaseURL` → `DATABASE_URL`
- `AllowedIPs` → `ALLOWED_I_PS`

With prefix "MYAPP":
- `ServerHost` → `MYAPP_SERVER_HOST`

Nested structs:
- `Server.Host` → `SERVER_HOST`
- `Database.Credentials.Username` → `DATABASE_CREDENTIALS_USERNAME`

## 💡 Examples

### Configuration Loading
```go
// config.yaml
server:
  host: localhost
  port: 8080
database:
  url: postgres://localhost/mydb

// Environment variables (override YAML)
export SERVER_HOST=production-server
export SERVER_PORT=443

// Go code
var config MyConfig
err := haconfig.LoadFromFile("config.yaml", &config)
// Result: host=production-server, port=443, url=postgres://localhost/mydb
```

### Optional Values
```go
func findUser(id int) option.Option[User] {
    if user := database.Find(id); user != nil {
        return option.Some(*user)
    }
    return option.None[User]()
}

user := findUser(123)
if user.IsSome() {
    fmt.Printf("Found user: %s", user.Unwrap().Name)
}
```

### Result Error Handling
```go
func divide(a, b float64) result.Result[float64] {
    if b == 0 {
        return result.Err[float64](errors.New("division by zero"))
    }
    return result.Ok(a / b)
}

res := divide(10, 2)
if res.IsOk() {
    fmt.Printf("Result: %f", res.Unwrap())
} else {
    fmt.Printf("Error: %s", res.UnwrapErr())
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🏗️ Architecture

This project uses Go workspaces to manage multiple related modules in a single repository. Each module is independently versioned and can be used separately or together.

- **Modular Design**: Each utility is a separate module
- **Type Safety**: Extensive use of Go generics for type-safe APIs
- **Testing**: Comprehensive test coverage for all modules
- **Performance**: Benchmarks included for critical paths
- **Documentation**: Well-documented APIs with examples
