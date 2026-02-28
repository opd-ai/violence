# Multi-Platform Build Matrix

This document describes the Violence build system for producing binaries across multiple platforms and architectures.

## Supported Platforms

The build matrix produces artifacts for the following targets:

### Linux
- **amd64** (x86_64): Standard 64-bit Intel/AMD processors
- **arm64** (aarch64): ARM 64-bit processors (e.g., AWS Graviton, Raspberry Pi 4+)

### macOS
- **universal**: Single binary supporting both Intel (amd64) and Apple Silicon (arm64)

### Windows
- **amd64** (x86_64): Standard 64-bit Intel/AMD processors

### WebAssembly
- **wasm**: Browser-based execution via WebAssembly
  - Includes `wasm_exec.js` loader from Go toolchain

### iOS
- **ios**: iOS devices (iPhone, iPad) via gomobile
  - Produces `.xcframework` for embedding in iOS apps
  - Unsigned `.ipa` artifact for testing (requires Apple Developer cert for distribution)

### Android
- **android**: Android devices via gomobile
  - Produces `.aar` (Android Archive) library
  - Can be integrated into Android apps via Gradle

## Build Workflow

The build matrix is defined in `.github/workflows/build.yml` and runs on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`
- Git tags matching `v*` (triggers release builds)

### Build Jobs

#### 1. `build-linux`
Builds for Linux amd64 and arm64 using a matrix strategy.
- **Runner**: `ubuntu-latest`
- **Cross-compilation**: Uses `gcc-aarch64-linux-gnu` for ARM64 builds
- **Dependencies**: Mesa, ALSA, X11 libraries via apt
- **Output**: `violence-linux-amd64`, `violence-linux-arm64`

#### 2. `build-macos`
Builds universal macOS binary supporting both Intel and Apple Silicon.
- **Runner**: `macos-latest`
- **Process**: 
  1. Build separate amd64 and arm64 binaries
  2. Combine using `lipo -create` to create universal binary
- **Output**: `violence-darwin-universal`

#### 3. `build-windows`
Builds for Windows amd64.
- **Runner**: `windows-latest`
- **Output**: `violence-windows-amd64.exe`

#### 4. `build-wasm`
Builds WebAssembly binary for browser execution.
- **Runner**: `ubuntu-latest`
- **CGO**: Disabled (WASM doesn't support CGO)
- **Output**: `violence.wasm`, `wasm_exec.js`

#### 5. `build-ios`
Builds iOS framework using gomobile.
- **Runner**: `macos-latest`
- **Process**:
  1. Install and initialize gomobile
  2. Build `.xcframework` for iOS targets
  3. Create unsigned `.ipa` archive
- **Output**: `Violence.xcframework`, `violence-ios-unsigned.ipa`
- **Note**: Production .ipa requires Apple Developer signing certificate

#### 6. `build-android`
Builds Android archive using gomobile.
- **Runner**: `ubuntu-latest`
- **Dependencies**: JDK 17, Android SDK, NDK 26.1.10909125
- **Process**:
  1. Install and initialize gomobile
  2. Build `.aar` Android Archive
  3. Create minimal APK wrapper structure
- **Output**: `violence.aar`
- **Note**: Full APK requires signing keys and complete Android project

#### 7. `summary`
Aggregates build results and reports overall success/failure.
- Depends on all build jobs (6 platforms)
- Fails if any platform build fails
- Provides consolidated status report

## Artifacts

All build artifacts are uploaded to GitHub Actions with 30-day retention:
- `violence-linux-amd64` (Linux x86_64 binary)
- `violence-linux-arm64` (Linux ARM64 binary)
- `violence-darwin-universal` (macOS universal binary)
- `violence-windows-amd64` (Windows x86_64 executable)
- `violence-wasm` (WebAssembly binary + loader)
- `violence-ios` (iOS .xcframework + unsigned .ipa)
- `violence-android` (Android .aar library)

## Local Building

### Build for Current Platform
```bash
go build -v -o violence .
```

### Cross-Compile for Specific Platform
```bash
# Linux amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -v -o violence-linux-amd64 .

