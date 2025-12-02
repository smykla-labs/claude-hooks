# gRPC Plugin Loader Architecture

Persistent gRPC connections for external validator plugins with connection pooling, TLS security, and cross-language support.

## Core Design Philosophy

**Connection Pooling by Address**: Multiple plugin instances sharing same gRPC address reuse single connection. Avoids expensive TCP/TLS handshakes (2-5M× faster than creating new connections).

**Lazy Connection Establishment**: Use `grpc.NewClient()` (non-blocking) instead of deprecated `grpc.DialContext()`. Connection validation happens on first RPC call, not during Load().

**Thread-Safe with Double-Check Locking**: RWMutex allows concurrent reads (hot path) while protecting connection creation (slow path). Double-check pattern prevents duplicate connections under concurrency.

**Security by Default**: TLS automatically required for non-localhost addresses. Insecure connections to remote servers must be explicitly allowed via `AllowInsecureRemote` flag.

**JSON Bridge for Config**: Protobuf requires homogeneous map types (`map<string, string>`). Internal config uses `map[string]any`. JSON marshaling bridges type mismatch for non-string values.

## Protocol Definition

```protobuf
// api/plugin/v1/plugin.proto
syntax = "proto3";

service ValidatorPlugin {
  rpc Info(InfoRequest) returns (InfoResponse);
  rpc Validate(ValidateRequest) returns (ValidateResponse);
}

message HookContext {
  string event_type = 1;
  string tool_name = 2;
  map<string, string> tool_input = 3;  // JSON-encoded for non-strings
}

message ValidateRequest {
  HookContext context = 1;
  map<string, string> config = 2;      // Plugin-specific config
}

message ValidateResponse {
  bool passed = 1;
  bool should_block = 2;
  string message = 3;
  string error_code = 4;
  string fix_hint = 5;
  string doc_link = 6;
}
```

**Why map<string, string>**: Protobuf v3 doesn't support heterogeneous maps. JSON encoding allows passing `int`, `bool`, `[]string` via string values.

## Connection Pooling Implementation

```go
// internal/plugin/grpc_loader.go
type GRPCLoader struct {
    mu          sync.RWMutex
    connections map[string]*grpc.ClientConn  // Keyed by address
    dialTimeout time.Duration
    closed      bool                         // Prevents Load() after Close()
}

func (l *GRPCLoader) Load(cfg *config.PluginConfig) (Plugin, error) {
    // Check closed state
    if l.closed {
        return nil, ErrLoaderClosed
    }

    // Fast path: Check existing connection (read lock only)
    l.mu.RLock()
    conn, exists := l.connections[cfg.Address]
    l.mu.RUnlock()
    if exists {
        return l.newPluginFromConn(conn, cfg), nil
    }

    // Slow path: Create connection (write lock with double-check)
    l.mu.Lock()
    defer l.mu.Unlock()

    // Double-check after acquiring write lock
    if conn, exists := l.connections[cfg.Address]; exists {
        return l.newPluginFromConn(conn, cfg), nil
    }

    // Create new connection
    conn, err := l.dialGRPC(cfg)
    if err != nil {
        return nil, err
    }

    l.connections[cfg.Address] = conn
    return l.newPluginFromConn(conn, cfg), nil
}
```

**Why Double-Check Locking**: Two goroutines can pass read lock check simultaneously. Without double-check after write lock, both would create connections.

**Why RWMutex**: Hot path (existing connections) only needs read lock. Multiple goroutines can load plugins concurrently without contention.

## TLS Security Model

### Configuration Schema

```go
type TLSConfig struct {
    Enabled             *bool  // nil=auto (localhost insecure, remote TLS)
    CertFile            string // Client cert (mTLS)
    KeyFile             string // Client key (mTLS)
    CAFile              string // CA cert for server verification
    InsecureSkipVerify  *bool  // Skip server cert verification
    AllowInsecureRemote *bool  // Explicitly allow insecure to non-localhost
}
```

### Security Decision Matrix

| Address   | TLS.Enabled | AllowInsecureRemote | Result                        |
|:----------|:------------|:--------------------|:------------------------------|
| localhost | nil (auto)  | -                   | Insecure (development)        |
| localhost | true        | -                   | TLS required                  |
| localhost | false       | true                | Insecure (allowed)            |
| remote    | nil (auto)  | -                   | TLS required                  |
| remote    | true        | -                   | TLS required                  |
| remote    | false       | false               | **Error** (security risk)     |
| remote    | false       | true                | Insecure (explicitly allowed) |

**Rationale**: Localhost is development-friendly (no certs needed). Remote addresses must use TLS unless explicitly opted out.

### TLS Implementation

