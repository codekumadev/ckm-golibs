# ckm-golibs

Reusable Go utilities for dependency injection, database connectivity, and structured logging. The tables below provide a quick reference for each package and its exported types and functions.

## Installation

- Root module: `go get github.com/lu69x/ckm-golibs@latest`
- Packages can be imported individually, for example:
  - `github.com/lu69x/ckm-golibs/log`
  - `github.com/lu69x/ckm-golibs/database`
  - `github.com/lu69x/ckm-golibs/containers`

Go version: 1.23+

---

## Package `containers`

Lightweight, reflection-based dependency injection container with support for named registrations, singletons, value providers, resolution, and struct field injection via tags.

| Item | Kind | Description |
| --- | --- | --- |
| `Container` | Type | Map-backed container keyed by `reflect.Type` and optional name. |
| `New()` | Function | Creates a fresh container instance. |
| `Register(resolver any)` | Method | Registers a factory function; returns a new instance per resolve. |
| `NamedRegister(name string, resolver any)` | Method | Same as `Register` but under a name. |
| `Singleton(resolver any)` | Method | Registers a factory executed once; the instance is reused. |
| `NamedSingleton(name string, resolver any)` | Method | Named variant of `Singleton`. |
| `Value(resolver any)` | Method | Registers an already-constructed instance (value singleton). |
| `NamedValue(name string, resolver any)` | Method | Named variant of `Value`. |
| `Resolve(ptr any) error` | Method | Resolves into the provided pointer; returns error if missing. |
| `NamedResolve(ptr any, name string) error` | Method | Resolves a named binding. |
| `MustResolve(ptr any)` | Method | Like `Resolve` but panics on error. |
| `MustNameResolve(ptr any, name string)` | Method | Like `NamedResolve` but panics on error. |
| `Inject(structPtr any) error` | Method | Populates struct fields using `inject:"<name>"` tag. |
| `Reset()` | Method | Clears all registrations. |

Global helpers mirror the instance API: `containers.Register`, `Singleton`, `Value`, `Resolve`, `NamedResolve`, `MustResolve`, `MustNamedResolve`, `Inject`, and `Reset`.

### Field injection

- Add a struct tag `inject:"name"` (empty string for default) on fields with concrete interface/type you registered in the container.
- `Inject(&myStruct)` will set those fields using unsafe-aware assignment to bypass unexported field restrictions where applicable.

### Containers usage example

```go
package main

import (
    "fmt"
    "github.com/lu69x/ckm-golibs/containers"
)

type Service interface{ Ping() string }
type svcImpl struct{ id string }
func (s *svcImpl) Ping() string { return "pong:" + s.id }

type Handler struct {
    SvcDefault Service  `inject:""`
    SvcBlue    Service  `inject:"blue"`
}

func main() {
    c := containers.New()

    c.Singleton(func() Service { return &svcImpl{id: "default"} })
    c.NamedSingleton("blue", func() Service { return &svcImpl{id: "blue"} })

    // Resolve
    var s Service
    if err := c.Resolve(&s); err != nil { panic(err) }
    fmt.Println(s.Ping()) // "pong:default"

    // Inject by tag
    h := &Handler{}
    if err := c.Inject(h); err != nil { panic(err) }
    fmt.Println(h.SvcDefault.Ping(), h.SvcBlue.Ping())
}
```

Notes:
- `Singleton` is concurrency-safe and initializes the instance once.
- Resolver functions may optionally return `(T, error)`; the error is propagated.
- `Resolve` expects a pointer to the abstraction type (concrete/interface).

---

## Package `database`

| Item | Kind | Description |
| --- | --- | --- |
| `IDatabase` | Interface | Abstraction over a database provider that exposes the shared GORM connection. |
| `SqlDbConfig` | Struct | Configuration required to connect to a SQL database. |
| `SqlDatabase` | Struct | Concrete implementation of `IDatabase` that manages the singleton connection. |
| `NewDatabase(conf SqlDbConfig, engine string)` | Function | Lazily constructs the shared connection using the requested engine. |
| `(*SqlDatabase) Connect()` | Method | Returns the singleton `*gorm.DB` connection. |

### `SqlDbConfig` fields

