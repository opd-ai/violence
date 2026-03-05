# Fastlane iOS Configuration

## Overview
Fastlane automates the iOS build, code signing, and App Store submission process for Violence. This document describes the configuration and usage.

## Prerequisites

### Required Tools
- Ruby 2.6+ (system Ruby or rbenv/rvm)
- Bundler gem: `gem install bundler`
- Xcode 15+ (for iOS builds)
- gomobile: `go install golang.org/x/mobile/cmd/gomobile@latest`

### Apple Developer Account
- Apple Developer Program membership ($99/year)
- App Store Connect access
- Team ID from developer.apple.com

## Installation

1. **Install dependencies:**
   ```bash
   bundle install
   ```

2. **Configure environment variables:**
   ```bash
   cp fastlane/.env.example fastlane/.env
   # Edit fastlane/.env with your credentials
   ```

3. **Setup iOS project structure:**
   ```bash
   bundle exec fastlane ios setup
   ```

## Lanes

### Build Lane
Builds and signs the iOS .ipa for distribution:
```bash
bundle exec fastlane ios build
```

**What it does:**
1. Builds Violence.xcframework using gomobile
2. Retrieves code signing certificates via match
3. Builds signed .ipa with gym
4. Outputs to `./build/Violence.ipa`

### Beta Lane
Uploads to TestFlight for beta testing:
```bash
bundle exec fastlane ios beta
```

**What it does:**
1. Runs build lane
2. Uploads .ipa to TestFlight
3. Makes build available to internal testers

### Release Lane
Submits to App Store for review:
```bash
bundle exec fastlane ios release
```

**What it does:**
1. Runs build lane
2. Uploads .ipa to App Store Connect
3. Submits for App Store review
4. Sets to manual release (not automatic)

### Setup Signing Lane
Configures code signing certificates:
```bash
bundle exec fastlane ios setup_signing
```

**What it does:**
1. Creates/syncs App Store distribution certificate
2. Creates/syncs provisioning profile
3. Stores in git repository (configured in Matchfile)

### Sync Signing Lane
Syncs existing certificates (read-only):
```bash
bundle exec fastlane ios sync_signing
```

### Test Lane
Runs iOS unit tests:
```bash
bundle exec fastlane ios test
```

## Code Signing with Match

Match stores certificates and provisioning profiles in a private Git repository for team sharing.

### Initial Setup

1. **Create certificate repository:**
   ```bash
   # Create private GitHub repository: violence-certificates
   # Update MATCH_GIT_URL in fastlane/.env
   ```

2. **Generate certificates:**
   ```bash
   bundle exec fastlane ios setup_signing
   ```

3. **Enter encryption password when prompted** (store securely)

### Team Usage

Team members sync certificates with:
```bash
bundle exec fastlane ios sync_signing
```

## Environment Variables

### Required Variables (.env file)

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_IDENTIFIER` | Bundle identifier | `com.opd-ai.violence` |
| `APPLE_ID` | Apple Developer account email | `dev@opd-ai.com` |
| `TEAM_ID` | Developer Team ID | `AB12CD34EF` |
| `ITC_TEAM_ID` | iTunes Connect Team ID | `AB12CD34EF` |
| `MATCH_GIT_URL` | Certificate repository URL | `https://github.com/opd-ai/violence-certificates` |
| `MATCH_PASSWORD` | Match encryption password | `your-secure-password` |

### Recommended: App Store Connect API

Using App Store Connect API keys avoids two-factor authentication prompts:

| Variable | Description |
|----------|-------------|
| `ASC_KEY_ID` | API Key ID from App Store Connect |
| `ASC_ISSUER_ID` | Issuer ID from App Store Connect |
| `ASC_KEY_CONTENT` | Base64-encoded .p8 key file content |

Generate API keys at: https://appstoreconnect.apple.com/access/api

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Setup Ruby
  uses: ruby/setup-ruby@v1
  with:
    ruby-version: '3.2'
    bundler-cache: true