# Linux arm64 (requires cross-compiler)
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -v -o violence-linux-arm64 .

# macOS universal (requires macOS host)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -v -o violence-darwin-amd64 .
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -v -o violence-darwin-arm64 .
lipo -create -output violence-darwin-universal violence-darwin-amd64 violence-darwin-arm64

# Windows amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v -o violence-windows-amd64.exe .

# WASM
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -v -o violence.wasm .
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# iOS (requires macOS)
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
gomobile bind -target=ios -o Violence.xcframework .

# Android (requires Android SDK/NDK)
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
export ANDROID_NDK_HOME=/path/to/android/ndk
gomobile bind -target=android -o violence.aar .
```

## Dependencies

### Runtime Dependencies
Violence requires the following runtime libraries on Linux:
- **OpenGL**: `libgl1-mesa` or equivalent
- **ALSA**: `libasound2` for audio
- **X11**: `libxcursor`, `libxi`, `libxinerama`, `libxrandr`, `libxxf86vm`

On macOS and Windows, system libraries are statically linked or provided by the OS.

### Build Dependencies
- **Go 1.24+**: Required for all platforms
- **GCC/Clang**: Required for CGO builds (Linux, macOS, Windows native)
- **Cross-compilers**: Required for ARM64 builds on x86_64 hosts
  - Linux: `gcc-aarch64-linux-gnu`
  - macOS: Xcode Command Line Tools (automatically includes both architectures)

## CGO Requirements

Violence uses CGO for graphics and audio subsystems. This means:
- C compiler required for all platforms except WASM
- Cross-compilation requires appropriate cross-compiler toolchain
- Static linking not possible for all dependencies (especially graphics drivers)

WASM builds disable CGO (`CGO_ENABLED=0`) and use pure-Go implementations where possible.

## Platform-Specific Notes

### Linux ARM64
Cross-compiling for ARM64 on x86_64 hosts requires:
```bash
sudo apt-get install gcc-aarch64-linux-gnu
export CC=aarch64-linux-gnu-gcc
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -v .
```

### macOS Universal Binary
The `lipo` tool combines separate architecture binaries:
```bash
lipo -create -output violence-universal violence-amd64 violence-arm64
lipo -info violence-universal  # Verify architectures
```

### Windows
Windows builds on non-Windows hosts require MinGW-w64:
```bash
sudo apt-get install mingw-w64
export CC=x86_64-w64-mingw32-gcc
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -v .
```

### WASM
WASM builds produce two files:
- `violence.wasm`: The compiled binary
- `wasm_exec.js`: Go's WASM runtime loader

To run in a browser:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("violence.wasm"), go.importObject)
            .then(result => go.run(result.instance));
    </script>
</head>
<body></body>
</html>
```

### iOS
iOS builds use gomobile to create `.xcframework` files:
- Framework can be embedded in Xcode iOS projects
- Requires macOS with Xcode Command Line Tools
- Production .ipa requires Apple Developer Program membership and signing certificate
- Unsigned builds are for development/testing only

To integrate in iOS app:
1. Add `Violence.xcframework` to Xcode project
2. Import framework: `import Violence`
3. Call exported Go functions from Swift/Objective-C

### Android
Android builds use gomobile to create `.aar` libraries:
- AAR can be added as dependency in Gradle projects
- Requires Android SDK and NDK installation
- Production APK requires signing keys configured in Gradle

To integrate in Android app:
1. Copy `violence.aar` to `app/libs/`
2. Add to `build.gradle`: `implementation files('libs/violence.aar')`
3. Call exported Go functions from Java/Kotlin

## Future Enhancements

Planned additions to the build matrix:
- **Binary signing**: GPG for Linux/Windows, codesigning for macOS, signing for mobile
- **Release automation**: Automatic draft releases with checksums
- **Docker images**: Server builds for containerized deployment
- **Mobile app wrappers**: Complete iOS/Android apps with UI integration

See PLAN.md steps 32-34 for implementation roadmap.
