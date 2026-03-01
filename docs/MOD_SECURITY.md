# Mod Security Evaluation

## Overview

Violence supports user-generated content through a modding system. This document evaluates security approaches for sandboxing untrusted mod code, compares Go plugins vs WebAssembly (WASM), and provides a recommendation for safe mod execution.

## Security Threat Model

### Attack Vectors

**Malicious Mod Capabilities Without Sandboxing**:

1. **File System Access**: Read/write arbitrary files, steal save data, crypto wallets
2. **Network Access**: Exfiltrate player data, participate in botnets, DDoS attacks
3. **Memory Corruption**: Buffer overflows, use-after-free, arbitrary code execution
4. **Resource Exhaustion**: Infinite loops, memory leaks, disk filling
5. **System Calls**: Execute arbitrary commands, install malware, privilege escalation
6. **Supply Chain**: Trojanized dependencies in mod's `go.mod`

### Affected Assets

- **Player Data**: Save files, config, credentials, chat logs
- **System Resources**: CPU, RAM, disk, network bandwidth
- **Game Integrity**: Cheating, griefing, server crashes
- **Privacy**: IP address, hardware info, installed software

### Threat Actors

- **Script Kiddies**: Copy-paste malware, minimal skill
- **Griefers**: Crash servers, disrupt gameplay
- **Data Thieves**: Steal credentials, crypto wallets
- **Nation-State**: Advanced persistent threats (unlikely for indie game, but possible)

## Option 1: Go Plugins (Current Approach)

### How Go Plugins Work

Go's `plugin` package allows loading `.so` (Linux) or `.dylib` (macOS) shared libraries at runtime:

```go
// Load plugin
p, err := plugin.Open("mods/my_mod.so")
if err != nil {
    return err
}

// Lookup symbol
symbol, err := p.Lookup("Init")
if err != nil {
    return err
}

// Type assert and call
initFunc := symbol.(func() error)
initFunc()
```

**Mod Compilation**:
```bash
go build -buildmode=plugin -o my_mod.so my_mod.go
```

### Security Analysis

#### ✅ Advantages

1. **Native Performance**: No overhead, runs at native speed
2. **Full Go Ecosystem**: Mods use standard library, third-party packages
3. **Type Safety**: Compile-time type checking, runtime panics caught by Go runtime
4. **Debugging**: Standard Go tooling (delve, pprof)

#### ❌ Critical Security Flaws

