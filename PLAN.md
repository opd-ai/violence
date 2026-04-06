# Rust → Pure Go Migration Plan

## Executive Summary

This codebase contains **no Rust source code** (zero `.rs` files, no `Cargo.toml`, no `build.rs`). The sole Rust dependency is **wasmer-go** (`github.com/wasmerio/wasmer-go v1.0.4`), a Go binding to the Wasmer WebAssembly runtime written in Rust. Wasmer is consumed as a pre-compiled native library distributed inside the Go module — developers never invoke `cargo` or `rustc`.

The wasmer-go library is used exclusively in **`pkg/mod/wasm_loader.go`** to provide a sandboxed WASM execution environment for untrusted game mods. Platform stubs already exist for Windows (`wasm_loader_windows.go`) and JS/WASM (`wasm_loader_js.go`) where wasmer-go cannot compile.

**Migration feasibility: HIGH.** Replacing wasmer-go with **wazero** (`github.com/tetratelabs/wazero`), a pure-Go WebAssembly runtime with zero CGo dependencies, is a drop-in substitution that eliminates the Rust toolchain dependency entirely while preserving (and in some cases improving) cross-platform support. The wazero project is mature, actively maintained, and already used in production by major Go projects. It supports the same WASM 1.0 spec targeted by wasmer-go, plus WASI preview 1 for future file-system sandboxing.

**Performance outlook:** Wazero's interpreter is ~1.5–3× slower than Wasmer's Cranelift JIT for raw WASM execution in CPU-bound tight loops. However, the mod system's workload profile — infrequent event-hook invocations (`weapon.fire`, `enemy.spawn`, `level.generate`) executing a few hundred WASM instructions each — means mod hooks currently consume **<0.1% of the 16.6ms frame budget**. Even a 3× increase in raw WASM execution time translates to <0.3% of frame budget, well within the **≤10% total frame-time impact** threshold. A synthetic microbenchmark regression of ~1.5–3× on tight-loop WASM kernels is expected and explicitly accepted, since it does not produce a user-observable frame-time regression.

---

## Component Inventory

| Component | File(s) | Purpose | Rust Dependency |
|-----------|---------|---------|-----------------|
| **WASM Loader (core)** | `pkg/mod/wasm_loader.go` | Loads, compiles, instantiates WASM modules; manages module lifecycle; exposes host functions to mods | `wasmer-go/wasmer` — Engine, Store, Module, Instance, Memory, Function, ImportObject, ValueTypes |
| **WASM Config** | `pkg/mod/wasm_loader.go:16-43` | Security constraints: 64 MB memory limit, 1 billion instruction fuel limit, file-path sandbox | Uses wasmer-go store/engine for enforcement |
| **Host Function Bridge** | `pkg/mod/wasm_loader.go:266-291` | `log_message(ptr, len)` — only implemented host function; registered in `env` namespace via `wasmer.NewFunction` | `wasmer.NewFunction`, `wasmer.NewFunctionType`, `wasmer.NewValueTypes`, `wasmer.Value` |
| **Module Call Interface** | `pkg/mod/wasm_loader.go:294-307` | `WASMModule.Call()` — invokes exported WASM functions by name with variadic args | `instance.Exports.GetFunction` returns `wasmer.NativeFunction` |
| **Export Introspection** | `pkg/mod/wasm_loader.go:310-313` | `WASMModule.HasExport()` — checks if a WASM module exports a named function | `instance.Exports.GetFunction` |
| **Windows Stub** | `pkg/mod/wasm_loader_windows.go` | No-op implementation returning `ErrNotSupported` for all WASM operations | None (build-tag excluded) |
| **JS/WASM Stub** | `pkg/mod/wasm_loader_js.go` | No-op implementation returning `ErrNotSupported` for all WASM operations | None (build-tag excluded) |
| **Mod Loader Integration** | `pkg/mod/mod.go:36-61` | `Loader` struct holds `*WASMLoader`; `NewLoader()` calls `NewWASMLoader()` | Indirect — constructs wasmer-backed loader |
| **go.mod Dependency** | `go.mod` | `github.com/wasmerio/wasmer-go v1.0.4` | Direct module dependency |
| **go.sum Checksum** | `go.sum` | Integrity hash for wasmer-go module | Checksum entry |

---

## GC Risk Assessment

