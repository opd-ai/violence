# Fastlane Android Configuration

## Overview
This document describes the Fastlane automation for building, signing, and distributing Violence to Google Play Console.

## Prerequisites

### 1. Install Fastlane
```bash
# Install Ruby (macOS/Linux)
brew install ruby
# or use system Ruby

# Install Fastlane
gem install fastlane

# Install Bundler for dependency management
gem install bundler
```

### 2. Install Android SDK and Build Tools
```bash
# Install Android SDK via Android Studio or command line tools
# Ensure ANDROID_HOME is set in your environment
export ANDROID_HOME=$HOME/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
```

### 3. Generate Signing Keystore
```bash
cd fastlane
fastlane android generate_keystore

# Or manually with keytool:
keytool -genkey -v \
  -keystore ../android/violence-release.keystore \
  -alias violence \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000
```

Store the keystore file securely and **never commit it to Git**.

### 4. Configure Google Play API Access

1. Go to [Google Play Console](https://play.google.com/console)
2. Select your app (or create a new one)
3. Navigate to **Setup > API access**
4. Create a new service account or link an existing one
5. Grant necessary permissions (Release Manager or higher)
6. Download the JSON key file
7. Save it as `android/google-play-api-key.json`
8. **Never commit this file to Git** (it's gitignored)

### 5. Configure Environment Variables
```bash
cd fastlane
cp .env.example .env
```

Edit `.env` and fill in your Android values:
```bash
# Android Configuration
ANDROID_PACKAGE_NAME=ai.opd.violence
ANDROID_KEYSTORE_PATH=../android/violence-release.keystore
ANDROID_KEYSTORE_PASSWORD=your_keystore_password
ANDROID_KEY_ALIAS=violence
ANDROID_KEY_PASSWORD=your_key_password
GOOGLE_PLAY_JSON_KEY_PATH=../android/google-play-api-key.json
```

## Available Lanes

### Build
Build and sign the Android App Bundle (.aab):
```bash
fastlane android build
```

This lane:
1. Runs `gomobile bind -target=android` to create `violence.aar`
2. Builds the Android project with Gradle
3. Signs the `.aab` with your release keystore
4. Outputs to `android/app/build/outputs/bundle/release/app-release.aab`

### Internal Track
Upload to Google Play Internal Testing track:
```bash
fastlane android internal
```

Use this for:
- Internal QA testing
- Team testing before beta
- Rapid iteration cycles

### Beta Track
Upload to Google Play Beta track:
```bash
fastlane android beta
```

Use this for:
- Public beta testing
- Opt-in beta testers
- Pre-release validation with metadata/screenshots

### Production Release
Submit to Google Play Production with staged rollout:
```bash
fastlane android release
```

This lane:
1. Builds and signs the `.aab`
2. Uploads to Production track with **10% rollout**
3. Includes metadata and screenshots
4. Does **not** auto-publish (requires manual approval in Console)

### Promote Rollout
Increase rollout percentage to 100%:
```bash
fastlane android promote
```

Use this after monitoring crash reports and reviews on the 10% rollout.

### Setup
Create Android project structure:
```bash
fastlane android setup
```

This creates:
- `android/app/src/main/java/` (Java/Kotlin source)
- `android/app/libs/` (for violence.aar)
- `android/app/src/main/res/` (resources)

### Test
Run Android instrumentation tests:
```bash
fastlane android test
```

Requires a connected device or emulator.

## Project Structure

```
violence/
├── android/
│   ├── app/
│   │   ├── build.gradle           # App-level Gradle config
│   │   ├── libs/
│   │   │   └── violence.aar       # Gomobile output
│   │   └── src/
│   │       └── main/
│   │           ├── AndroidManifest.xml
│   │           ├── java/
│   │           │   └── ai/opd/violence/  # Java/Kotlin source
│   │           └── res/            # Resources (icons, strings, etc.)
│   ├── build.gradle                # Project-level Gradle config
│   ├── violence-release.keystore   # Signing keystore (gitignored)
│   └── google-play-api-key.json    # API credentials (gitignored)
└── fastlane/
    ├── Appfile
    ├── Fastfile
    ├── .env                        # Environment variables (gitignored)
    └── .env.example
```

## CI/CD Integration

### GitHub Actions
```yaml
name: Android Release

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.2'
          bundler-cache: true
      
      - name: Set up JDK 17
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'
      
      - name: Set up Android SDK
        uses: android-actions/setup-android@v3
      
      - name: Install Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Install gomobile
        run: |
          go install golang.org/x/mobile/cmd/gomobile@latest
          gomobile init
      
      - name: Decode secrets
        env:
          KEYSTORE_BASE64: ${{ secrets.ANDROID_KEYSTORE_BASE64 }}
          GOOGLE_PLAY_JSON_BASE64: ${{ secrets.GOOGLE_PLAY_JSON_BASE64 }}
        run: |
          echo "$KEYSTORE_BASE64" | base64 -d > android/violence-release.keystore
          echo "$GOOGLE_PLAY_JSON_BASE64" | base64 -d > android/google-play-api-key.json
      
      - name: Deploy to Google Play Internal
        env:
          ANDROID_PACKAGE_NAME: ${{ secrets.ANDROID_PACKAGE_NAME }}
          ANDROID_KEYSTORE_PATH: ../android/violence-release.keystore
          ANDROID_KEYSTORE_PASSWORD: ${{ secrets.ANDROID_KEYSTORE_PASSWORD }}
          ANDROID_KEY_ALIAS: ${{ secrets.ANDROID_KEY_ALIAS }}
          ANDROID_KEY_PASSWORD: ${{ secrets.ANDROID_KEY_PASSWORD }}
          GOOGLE_PLAY_JSON_KEY_PATH: ../android/google-play-api-key.json
        run: |
          cd fastlane
          bundle exec fastlane android internal
```

### GitLab CI
```yaml
android-deploy:
  stage: deploy
  image: ruby:3.2
  before_script:
    - apt-get update && apt-get install -y openjdk-17-jdk
    - gem install bundler
    - bundle install
    - export ANDROID_HOME=/opt/android-sdk
  script:
    - echo "$ANDROID_KEYSTORE_BASE64" | base64 -d > android/violence-release.keystore
    - echo "$GOOGLE_PLAY_JSON_BASE64" | base64 -d > android/google-play-api-key.json
    - cd fastlane
    - bundle exec fastlane android internal
  only:
    - tags
```

## Production Checklist

Before submitting to Google Play Production:

- [ ] App builds successfully with `fastlane android build`
- [ ] Version code and version name updated in `build.gradle`
- [ ] All strings externalized to `strings.xml` for localization
- [ ] App icon and screenshots prepared (1024x500 feature graphic, screenshots for phone/tablet/TV)
- [ ] Privacy policy URL added to Google Play listing
- [ ] Age rating questionnaire completed (IARC)
- [ ] Content rating certificates obtained if required
- [ ] Store listing metadata complete (title, short description, full description)
- [ ] Release notes written for this version
- [ ] Proguard/R8 enabled for code shrinking and obfuscation
- [ ] Crash reporting enabled (Firebase Crashlytics, Sentry, etc.)
- [ ] Tested on physical devices (minimum API level 21)
- [ ] Google Play pre-launch report reviewed (no critical crashes)
- [ ] Internal testing completed with >20 testers for 14+ days (if first release)
- [ ] Beta testing completed with feedback addressed
- [ ] Signing keystore backed up securely (store password in password manager)
- [ ] Google Play API key file backed up securely

## Troubleshooting

### "Gradle task failed"
- Ensure `ANDROID_HOME` is set correctly
- Verify Android SDK Build Tools are installed
- Check `android/build.gradle` for syntax errors

### "Keystore not found"
- Verify `ANDROID_KEYSTORE_PATH` in `.env`
- Ensure keystore file exists at the specified path
- Run `fastlane android generate_keystore` if needed

### "Google Play API authentication failed"
- Verify `GOOGLE_PLAY_JSON_KEY_PATH` points to valid JSON key
- Ensure service account has "Release Manager" role in Google Play Console
- Check that the app exists in Google Play Console

### "App not available in any country"
- Go to Google Play Console > Production > Countries/regions
- Select countries for distribution
- Apps must be available in at least one country

### "Version code already exists"
- Increment `versionCode` in `android/app/build.gradle`
- Google Play requires strictly increasing version codes

## Related Documentation
- [MOBILE_BUILD.md](./MOBILE_BUILD.md) - Gomobile build process
- [FASTLANE_IOS.md](./FASTLANE_IOS.md) - iOS Fastlane configuration
- [Official Fastlane Android Docs](https://docs.fastlane.tools/getting-started/android/setup/)
- [Google Play Developer Guide](https://developer.android.com/distribute/googleplay)
