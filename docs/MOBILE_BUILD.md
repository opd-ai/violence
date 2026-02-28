# Mobile Build Implementation (iOS/Android)

## Overview
This document describes the gomobile-based build pipeline for producing iOS and Android artifacts from the Violence Go codebase.

## Implementation

### iOS Build (`build-ios` job)
- **Platform**: macOS runner (required for iOS toolchain)
- **Tool**: `gomobile bind -target=ios`
- **Outputs**:
  - `Violence.xcframework`: iOS framework bundle
  - `violence-ios-unsigned.ipa`: Unsigned archive for testing
- **Limitations**: Production .ipa requires Apple Developer signing certificate

### Android Build (`build-android` job)
- **Platform**: Ubuntu runner
- **Dependencies**:
  - JDK 17 (Temurin distribution)
  - Android SDK (via android-actions/setup-android)
  - Android NDK 26.1.10909125
- **Tool**: `gomobile bind -target=android`
- **Outputs**:
  - `violence.aar`: Android Archive library
- **Integration**: Can be embedded in Android apps via Gradle

## Gomobile Workflow

Both platforms use the same basic workflow:

1. Install Go 1.24+
2. Install platform-specific dependencies
3. Install gomobile: `go install golang.org/x/mobile/cmd/gomobile@latest`
4. Initialize gomobile: `gomobile init`
5. Build for target platform: `gomobile bind -target=<platform>`

## Framework Integration

### iOS
```swift
import Violence

// Call exported Go functions from Swift
```

### Android
```gradle
// app/build.gradle
dependencies {
    implementation files('libs/violence.aar')
}
```

```java
// Java/Kotlin
import violence.Violence;

// Call exported Go functions
```

## Testing
Comprehensive test suite validates:
- Workflow YAML structure
- Job dependencies (summary depends on all platforms)
- Mobile-specific job configuration (gomobile, SDK setup)
- Documentation coverage
- Platform count accuracy (6 platforms total)

## Production Considerations

### iOS Distribution
- Unsigned .ipa is for development/testing only
- Production distribution requires:
  - Apple Developer Program membership
  - Code signing certificate
  - Provisioning profile
  - Notarization for macOS distribution

### Android Distribution
- .aar library is unsigned
- Production APK requires:
  - Signing keys configured in Gradle
  - Complete Android project structure
  - Google Play Console account for distribution

## Future Enhancements
- Complete iOS app wrapper with UI integration
- Complete Android app wrapper with UI integration
- Automated signing for mobile platforms
- App store deployment automation
- Mobile-specific touch controls (see PLAN.md Known Gaps)

## Related Files
- `.github/workflows/build.yml`: CI/CD pipeline
- `docs/BUILD_MATRIX.md`: Build matrix documentation
- `build_test.go`: Build configuration tests