1. **No Sandboxing**: Plugins run in same process/memory space as game
   - Full access to game state, memory, file system, network
   - Can call any exported function, modify global variables
   - No capability restrictions (can't limit syscalls)

2. **Platform-Specific**: Only works on Linux and macOS (not Windows, BSD, mobile)
   - Windows uses DLLs, incompatible with Go plugins
   - Requires separate builds per platform

3. **Version Locking**: Plugin must be built with **exact same Go version** as game
   - Go 1.21.3 plugin won't load in Go 1.21.4 game
   - Forces users to rebuild mods on every game update

4. **Irreversible Load**: Once loaded, cannot unload plugins
   - Memory leaks persist entire session
   - Can't hotswap mods without restarting game

5. **Symbol Pollution**: All exported symbols visible to plugin
   - Plugin can call internal functions not intended for modding API
   - Easy to bypass intended API boundaries

6. **Supply Chain Risk**: Plugin's dependencies unvetted
   - Mod pulls in malicious `go.mod` dependency
   - No dependency signature verification

#### 🔴 Exploits Demonstrated in the Wild

**Real-World Plugin Vulnerabilities**:

- **Minecraft Forge (Java)**: Mods commonly include remote code execution (RCE) payloads via reflection API abuse
- **Garry's Mod (Lua)**: Before sandboxing, Lua mods stole Steam credentials via `os.execute()`
- **Vim Plugins (Vimscript)**: Malicious plugins exfiltrate files via `:!curl`

**Go Plugin Specific**:

- **No Known Exploits**: Go plugins rarely used in production (unstable API)
- **Theoretical Attack**: Plugin calls `os.Exec("/bin/sh")` to spawn shell

### Verdict: ⛔ UNSAFE for Untrusted Mods

Go plugins provide **zero security** against malicious code. Acceptable only for:
- Trusted first-party mods
- Closed development environments
- Internal tooling (not user-facing)

## Option 2: WebAssembly (WASM) Sandboxing

### How WASM Works

WebAssembly is a binary instruction format designed for safe, sandboxed execution:

```
┌──────────────────┐
│   Mod Code       │
│  (Go/Rust/C)     │
└────────┬─────────┘
         │ Compile
         ▼
┌──────────────────┐       ┌──────────────────┐
│  WASM Binary     │──────►│  WASM Runtime    │
│  (.wasm file)    │       │  (Wasmer/Wasmtime)│
└──────────────────┘       └────────┬─────────┘
                                    │
                                    ▼
                           ┌──────────────────┐
                           │   Host Game      │
                           │   (Violence)     │
                           └──────────────────┘
```

**Mod Compilation** (TinyGo, supports WASM):
```bash
tinygo build -o mod.wasm -target wasi mod.go
```

**Runtime Integration** (using Wasmer):
```go
import "github.com/wasmerio/wasmer-go/wasmer"

// Load WASM module
wasmBytes, _ := os.ReadFile("mods/my_mod.wasm")
engine := wasmer.NewEngine()
store := wasmer.NewStore(engine)
module, _ := wasmer.NewModule(store, wasmBytes)

// Create instance with limited imports
importObject := wasmer.NewImportObject()
instance, _ := wasmer.NewInstance(module, importObject)

// Call exported function
initFunc, _ := instance.Exports.GetFunction("init")
initFunc()
```

### Security Analysis

#### ✅ Advantages

1. **Strong Sandboxing**: WASM runs in isolated linear memory
   - No access to host memory, file system, network by default
   - All host interactions via explicit imports (capability-based security)
   - Cannot escape sandbox without host granting permission

2. **Cross-Platform**: WASM runs identically on Windows, Linux, macOS, mobile, web
   - Single `.wasm` binary works everywhere
   - No recompilation needed per platform

3. **Version Agnostic**: WASM binary stable across Go/TinyGo versions
   - Game updates don't break mods (stable ABI)
   - Mod compiled with TinyGo 0.28 works with 0.30+ runtime

4. **Fine-Grained Permissions**: Host explicitly grants capabilities
   - Example: Mod requests file read → game shows permission dialog
   - Can revoke permissions at runtime

5. **Resource Limits**: Enforce memory, CPU, fuel limits
   - Prevent infinite loops (fuel exhaustion)
   - Cap memory usage (e.g., 64MB per mod)

6. **Hotswapping**: Load/unload modules dynamically
   - Enable/disable mods without restart
   - Update mods mid-session

7. **Determinism**: WASM execution is deterministic (important for multiplayer)
   - Same input → same output (reproducible)
   - Easier to debug desyncs

#### ❌ Disadvantages

1. **Performance Overhead**: 5-30% slower than native
   - JIT compilation helps (Wasmer/Wasmtime use Cranelift)
   - AOT compilation mitigates (compile WASM to native ahead-of-time)
   - Negligible for most mods (not tight inner loops)

2. **Limited Go Support**: TinyGo required (not full Go)
   - No goroutines, channels, `reflect`, `cgo`
   - Subset of standard library (`fmt`, `math`, `strings` work; `net/http` doesn't)
   - Many third-party packages unsupported

3. **API Complexity**: Host must define WASM imports
   - More boilerplate than plugin `Lookup()`
   - Type marshaling between Go and WASM (use MessagePack or Protobuf)

4. **Debugging Challenges**: WASM debugging tools immature
   - Stack traces less helpful (function indices, not names)
   - Printf debugging still works (`wasi.fd_write`)

5. **Initial Learning Curve**: Developers unfamiliar with WASM/WASI
   - Documentation improving but not mainstream yet

### Capability-Based Security Example

**Secure API Design**:
```go
// Host-side (Violence game)
type ModAPI struct {
    AllowFileRead  bool
    AllowFileWrite bool
    AllowNetwork   bool
}

func (api *ModAPI) ReadFile(path string) ([]byte, error) {
    if !api.AllowFileRead {
        return nil, errors.New("permission denied: file read")
    }
    // Whitelist paths (only mods/ directory)
    if !strings.HasPrefix(path, "mods/") {
        return nil, errors.New("access denied: outside mods directory")
    }
    return os.ReadFile(path)
}

// Export to WASM
importObject := wasmer.NewImportObject()
importObject.Register("env", map[string]wasmer.IntoExtern{
    "read_file": wasmer.NewFunction(store, 
        wasmer.NewFunctionType(...),
        api.ReadFile,
    ),
})
```

**Mod-side** (TinyGo):
```go
//export init
func init() {
    data := readFile("mods/config.txt")
    // Cannot access ../../../etc/passwd (blocked by host)
}

//go:wasm-module env
//export read_file
func readFile(path string) []byte
```

### Performance Comparison

| Operation | Native Go | Go Plugin | WASM (Wasmer) | WASM (AOT) |
|-----------|-----------|-----------|---------------|------------|
| Function Call | 1.0x | 1.0x | 1.1x | 1.0x |
| String Manipulation | 1.0x | 1.0x | 1.2x | 1.05x |
| Math (float64) | 1.0x | 1.0x | 1.0x | 1.0x |
| Memory Allocation | 1.0x | 1.0x | 1.3x | 1.1x |
| File I/O (via import) | 1.0x | 1.0x | 1.5x | 1.2x |

**Benchmark Source**: Wasmer benchmarks (2024), WASM-4 game engine

### WASM Runtimes Comparison

| Runtime | Language | JIT | AOT | WASI Support | Go Bindings | Maturity |
|---------|----------|-----|-----|--------------|-------------|----------|
| Wasmer | Rust | ✅ | ✅ | ✅ | ✅ (wasmer-go) | Production |
| Wasmtime | Rust | ✅ | ✅ | ✅ | ✅ (wasmtime-go) | Production |
| Wazero | Go | ❌ | ✅ | ✅ | Native | Beta |

**Recommendation**: **Wasmer** (best balance of performance and maturity)

### Verdict: ✅ SAFE for Untrusted Mods

WASM provides industrial-grade sandboxing with acceptable performance tradeoff.

## Option 3: Hybrid Approach (WASM + Signed Plugins)

### Concept

- **Untrusted Mods**: Run in WASM sandbox (default)
- **Verified Mods**: Native plugins if signed by trusted authors
- **User Choice**: Players opt-in to native plugins (warning dialog)

### Implementation

```go
type ModLoader struct {
    TrustedKeys []ed25519.PublicKey
}

func (ml *ModLoader) LoadMod(path string) error {
    // Check for signature
    sig, err := os.ReadFile(path + ".sig")
    if err == nil && ml.VerifySignature(path, sig) {
        // Trusted mod → load as native plugin
        return ml.LoadPlugin(path)
    }
    
    // Untrusted mod → load as WASM
    return ml.LoadWASM(path)
}
```

**Signature Creation** (mod author):
```bash
# Author signs mod with private key
openssl dgst -sha256 -sign author_private.pem -out mod.wasm.sig mod.wasm

# Distribute public key
# Players add to trusted keys: ~/.violence/trusted_mod_authors.txt
```

### Advantages

- **Best of Both Worlds**: Performance for trusted mods, safety for untrusted
- **Gradual Trust**: Players expand trusted author list over time
- **Reputation System**: Community curates trusted authors

### Disadvantages

- **Complexity**: Maintain two mod loaders
- **False Security**: Stolen signing keys compromise trust
- **Key Distribution**: How do players discover legitimate public keys?

### Verdict: 🟡 Acceptable Compromise

Good for ecosystems with established mod authors (e.g., Steam Workshop). Overkill for new games.

## Option 4: Language-Specific Sandboxes

### Lua (Embedded Scripting)

**Implementation**: `github.com/yuin/gopher-lua`

**Sandboxing**:
```go
L := lua.NewState()
defer L.Close()

// Disable dangerous functions
L.SetGlobal("os", lua.LNil)
L.SetGlobal("io", lua.LNil)
L.SetGlobal("dofile", lua.LNil)

// Load mod
L.DoFile("mods/my_mod.lua")
```

**Pros**:
- Mature sandboxing (used by World of Warcraft, Roblox)
- Easy to learn (simple syntax)
- Excellent performance (LuaJIT)

**Cons**:
- Not Go (different language for modders)
- API design burden (export every function manually)
- Sandbox escapes exist (require careful API design)

### JavaScript (QuickJS, Goja)

**Implementation**: `github.com/dop251/goja`

**Pros**:
- Familiar to web developers
- Large ecosystem (npm packages)

**Cons**:
- Slower than Lua
- Goja doesn't support all ES6 features
- Node.js APIs (fs, http) must be reimplemented

### Verdict: 🟡 Viable Alternative

Lua preferred over JS for game modding (better performance, simpler). Both inferior to WASM for security.

## Comparison Matrix

| Criterion | Go Plugin | WASM | Signed Plugin | Lua |
|-----------|-----------|------|---------------|-----|
| **Security** | ⛔ None | ✅ Strong | 🟡 Moderate | 🟡 Moderate |
| **Performance** | ✅ Native | 🟡 95% | ✅ Native | ✅ 90% |
| **Cross-Platform** | ❌ Linux/Mac | ✅ All | ❌ Linux/Mac | ✅ All |
| **Version Stability** | ❌ Locked | ✅ Stable | ❌ Locked | ✅ Stable |
| **Ease of Use (Modders)** | ✅ Full Go | 🟡 TinyGo | ✅ Full Go | ✅ Simple |
| **Ease of Use (Devs)** | ✅ Simple | 🟡 Complex | 🟡 Complex | 🟡 Moderate |
| **Ecosystem** | ✅ Full | 🟡 Limited | ✅ Full | ✅ Large |

## Recommendation

### Primary: **WebAssembly (WASM) with Wasmer**

**Rationale**:
1. **Security First**: Sandboxing is non-negotiable for untrusted mods
2. **Cross-Platform**: Single binary works on all platforms (critical for mobile/web ports)
3. **Future-Proof**: WASM adoption growing (Docker, Cloudflare Workers, Unity)
4. **Acceptable Tradeoffs**: 5-10% overhead negligible for mod logic (not rendering)

**Migration Path**:
1. **v5.1**: Implement WASM loader with Wasmer, define initial mod API
2. **v5.2**: Port 3-5 reference mods to TinyGo WASM (weapon pack, map generator)
3. **v5.3**: Deprecate Go plugins, provide migration guide
4. **v6.0**: Remove plugin support entirely

### Fallback: **Lua Scripting**

If WASM proves too restrictive (TinyGo limitations too severe):

**Rationale**:
- Proven in game industry (WoW, Garry's Mod, Roblox)
- Better performance than JS
- Easier API bindings than WASM

**Migration Path**:
1. **v5.1**: Add `github.com/yuin/gopher-lua`, define Lua API
2. **v5.2**: Provide Lua mod examples
3. **v6.0**: Maintain both WASM and Lua (different use cases)

### NOT Recommended: **Go Plugins**

Continue supporting plugins **only** for:
- Development/testing (trusted environment)
- Server-side admin tools (controlled deployment)

Mark plugins as "unsafe" in documentation, require opt-in flag:
```bash
violence-server --enable-unsafe-plugins
```

## Implementation Plan

### Phase 1: WASM Proof of Concept (2 weeks)

```bash
# Install Wasmer
go get github.com/wasmerio/wasmer-go/wasmer

# Create pkg/modding/wasm_loader.go
package modding

import "github.com/wasmerio/wasmer-go/wasmer"

type WASMLoader struct {
    engine *wasmer.Engine
    store  *wasmer.Store
}

func (wl *WASMLoader) LoadMod(path string) (*WASMMod, error) {
    wasmBytes, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    module, err := wasmer.NewModule(wl.store, wasmBytes)
    if err != nil {
        return nil, err
    }
    
    instance, err := wasmer.NewInstance(module, wl.createImports())
    if err != nil {
        return nil, err
    }
    
    return &WASMMod{instance: instance}, nil
}
```

### Phase 2: Define Mod API (4 weeks)

**Core API Categories**:
1. **Events**: `OnPlayerSpawn`, `OnEnemyKilled`, `OnLevelLoad`
2. **Entities**: `SpawnEntity`, `SetPosition`, `GetHealth`
3. **Assets**: `LoadTexture`, `LoadSound`, `LoadModel`
4. **UI**: `ShowNotification`, `CreateMenu`, `DrawHUD`
5. **Storage**: `SaveData`, `LoadData` (sandboxed paths)

**API Definition** (`pkg/modding/api.go`):
```go
type ModAPI interface {
    // Events
    RegisterEventHandler(event string, handler func(EventData))
    
    // Entities
    SpawnEntity(entityType string, x, y float64) EntityID
    SetEntityPosition(id EntityID, x, y float64)
    
    // Assets (returns handles, not raw data)
    LoadTexture(path string) TextureID
    PlaySound(soundID SoundID)
    
    // UI
    ShowNotification(message string)
}
```

### Phase 3: Reference Mods (2 weeks)

**Example Mods**:
1. **Weapon Pack**: Add 3 new weapons
2. **Map Generator**: Custom BSP generation algorithm
3. **HUD Mod**: Custom health bar design
4. **Difficulty Modifier**: Double enemy health

### Phase 4: Documentation (1 week)

- `docs/MODDING_WASM.md`: WASM mod creation guide
- `docs/MOD_API.md`: Complete API reference
- `examples/mods/`: Reference mod source code

### Phase 5: Deprecation (v6.0)

- Remove `plugin.Open()` code paths
- Archive plugin examples to `legacy/plugins/`
- Update `docs/MODDING.md` to focus on WASM

## Testing Strategy

### Security Testing

**Malicious Mod Test Cases**:
1. **File Exfiltration**: Attempt to read `/etc/passwd` → should fail
2. **Network Access**: Attempt HTTP request → should fail (unless import granted)
3. **Resource Exhaustion**: Infinite loop → should hit fuel limit
4. **Memory Corruption**: Out-of-bounds write → WASM runtime traps
5. **Denial of Service**: Allocate 10GB memory → should hit limit

**Test Harness**:
```go
func TestModSecurity(t *testing.T) {
    loader := NewWASMLoader()
    mod, err := loader.LoadMod("testdata/malicious_file_read.wasm")
    require.NoError(t, err)
    
    // Call mod function that tries to read /etc/passwd
    result, err := mod.Call("attempt_read_passwd")
    
    // Should fail gracefully
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

### Performance Testing

**Benchmark Suite**:
```go
func BenchmarkWASMModCall(b *testing.B) {
    loader := NewWASMLoader()
    mod, _ := loader.LoadMod("testdata/benchmark_mod.wasm")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mod.Call("noop")
    }
}

// Expected: <100ns per call (vs ~10ns native)
```

## Open Questions

1. **WASI Preview 2**: Wait for stabilization or use Preview 1?
   - **Answer**: Use Preview 1 (stable), migrate to Preview 2 when ready

2. **Component Model**: Use WASM Component Model for better ABI?
   - **Answer**: Too early (2026), stick with basic WASM for now

3. **Rust Mods**: Support Rust in addition to Go?
   - **Answer**: Yes, WASM is language-agnostic (Rust compiles to WASM easily)

4. **Mod Marketplace**: Centralized mod distribution?
   - **Answer**: Deferred to v6.0+ (focus on local mods first)

## Conclusion

**WASM sandboxing is the clear winner** for untrusted mod execution. The performance overhead is acceptable, security guarantees are strong, and cross-platform support aligns with Violence's goals. While Go plugins are simpler to implement, they pose unacceptable security risks for user-generated content.

**Action Items**:
1. ✅ Deprecate Go plugins for untrusted mods (mark as unsafe)
2. 🔄 Implement WASM loader with Wasmer (v5.1 priority)
3. 🔄 Define stable mod API with capability-based security
4. 🔄 Create 3-5 reference WASM mods
5. 🔄 Publish modding guide with security best practices
6. 📅 Full plugin removal in v6.0

**Security Posture**: Moving from ⛔ **CRITICAL RISK** (plugins) to ✅ **SECURE** (WASM) is essential before enabling community modding.