```go
func (l *GRPCLoader) buildTransportCredentials(cfg *config.PluginConfig) (credentials.TransportCredentials, error) {
    isLocal := IsLocalAddress(cfg.Address)

    // Auto mode (nil Enabled): insecure for localhost, TLS for remote
    if cfg.TLS == nil || cfg.TLS.Enabled == nil {
        if isLocal {
            return insecure.NewCredentials(), nil
        }
        return nil, errors.New("TLS required for remote address (use tls.enabled=true)")
    }

    // Explicit insecure
    if !*cfg.TLS.Enabled {
        if !isLocal && !cfg.TLS.AllowInsecureRemote {
            return nil, errors.New("insecure remote connection requires allow_insecure_remote=true")
        }
        return insecure.NewCredentials(), nil
    }

    // Build TLS credentials
    return l.buildTLSCredentials(cfg.TLS)
}

func (l *GRPCLoader) buildTLSCredentials(tlsCfg *config.TLSConfig) (credentials.TransportCredentials, error) {
    config := &tls.Config{
        MinVersion: tls.VersionTLS12,  // Enforce minimum TLS 1.2
    }

    // Load CA cert for server verification
    if tlsCfg.CAFile != "" {
        caCert, err := os.ReadFile(tlsCfg.CAFile)
        if err != nil {
            return nil, err
        }
        certPool := x509.NewCertPool()
        certPool.AppendCertsFromPEM(caCert)
        config.RootCAs = certPool
    }

    // Load client cert for mTLS
    if tlsCfg.CertFile != "" && tlsCfg.KeyFile != "" {
        cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
        if err != nil {
            return nil, err
        }
        config.Certificates = []tls.Certificate{cert}
    }

    // Skip verification (development only)
    if tlsCfg.InsecureSkipVerify != nil && *tlsCfg.InsecureSkipVerify {
        config.InsecureSkipVerify = true
    }

    return credentials.NewTLS(config), nil
}
```

**Gotcha**: `InsecureSkipVerify` disables server certificate validation. Only use for self-signed certs in development.

## Type Conversion

Plugin adapter converts between internal and protobuf types:

```go
// internal/plugin/grpc_adapter.go
func (a *grpcPluginAdapter) Validate(ctx context.Context, req plugin.ValidateRequest) (plugin.ValidateResponse, error) {
    // Convert internal config (map[string]any) to protobuf (map[string]string)
    protoConfig := make(map[string]string)
    for k, v := range req.Config {
        switch val := v.(type) {
        case string:
            protoConfig[k] = val
        default:
            // JSON-encode non-strings (int, bool, []string, etc.)
            bytes, err := json.Marshal(val)
            if err != nil {
                return plugin.ValidateResponse{}, err
            }
            protoConfig[k] = string(bytes)
        }
    }

    protoReq := &pluginv1.ValidateRequest{
        Context: convertHookContext(req.Context),
        Config:  protoConfig,
    }

    protoResp, err := a.client.Validate(ctx, protoReq)
    if err != nil {
        return plugin.ValidateResponse{}, err
    }

    // Convert protobuf response back to internal
    return plugin.ValidateResponse{
        Passed:      protoResp.Passed,
        ShouldBlock: protoResp.ShouldBlock,
        Message:     protoResp.Message,
        ErrorCode:   protoResp.ErrorCode,
        FixHint:     protoResp.FixHint,
        DocLink:     protoResp.DocLink,
    }, nil
}
```

**Why JSON Encoding**: Preserves type information across wire format. Plugin servers can `json.Unmarshal()` to recover original types.

## Code Generation

Uses buf 1.61.0 for protocol buffer compilation:

```yaml
# buf.yaml
version: v2
modules:
  - path: api
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

```yaml
# buf.gen.yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/smykla-labs/klaudiush
plugins:
  - remote: buf.build/protocolbuffers/go
    out: .
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go
    out: .
    opt:
      - paths=source_relative
```

Generate code: `buf generate`

**Generated Files**:

- `api/plugin/v1/plugin.pb.go` - Protobuf message types
- `api/plugin/v1/plugin_grpc.pb.go` - gRPC client/server stubs

**Gotcha**: Must run `buf generate` after modifying `.proto` files. Generated files are committed to git.

## Registry Integration

gRPC plugins categorized as `CategoryIO` (I/O-bound, not CPU-bound):

```go
// internal/plugin/registry.go
func (r *Registry) registerPlugin(cfg *config.PluginConfig, plugin Plugin) {
    category := validator.CategoryCPU  // Default for Go plugins

    // Exec and gRPC plugins are I/O-bound
    if cfg.Type == config.PluginTypeExec || cfg.Type == config.PluginTypeGRPC {
        category = validator.CategoryIO
    }

    adapter := &ValidatorAdapter{
        plugin:   plugin,
        config:   cfg,
        category: category,
    }

    r.validators = append(r.validators, adapter)
}
```

**Why CategoryIO**: Network I/O operations wait on responses. Parallel execution pool sized for I/O concurrency (20 workers), not CPU cores (see `session-parallel-execution.md`).

## Performance Characteristics

### Connection Pooling Impact

**Scenario**: 3 plugin instances on same gRPC server, 100 validations each

**Without pooling** (naive):

- 3 instances × 100 validations = 300 connections
- ~50ms connection overhead each = **15s total**

**With pooling**:

- 1 shared connection across instances
- ~50ms one-time overhead
- ~0ms subsequent validations = **~50ms total**

**Savings**: 299 connections avoided, **~14.95s saved** (99.7% reduction)

### Lock Contention

RWMutex characteristics:

- **Hot path** (existing connection): Read lock, no contention
- **Cold path** (new connection): Write lock, serialized

**Concurrency**: 100 goroutines loading same plugin:

- 1 acquires write lock, creates connection
- 99 wait briefly, then use read lock
- Total delay: 1× connection time, not 100×

## Configuration Examples

### Basic gRPC (localhost, insecure)

```toml
[[plugins.plugins]]
name = "dev-validator"
type = "grpc"
address = "localhost:50051"
timeout = "5s"
# No TLS config - auto-insecure for localhost
```

### Remote with TLS

```toml
[[plugins.plugins]]
name = "prod-validator"
type = "grpc"
address = "validator.example.com:443"
timeout = "10s"