- name: Build iOS with Fastlane
  env:
    APP_IDENTIFIER: ${{ secrets.APP_IDENTIFIER }}
    APPLE_ID: ${{ secrets.APPLE_ID }}
    TEAM_ID: ${{ secrets.TEAM_ID }}
    MATCH_GIT_URL: ${{ secrets.MATCH_GIT_URL }}
    MATCH_PASSWORD: ${{ secrets.MATCH_PASSWORD }}
    ASC_KEY_ID: ${{ secrets.ASC_KEY_ID }}
    ASC_ISSUER_ID: ${{ secrets.ASC_ISSUER_ID }}
    ASC_KEY_CONTENT: ${{ secrets.ASC_KEY_CONTENT }}
  run: bundle exec fastlane ios beta
```

### Required GitHub Secrets

Add secrets at: Repository Settings → Secrets → Actions

- `APP_IDENTIFIER`
- `APPLE_ID`
- `TEAM_ID`
- `MATCH_GIT_URL`
- `MATCH_PASSWORD`
- `ASC_KEY_ID`
- `ASC_ISSUER_ID`
- `ASC_KEY_CONTENT`

## Troubleshooting

### Certificate Errors

**Problem:** "No certificate found"

**Solution:**
```bash
bundle exec fastlane ios setup_signing
```

### Two-Factor Authentication

**Problem:** Repeated 2FA prompts

**Solution:** Configure App Store Connect API key (recommended) or use application-specific password

### Build Failures

**Problem:** "Xcode project not found"

**Solution:**
1. Ensure iOS project exists in `ios/` directory
2. Link Violence.xcframework to project
3. Configure signing in Xcode

### Match Repository Access

**Problem:** "Could not access git repository"

**Solution:**
1. Verify MATCH_GIT_URL is correct
2. Ensure GitHub access (SSH key or token)
3. Check repository exists and is private

## File Structure

```
violence/
├── Gemfile                      # Ruby dependencies
├── Gemfile.lock                 # Locked dependency versions
├── fastlane/
│   ├── Fastfile                 # Lane definitions
│   ├── Appfile                  # App configuration
│   ├── Matchfile                # Certificate management config
│   ├── .env                     # Environment variables (gitignored)
│   ├── .env.example             # Environment template
│   └── .gitignore               # Fastlane gitignore
└── ios/
    ├── Violence.xcodeproj       # Xcode project (create manually)
    └── Violence.xcworkspace     # Workspace (created by CocoaPods)
```

## Production Checklist

Before first App Store submission:

- [ ] Create iOS project in `ios/` directory
- [ ] Link Violence.xcframework to project
- [ ] Configure app identifier: `com.opd-ai.violence`
- [ ] Setup bundle install: `bundle install`
- [ ] Configure environment variables in `fastlane/.env`
- [ ] Generate certificates: `bundle exec fastlane ios setup_signing`
- [ ] Test build: `bundle exec fastlane ios build`
- [ ] Verify .ipa: `unzip -l build/Violence.ipa`
- [ ] Upload to TestFlight: `bundle exec fastlane ios beta`
- [ ] Test on physical device via TestFlight
- [ ] Submit for review: `bundle exec fastlane ios release`

## Related Documentation

- [MOBILE_BUILD.md](MOBILE_BUILD.md) - Gomobile build pipeline
- [MOBILE_CONTROLS.md](MOBILE_CONTROLS.md) - Touch controls implementation
- [Fastlane Official Docs](https://docs.fastlane.tools/)
- [App Store Connect API](https://developer.apple.com/documentation/appstoreconnectapi)
- [Match Guide](https://docs.fastlane.tools/actions/match/)

## Security Notes

- **Never commit `.env` file** - contains sensitive credentials
- **Never commit certificates** - use match for team sharing
- **Use App Store Connect API** - avoids storing Apple ID password
- **Rotate API keys periodically** - revoke old keys in App Store Connect
- **Keep MATCH_PASSWORD secure** - needed to decrypt certificates