| Component | GC Risk | Rationale |
|-----------|---------|-----------|
| **WASM Loader (core)** | **Low** | The wazero replacement stores `CompiledModule` and `Runtime` as long-lived objects pinned to the `WASMLoader` struct. No per-frame allocation. Module compilation is a one-time cost at mod load. |
| **WASM Config** | **None** | Pure value-type struct (`uint32`, `uint64`, `bool`, `[]string`). No change from current implementation. |
| **Host Function Bridge** | **Low** | Host function closures are allocated once at module instantiation. The `log_message` callback creates a small `logrus.Fields` map per call — identical to current behavior. Wazero passes primitive `uint64` args on the stack, not heap. |
| **Module Call Interface** | **Low** | Wazero's `api.Function.Call()` accepts `uint64` variadic args. No boxing of primitives into `interface{}` (unlike wasmer-go's `[]wasmer.Value`). This is a **GC improvement** over wasmer-go. |
| **Export Introspection** | **None** | `module.ExportedFunctions()` returns a map built at compile time. Single map lookup, no allocation. |
| **Windows Stub** | **None** | Stubs return constant errors. With wazero, these stubs become **unnecessary** — wazero compiles on all platforms including Windows and JS/WASM without CGo. |
| **JS/WASM Stub** | **None → Eliminated** | Wazero supports `GOOS=js GOARCH=wasm` targets. The stub file can be removed entirely, enabling mod loading in the browser build. |
| **Mod Loader Integration** | **None** | Only changes are type names in struct fields. No allocation pattern change. |

---

## Mitigation Strategies

| Concern | Mitigation | Go Idiom |
|---------|------------|----------|
| **WASM module compilation allocations** | Compile modules once at load time; cache `wazero.CompiledModule` in `WASMLoader.modules` map — identical to current `wasmer.Module` caching | Long-lived objects, amortized allocation |
| **Host function closure captures** | Pre-allocate `logrus.Fields` map in closure factory; reuse via `sync.Pool` if call frequency warrants it | `sync.Pool` for ephemeral maps; or pre-allocated `logrus.Entry` per module |
| **Variadic call args** | Wazero's `api.Function.Call(ctx, ...uint64)` passes raw `uint64` values — no boxing. This **eliminates** the `[]wasmer.Value` heap allocation present in wasmer-go | Stack-allocated value types (`uint64`) |
| **Memory read/write for string passing** | Use `wazero.Memory.Read()` which returns a `[]byte` slice backed by the WASM linear memory buffer — zero-copy. For read-only string access without allocation, use `unsafe.String(&bytes[0], len(bytes))` (Go 1.20+). **Constraint**: the resulting string is only valid while the `[]byte` from `Memory.Read()` is not modified or garbage-collected — do not retain past the host function call boundary. | `unsafe.String(&bytes[0], len(bytes))` for zero-copy string reads within host function scope |
| **Context cancellation for fuel/timeout** | Wazero supports `context.WithTimeout` and `context.WithCancel` natively — replaces wasmer-go's fuel metering with standard Go context deadlines | `context.WithTimeout(ctx, 100*time.Millisecond)` per mod call |
| **Memory limits** | Use `wazero.RuntimeConfig.WithMemoryLimitPages()` to cap WASM linear memory. 64 MB = 1024 pages (64 KiB each) | `wazero.NewRuntimeConfig().WithMemoryLimitPages(1024)` |
| **Cross-platform compilation** | Wazero is pure Go with zero CGo. Eliminates POSIX `open_memstream` requirement. Windows and JS stubs become unnecessary. | Remove build tags; single implementation file |

---

## FFI Boundary Analysis

### Current FFI Chain

```
Go application code
  └─→ pkg/mod/wasm_loader.go (Go)
        └─→ wasmer-go (Go bindings)
              └─→ libwasmer (C ABI, compiled from Rust)
                    └─→ Cranelift JIT (Rust)
```

**FFI boundary count: 2** (Go→CGo, CGo→Rust)

### Post-Migration FFI Chain

```
Go application code
  └─→ pkg/mod/wasm_loader.go (Go)
        └─→ wazero (pure Go)
              └─→ Go compiler-generated native code
```

**FFI boundary count: 0**

### FFI Elimination Steps

1. Replace all `wasmer.*` type references with `wazero` equivalents (see API mapping table below)
2. Remove `//go:build !js && !windows` constraint from `wasm_loader.go`
3. Delete `wasm_loader_windows.go` and `wasm_loader_js.go` stub files
4. Remove `github.com/wasmerio/wasmer-go` from `go.mod`
5. Run `go mod tidy` to clean transitive dependencies

### API Mapping: wasmer-go → wazero