| Field | Type | Purpose |
| --- | --- | --- |
| `Host` | `string` | Hostname or address of the database server. |
| `Port` | `int` | TCP port for the database server. |
| `User` | `string` | Username used for authentication. |
| `Password` | `string` | Password used for authentication. |
| `DBName` | `string` | Database name to connect to. |
| `SSLMode` | `bool` | Toggles SSL mode in the generated DSN. |
| `Schema` | `string` | Schema (search path) selected after connecting. |

### Connection lifecycle

1. `NewDatabase` receives a `SqlDbConfig` and an engine name (`postgres`, `mysql`, or `sqlite`).
2. A `sync.Once` guard ensures the connection is initialized only once per process.
3. The function builds a driver-specific DSN and opens the connection with the matching GORM driver.
4. The constructed `*gorm.DB` is cached and returned via the `IDatabase` interface.
5. `Connect` retrieves that shared connection for callers.

## Package `log`

| Item | Kind | Description |
| --- | --- | --- |
| `Level` | Type alias | Mirrors `zapcore.Level` for consumers of the package. |
| `Field` | Type alias | Mirrors `zap.Field` to build structured log fields. |
| `Option` | Type alias | Mirrors `zap.Option` and exposes helper constructors. |
| Level constants | Constants | `DebugLevel`, `InfoLevel`, `WarnLevel`, `ErrorLevel`, `DPanicLevel`, `PanicLevel`, `FatalLevel`. |
| Field helpers | Variables | Re-exported helpers like `String`, `Int`, `Time`, etc. |
| Option helpers | Functions | `WithCaller`, `AddStacktrace`, `AddCallerSkip`, `AddCaller`. |
| `Logger` | Struct | Wraps a configured `*zap.Logger` and the chosen level. |
| Logging methods | Methods | `Debug`, `Info`, `Warn`, `Error`, `DPanic`, `Panic`, `Fatal`. |
| Context helpers | Methods | `WithFields`, `WithOptions` for deriving child loggers. |
| Constructors | Functions | `New`, `NewFile` create standard error or file-based loggers. |
| Internals | Functions | `generateUniqueLogFileName`, `getLogWriter`, `zapLogLevel`, `zapEncoder`, `new`. |
| `(*Logger) Sync()` | Method | Flushes buffered logs by delegating to zap. |

### Logger constructors

| Function | Output | Notes |
| --- | --- | --- |
| `New(logLevel, logFormat string)` | `*Logger` | Emits to `os.Stderr`, configures encoder, level, and caller metadata. |
| `NewFile(logLevel, logFormat, logPrefix string)` | `*Logger` | Writes to a uniquely named file generated from the prefix, timestamp, and UUID. |

Both constructors rely on `zapLogLevel` (string → level), `zapEncoder` (format → encoder), and `new` (core builder) to prepare the underlying zap logger. `NewFile` additionally uses `getLogWriter`, which opens the generated file and wraps it in a Zap `WriteSyncer`.

### Using `Logger`

| Method | Purpose |
| --- | --- |
| `Debug` / `Info` / `Warn` / `Error` / `DPanic` / `Panic` / `Fatal` | Emit a log entry at the matching severity with optional structured fields. |
| `WithFields(fields ...Field)` | Returns a child logger that always includes the provided fields. |
| `WithOptions(opts ...Option)` | Returns a child logger with extra zap options (e.g., caller or stack trace). |
| `Sync()` | Flushes buffered log entries. |

### Field helpers

Field helpers re-export zap constructors like `String`, `Int`, `Bool`, `Duration`, `Time`, and others so callers can describe structured context without importing zap directly.

## Usage example

```go
package main

import (
    "github.com/lu69x/ckm-golibs/database"
    clog "github.com/lu69x/ckm-golibs/log"
)

func main() {
    dbConf := database.SqlDbConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "secret",
        DBName:   "app",
        SSLMode:  false,
        Schema:   "public",
    }

    db := database.NewDatabase(dbConf, "postgres").Connect()
    _ = db // use the *gorm.DB connection

    logger := clog.New("info", "json")
    logger.Info("application started", clog.String("module", "main"))
}
```

### Logging to a file

```go
logger := clog.NewFile("debug", "console", "app")
logger.Debug("file logging enabled")
defer logger.Sync()
```