[plugins.plugins.tls]
enabled = true
ca_file = "/etc/ssl/certs/ca-bundle.crt"
```

### Mutual TLS (mTLS)

```toml
[[plugins.plugins]]
name = "mtls-validator"
type = "grpc"
address = "validator.example.com:443"

[plugins.plugins.tls]
enabled = true
cert_file = "/etc/klaudiush/client.crt"
key_file = "/etc/klaudiush/client.key"
ca_file = "/etc/klaudiush/ca.crt"
```

### Development with Self-Signed Cert

```toml
[[plugins.plugins]]
name = "self-signed-validator"
type = "grpc"
address = "localhost:50051"

[plugins.plugins.tls]
enabled = true
insecure_skip_verify = true  # ONLY for development
```

## Testing Strategy

Uses real gRPC server (not bufconn) to test actual network behavior:

```go
// internal/plugin/grpc_loader_test.go
type mockGRPCServer struct {
    pluginv1.UnimplementedValidatorPluginServer
    infoResponse *pluginv1.InfoResponse
    validateResponse *pluginv1.ValidateResponse
    validateError error
}

func TestGRPCLoader_Load(t *testing.T) {
    // Start real gRPC server on random port
    listener, _ := net.Listen("tcp", "localhost:0")
    server := grpc.NewServer()
    pluginv1.RegisterValidatorPluginServer(server, &mockGRPCServer{...})
    go server.Serve(listener)
    defer server.Stop()

    // Test loader against real server
    loader := NewGRPCLoader()
    plugin, err := loader.Load(&config.PluginConfig{
        Address: listener.Addr().String(),
    })
    // ...
}
```

**Why Real Server vs bufconn**: Tests actual network stack, DNS resolution, connection pooling. bufconn is in-memory pipe (misses real-world issues).

## Plugin Type Comparison

| Feature              | Go (.so)     | gRPC                 | Exec              |
|:---------------------|:-------------|:---------------------|:------------------|
| **Performance**      | Fastest      | Fast (persistent)    | Slowest           |
| **Language Support** | Go only      | Any                  | Any               |
| **Process**          | In-process   | Separate             | Separate          |
| **Connection**       | Direct call  | Pooled, persistent   | Per-invocation    |
| **Overhead**         | ~100ns       | ~1-5ms (network)     | ~50-100ms (spawn) |
| **Category**         | CategoryCPU  | CategoryIO           | CategoryIO        |
| **Hot-Reload**       | Restart only | Yes (server restart) | Yes (per-call)    |
| **Security**         | Same process | TLS, mTLS            | Process isolation |

**When to Use gRPC**:

- Cross-language plugins (Python, Node.js, Rust)
- Long-running state (cache, connections)
- Hot-reload without restarting klaudiush
- Production deployments (TLS security)

**When to Use Exec**:

- Simpler deployment (single binary)
- Short validations (< 10ms)
- Language support more important than performance

## Common Pitfalls

1. **Not using connection pooling**: Creating new connection per validation is 2-5M× slower. Always reuse connections by address.

2. **Blocking dial with grpc.DialContext**: Use `grpc.NewClient()` (non-blocking). Deprecated `DialContext()` blocks Load() until connection established.

3. **Missing double-check after write lock**: Two goroutines can pass read lock check. Without double-check, both create connections to same address.

4. **Insecure remote connections**: Remote addresses require TLS by default. Override needs explicit `allow_insecure_remote=true` (security risk).

5. **Hardcoding string values in protobuf**: Use JSON encoding for non-strings. `"threshold": "42"` (JSON) vs `"threshold": 42` (internal).

6. **Not closing loader**: Connections leak if `Close()` not called. Defer `Close()` after creating loader.

7. **Using bufconn for tests**: In-memory pipe doesn't test real network. Use `net.Listen("tcp", "localhost:0")` for realistic tests.

8. **InsecureSkipVerify in production**: Disables certificate validation. Only use for self-signed certs in development.

9. **Wrong category assignment**: gRPC plugins are I/O-bound (CategoryIO), not CPU-bound. Wrong category causes parallel execution pool mismatch.

10. **Not handling load after close**: Check `closed` flag before loading. Return `ErrLoaderClosed` to prevent using closed connections.