| wasmer-go | wazero | Notes |
|-----------|--------|-------|
| `wasmer.NewEngine()` | `wazero.NewRuntime(ctx)` | Runtime encapsulates engine + store |
| `wasmer.NewStore(engine)` | *(embedded in Runtime)* | No separate store concept |
| `wasmer.NewModule(store, bytes)` | `runtime.CompileModule(ctx, bytes)` | Returns `CompiledModule` |
| `wasmer.NewInstance(module, imports)` | `runtime.InstantiateModule(ctx, compiled, config)` | Returns `api.Module` |
| `wasmer.NewImportObject()` | `runtime.NewHostModuleBuilder("env")` | Builder pattern |
| `wasmer.NewFunction(store, type, fn)` | `builder.NewFunctionBuilder().WithFunc(fn).Export("name")` | Fluent API |
| `wasmer.NewFunctionType(params, results)` | *(inferred from Go function signature)* | No manual type construction |
| `wasmer.NewValueTypes(wasmer.I32)` | *(inferred)* | Wazero uses `uint32`/`uint64`/`float32`/`float64` natively |
| `instance.Exports.GetFunction(name)` | `module.ExportedFunction(name)` | Returns `api.Function` |
| `instance.Exports.GetMemory("memory")` | `module.Memory()` | Returns `api.Memory` |
| `fn(args...)` | `fn.Call(ctx, args...)` | Context-aware; returns `[]uint64` |
| `wasmer.Value` | `uint64` | No wrapper type; raw `uint64` values |
| Fuel metering | `context.WithTimeout` / `context.WithCancel` | Standard Go context |
| Memory limit (implicit) | `wazero.NewRuntimeConfig().WithMemoryLimitPages(n)` | Explicit page-level control |

---

## Key Risks and Blockers

### Risk 1: Synthetic Microbenchmark Regression (Medium)
- **Description**: Wazero's interpreter/compiler is ~1.5–3× slower than Wasmer's Cranelift JIT on tight WASM loops
- **Impact**: Only affects mods that perform heavy computation inside WASM (e.g., custom terrain generators running millions of iterations). No known mods exist yet; the mod API is at stub/v6.0-planned stage.
- **Mitigation**: Document the tradeoff. If a specific mod triggers measurable latency, offer a `wazero.NewRuntimeConfigCompiler()` option (uses Go's native compilation on supported platforms: amd64, arm64) vs. `wazero.NewRuntimeConfigInterpreter()` for maximum portability.
- **Acceptance**: ≤10% regression on event-hook mod calls (`weapon.fire`, `enemy.spawn`). Tight-loop microbenchmark regression is acceptable and explicitly called out.

### Risk 2: Host Function Signature Differences (Low)
- **Description**: Wazero uses native Go types (`uint32`, `uint64`) instead of wrapper types (`wasmer.Value`). The `log_message(ptr, len)` host function must be rewritten.
- **Impact**: Straightforward API translation. The current implementation is a stub that ignores the actual message content.
- **Mitigation**: Translate directly; `func(ctx context.Context, mod api.Module, ptr, length uint32)` replaces `func(args []wasmer.Value) ([]wasmer.Value, error)`.

### Risk 3: Fuel Metering Semantic Change (Low)
- **Description**: wasmer-go uses instruction-count fuel; wazero uses `context.WithTimeout` for execution limits.
- **Impact**: Timeout-based limits are coarser than instruction counting but sufficient for mod sandboxing. The `FuelLimit` config field semantics change from "instruction count" to "execution deadline."
- **Mitigation**: Map `FuelLimit` to a proportional timeout (e.g., 1 billion instructions ≈ 1 second). Document the semantic change. Alternatively, use wazero's experimental `WithCloseOnContextDone` for cooperative cancellation.

### Risk 4: Test Infrastructure (Low)
- **Description**: `pkg/mod/wasm_loader_test.go` and `wasm_security_test.go` test against wasmer-go API. Tests must be updated.
- **Impact**: Test coverage temporarily drops during migration. Tests must be rewritten to use wazero API.
- **Mitigation**: Port tests in the same PR as the loader rewrite. Maintain identical test scenarios. Add wazero-specific tests for context cancellation and memory limits.

### No Blockers Identified
- wazero is MIT-licensed (compatible with project licensing)
- wazero supports Go 1.20+ (project uses Go 1.24.0)
- wazero supports all target platforms (Linux, macOS, Windows, JS/WASM, iOS, Android)
- wazero has no CGo dependency (eliminates the Windows/JS platform gaps)

---

## Migration Checklist

- [ ] **1. Add wazero dependency**: Run `go get github.com/tetratelabs/wazero@latest`; verify no advisory vulnerabilities. Acceptance: `go build ./...` succeeds with both wasmer-go and wazero present.
- [ ] **2. Port `WASMLoader` core** (`pkg/mod/wasm_loader.go`): Replace `wasmer.Engine`/`Store`/`Module`/`Instance` with `wazero.Runtime`/`CompiledModule`/`api.Module`. Store `wazero.Runtime` and `wazero.CompiledModule` in `WASMLoader` and `WASMModule` structs. Acceptance: `WASMLoader` compiles with wazero types; no wasmer imports remain.
- [ ] **3. Port `WASMConfig` enforcement**: Map `MemoryLimitPages` via `wazero.NewRuntimeConfig().WithMemoryLimitPages((config.MemoryLimitBytes + 65535) / 65536)` (rounds up to avoid under-allocation). Map `FuelLimit` to `context.WithTimeout(ctx, time.Duration(config.FuelLimit/1e9) * time.Second)`. Acceptance: Loading a module with >64 MB memory request fails; long-running call is cancelled.
- [ ] **4. Port host function bridge** (`createImportObject`): Rewrite `log_message` using `runtime.NewHostModuleBuilder("env").NewFunctionBuilder().WithFunc(func(ctx context.Context, mod api.Module, ptr, length uint32) { ... }).Export("log_message").Instantiate(ctx)`. Read string from WASM memory via `mod.Memory().Read(ptr, length)`. Acceptance: WASM module calling `log_message` produces logrus output.
- [ ] **5. Port `WASMModule.Call()`**: Replace `instance.Exports.GetFunction(name)` with `module.ExportedFunction(name).Call(ctx, args...)`. Convert `interface{}` variadic args to `[]uint64`. Acceptance: Calling an exported WASM function returns correct result.
- [ ] **6. Port `WASMModule.HasExport()`**: Replace with `module.ExportedFunction(name) != nil`. Acceptance: Returns true for existing exports, false for missing.
- [ ] **7. Add `Close()` lifecycle**: Wazero requires explicit `Runtime.Close(ctx)` for resource cleanup. Add `Close()` method to `WASMLoader` and call it during game shutdown. Acceptance: No resource leak warnings; `wazero.Runtime` is closed on `WASMLoader` teardown.
- [ ] **8. Remove platform stubs**: Delete `pkg/mod/wasm_loader_windows.go` and `pkg/mod/wasm_loader_js.go`. Remove `//go:build !js && !windows` from `wasm_loader.go`. Acceptance: `GOOS=windows go build ./pkg/mod/...` and `GOOS=js GOARCH=wasm go build ./pkg/mod/...` both succeed.
- [ ] **9. Port unit tests** (`wasm_loader_test.go`): Rewrite tests to use wazero API. Maintain identical test scenarios (load module, call function, unload, security path validation). Acceptance: `go test -race ./pkg/mod/...` passes with ≥ current coverage.
- [ ] **10. Port security tests** (`wasm_security_test.go`): Rewrite path-isolation and permission tests for wazero. Add context-cancellation test for execution timeout. Acceptance: All security constraints enforced; timeout test cancels runaway module.
- [ ] **11. Add WASM execution benchmarks**: Create `wasm_loader_bench_test.go` with benchmarks for module compilation, Go→WASM call boundary overhead, and memory read. Acceptance: `go test -bench=. ./pkg/mod/...` produces baseline numbers. Go→WASM call-boundary overhead (time spent entering/exiting WASM, excluding WASM body execution) must be ≤10% slower than wasmer-go. Raw WASM execution speed may regress up to 3× for tight-loop microbenchmarks — this is acceptable because mod hooks execute <1000 WASM instructions per event and contribute <0.3% of frame budget.
- [ ] **12. Remove wasmer-go dependency**: Delete `github.com/wasmerio/wasmer-go` from `go.mod`. Run `go mod tidy`. Acceptance: `go mod tidy` completes cleanly; `grep wasmer go.mod` returns empty; `grep wasmer go.sum` returns empty.
- [ ] **13. Remove Rust toolchain from CI** (if present): Audit `.github/workflows/*.yml` for any Rust/Cargo steps. Remove if found. Acceptance: CI pipeline has no `rustc`, `cargo`, or `rustup` references.
- [ ] **14. Update build system**: Verify `go build -o violence .` succeeds without CGo on all target platforms. Verify `CGO_ENABLED=0 go build ./...` succeeds (wazero requires no CGo). Acceptance: Cross-compilation to linux/amd64, darwin/arm64, windows/amd64, js/wasm all succeed with `CGO_ENABLED=0`.
- [ ] **15. Update documentation**: Update `docs/MODDING_WASM.md` to reference wazero. Remove platform limitation notes for Windows and JS/WASM. Update `docs/MOD_SECURITY.md` to describe context-based execution limits. Acceptance: No references to wasmer-go remain in documentation.
- [ ] **16. Validate overall performance parity**: Run full test suite with race detector (`go test -race ./...`). Run game client with 3+ loaded WASM mods. Verify no frame-time regression (mod hooks complete within 1ms per event). Acceptance: All tests pass; no frame-time regression; mod load/call/unload lifecycle works end-to-end.
