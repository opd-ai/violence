# Mobile Publishing Guide

## Overview
This document provides a comprehensive, step-by-step checklist for submitting Violence to the Apple App Store (iOS) and Google Play Store (Android). Follow each section in order to ensure a successful first-time submission.

---

## Prerequisites

### Development Environment
- [ ] macOS (for iOS builds) or Linux/macOS (for Android)
- [ ] Xcode 15+ installed (iOS only)
- [ ] Android SDK and Build Tools installed (Android only)
- [ ] Go 1.24+ installed
- [ ] gomobile installed: `go install golang.org/x/mobile/cmd/gomobile@latest`
- [ ] Ruby 2.6+ installed (for Fastlane)
- [ ] Bundler installed: `gem install bundler`

### Developer Accounts
- [ ] Apple Developer Program membership ($99/year) - iOS
- [ ] Google Play Developer account ($25 one-time) - Android
- [ ] Access credentials documented in secure password manager

### Repository Setup
- [ ] Clone Violence repository: `git clone https://github.com/opd-ai/violence.git`
- [ ] Run `bundle install` to install Fastlane dependencies
- [ ] Verify gomobile build works: see [MOBILE_BUILD.md](MOBILE_BUILD.md)

---

## Phase 1: iOS Submission

### 1.1. Apple Developer Setup

#### App Store Connect Configuration
- [ ] Log in to [App Store Connect](https://appstoreconnect.apple.com/)
- [ ] Navigate to **My Apps** → **+ New App**
- [ ] Create new app entry:
  - **Platform**: iOS
  - **Name**: Violence
  - **Primary Language**: English (U.S.)
  - **Bundle ID**: `com.opd-ai.violence` (must match Xcode project)
  - **SKU**: `violence-ios` (or similar unique identifier)
  - **User Access**: Full Access
- [ ] Save app entry

#### Team Configuration
- [ ] Note your **Team ID** from [Apple Developer Account](https://developer.apple.com/account/) → Membership
- [ ] Note your **iTunes Connect Team ID** from App Store Connect → Users and Access → Keys

### 1.2. Code Signing Setup

#### Match Certificate Repository
- [ ] Create private GitHub repository: `violence-certificates`
- [ ] Initialize repository: `git init && git commit --allow-empty -m "Initial commit"`
- [ ] Push to GitHub: `git push -u origin main`

#### Environment Variables
- [ ] Copy Fastlane environment template: `cp fastlane/.env.example fastlane/.env`
- [ ] Edit `fastlane/.env` with iOS values:
  ```bash
  # iOS Configuration
  APP_IDENTIFIER=com.opd-ai.violence
  APPLE_ID=your-apple-id@example.com
  TEAM_ID=YOUR_TEAM_ID
  ITC_TEAM_ID=YOUR_ITC_TEAM_ID
  MATCH_GIT_URL=https://github.com/opd-ai/violence-certificates
  MATCH_PASSWORD=your-secure-password  # Store in password manager
  ```

#### Generate Certificates
- [ ] Run: `bundle exec fastlane ios setup_signing`
- [ ] Enter Match password when prompted (store securely)
- [ ] Verify certificates created in Match repository
- [ ] Document Match password in team password manager

#### Optional: App Store Connect API (Recommended)
- [ ] Generate API key at [App Store Connect → Users and Access → Keys](https://appstoreconnect.apple.com/access/api)
- [ ] Download `.p8` key file
- [ ] Add to `fastlane/.env`:
  ```bash
  ASC_KEY_ID=YOUR_KEY_ID
  ASC_ISSUER_ID=YOUR_ISSUER_ID
  ASC_KEY_CONTENT=<base64-encoded .p8 file>
  ```
- [ ] Encode `.p8` file: `base64 -i AuthKey_KEYID.p8 | pbcopy`

### 1.3. iOS Project Setup

#### Create Xcode Project
- [ ] Open Xcode → Create New Project → iOS App
- [ ] Project name: **Violence**
- [ ] Organization identifier: **ai.opd**
- [ ] Bundle identifier: **com.opd-ai.violence** (matches App Store Connect)
- [ ] Save to `ios/` directory in Violence repository

#### Link gomobile Framework
- [ ] Build Violence.xcframework: `gomobile bind -target=ios -o Violence.xcframework .`
- [ ] Drag `Violence.xcframework` into Xcode project under **Frameworks**
- [ ] Select **Embed & Sign** in General → Frameworks, Libraries, and Embedded Content

#### Configure Signing
- [ ] Open Xcode → Violence target → Signing & Capabilities
- [ ] Enable **Automatically manage signing**
- [ ] Select Team: Your Apple Developer Team
- [ ] Verify bundle identifier matches: `com.opd-ai.violence`
- [ ] Build to validate signing: `Cmd+B`

### 1.4. App Store Metadata

#### Metadata Files Location
All metadata is in `fastlane/metadata/ios/en-US/`:

- [ ] Verify `name.txt`: "Violence"
- [ ] Verify `description.txt`: Full 4000-character description
- [ ] Verify `keywords.txt`: Comma-separated keywords
- [ ] Verify `promotional_text.txt`: Short promotional message
- [ ] Verify `marketing_url.txt`: Project homepage
- [ ] Verify `support_url.txt`: Support/documentation URL
- [ ] Verify `privacy_url.txt`: Privacy policy URL (see [PRIVACY_POLICY.md](PRIVACY_POLICY.md))
- [ ] Verify `copyright.txt`: "2026 OPD AI"

#### Screenshots
Prepare screenshots in `fastlane/metadata/ios/en-US/screenshots/`:

- [ ] **iPhone 6.7"** (iPhone 14 Pro Max): 1290×2796 pixels (required)
  - [ ] Minimum 3 screenshots: Main menu, gameplay, multiplayer
- [ ] **iPhone 6.5"** (iPhone 11 Pro Max): 1242×2688 pixels (fallback)
  - [ ] Same 3 screenshots as 6.7"
- [ ] **iPad Pro 12.9" (3rd gen)**: 2048×2732 pixels (optional but recommended)
  - [ ] Minimum 3 screenshots

**Screenshot Content Suggestions**:
1. Main menu with genre selection UI
2. First-person gameplay showing corridor combat
3. Multiplayer lobby or active match
4. HUD display with health/ammo indicators
5. Mod browser UI (if showcasing user content)

#### Age Rating
- [ ] Review [AGE_RATINGS.md](AGE_RATINGS.md) for ESRB questionnaire responses
- [ ] During App Store Connect submission, select:
  - **Rating**: 13+ (TEEN equivalent)
  - **Content descriptors**:
    - [x] Frequent/Intense Cartoon or Fantasy Violence
    - [x] Infrequent/Mild Realistic Violence
    - [x] Unrestricted Web Access (multiplayer chat)

### 1.5. iOS Build and Test

#### Build .ipa
- [ ] Run: `bundle exec fastlane ios build`
- [ ] Verify output: `build/Violence.ipa` exists
- [ ] Validate .ipa structure: `unzip -l build/Violence.ipa`
- [ ] Check for expected files: `Payload/Violence.app/Violence`

#### Upload to TestFlight
- [ ] Run: `bundle exec fastlane ios beta`
- [ ] Wait for App Store Connect processing (10-30 minutes)
- [ ] Log in to [App Store Connect → TestFlight](https://appstoreconnect.apple.com/)
- [ ] Verify build appears under **iOS Builds**
- [ ] Add test information (What to Test notes)

#### Internal Testing
- [ ] Add internal testers in TestFlight → Internal Testing
- [ ] Distribute build to testers
- [ ] Testers install via TestFlight app
- [ ] Collect feedback on:
  - [ ] Crashes or freezes
  - [ ] Controls responsiveness (see [MOBILE_CONTROLS.md](MOBILE_CONTROLS.md))
  - [ ] Multiplayer connectivity
  - [ ] Mod browser functionality
- [ ] Fix critical bugs and re-upload if needed

### 1.6. iOS Submission

#### Pre-Submission Checklist
- [ ] All TestFlight testing complete (minimum 3 testers, 7 days recommended)
- [ ] No critical crashes in TestFlight crash reports
- [ ] All screenshots uploaded and formatted correctly
- [ ] Privacy policy URL accessible: `https://your-domain.com/privacy`
- [ ] Support URL accessible: `https://your-domain.com/support`
- [ ] Age rating questionnaire completed
- [ ] App Store description proofread (no typos or policy violations)

#### Submit for Review
- [ ] Run: `bundle exec fastlane ios release`
- [ ] Alternatively, submit manually in App Store Connect:
  - [ ] Navigate to **App Store** → **iOS App** → **+ Version**
  - [ ] Enter version number (e.g., 1.0.0)
  - [ ] Upload build via Xcode or Fastlane
  - [ ] Fill in **What's New in This Version** (release notes)
  - [ ] Select **Manual Release** (recommended for first release)
  - [ ] Click **Submit for Review**

#### Review Monitoring
- [ ] Monitor App Store Connect for status updates:
  - **Waiting for Review**: Queued (1-7 days typical)
  - **In Review**: Being tested by Apple (1-2 days)
  - **Pending Developer Release**: Approved (ready to publish)
  - **Ready for Sale**: Live on App Store
  - **Rejected**: Review feedback provided
- [ ] If rejected, address issues and resubmit
- [ ] Common rejection reasons:
  - Crashes during review (fix and resubmit)
  - Incomplete metadata or screenshots
  - Misleading description or screenshots
  - Privacy policy missing or incomplete
  - In-app purchase not clearly cosmetic-only

---

## Phase 2: Android Submission

### 2.1. Google Play Console Setup

#### Create Application
- [ ] Log in to [Google Play Console](https://play.google.com/console)
- [ ] Click **Create app**
- [ ] App details:
  - **App name**: Violence
  - **Default language**: English (United States)
  - **App or game**: Game
  - **Free or paid**: Free
  - **Declarations**: Complete all required declarations
- [ ] Accept Google Play Developer Program Policies
- [ ] Click **Create app**

#### Dashboard Setup
- [ ] Navigate to **Dashboard** → complete all required tasks:
  - [ ] Set up app details
  - [ ] Provide store listing details
  - [ ] Upload app icon (512×512 PNG)
  - [ ] Upload feature graphic (1024×500 PNG)
  - [ ] Provide content rating (IARC questionnaire)
  - [ ] Set target audience and content
  - [ ] Select app category

### 2.2. Google Play API Access

#### Service Account Creation
- [ ] In Google Play Console, navigate to **Setup → API access**
- [ ] Click **Create new service account**
- [ ] Follow link to Google Cloud Console
- [ ] Create service account:
  - **Name**: `violence-fastlane-deploy`
  - **Description**: "Fastlane deployment automation"
  - **Role**: Service Account User
- [ ] Click **Create and Continue**
- [ ] Grant access: **Service Accounts → Actions → Manage keys**
- [ ] Add key → **Create new key → JSON**
- [ ] Download JSON key file
- [ ] Save as `android/google-play-api-key.json` (gitignored)

#### Grant Permissions
- [ ] Return to Google Play Console → **API access**
- [ ] Find service account in list
- [ ] Click **Grant access**
- [ ] Permissions:
  - [x] Releases: Create, read, and manage releases
  - [x] Release to production: Create and manage production releases
- [ ] Invite user

### 2.3. Signing Key Setup

#### Generate Release Keystore
- [ ] Run: `bundle exec fastlane android generate_keystore`
- [ ] Or manually:
  ```bash
  keytool -genkey -v \
    -keystore android/violence-release.keystore \
    -alias violence \
    -keyalg RSA \
    -keysize 2048 \
    -validity 10000
  ```
- [ ] Enter and record keystore password (store in password manager)
- [ ] Enter and record key password (store in password manager)
- [ ] Fill in certificate details (organization: OPD AI)

#### Backup Keystore Securely
- [ ] Copy `android/violence-release.keystore` to secure backup location
- [ ] Document passwords in team password manager
- [ ] **CRITICAL**: Losing keystore prevents future app updates

#### Environment Variables
- [ ] Edit `fastlane/.env` with Android values:
  ```bash
  # Android Configuration
  ANDROID_PACKAGE_NAME=ai.opd.violence
  ANDROID_KEYSTORE_PATH=../android/violence-release.keystore
  ANDROID_KEYSTORE_PASSWORD=your_keystore_password
  ANDROID_KEY_ALIAS=violence
  ANDROID_KEY_PASSWORD=your_key_password
  GOOGLE_PLAY_JSON_KEY_PATH=../android/google-play-api-key.json
  ```

### 2.4. Android Project Setup

#### Gradle Configuration
- [ ] Verify `android/build.gradle` exists with correct package name: `ai.opd.violence`
- [ ] Update `versionCode` and `versionName` in `android/app/build.gradle`:
  ```gradle
  android {
      defaultConfig {
          applicationId "ai.opd.violence"
          versionCode 1       // Increment for each release
          versionName "1.0.0" // Semantic version
      }
  }
  ```

#### Link gomobile .aar
- [ ] Build Violence.aar: `gomobile bind -target=android -o android/app/libs/violence.aar .`
- [ ] Add to `android/app/build.gradle`:
  ```gradle
  dependencies {
      implementation(name:'violence', ext:'aar')
  }
  repositories {
      flatDir {
          dirs 'libs'
      }
  }
  ```

### 2.5. Google Play Metadata

#### Store Listing
All metadata is in `fastlane/metadata/android/en-US/`:

- [ ] Verify `title.txt`: "Violence" (max 50 characters)
- [ ] Verify `short_description.txt`: Max 80 characters
- [ ] Verify `full_description.txt`: Max 4000 characters (HTML formatted)
- [ ] Verify `video.txt`: YouTube trailer URL (optional)

#### Graphics Assets
Prepare in `fastlane/metadata/android/en-US/images/`:

- [ ] **Icon**: 512×512 PNG (32-bit PNG with transparency)
- [ ] **Feature graphic**: 1024×500 PNG (no transparency)
- [ ] **Phone screenshots**: 16:9 or 9:16 aspect ratio (minimum 2, maximum 8)
  - Resolution: 320px to 3840px
  - Suggested: 1080×1920 (portrait) or 1920×1080 (landscape)
- [ ] **7-inch tablet screenshots**: Optional
- [ ] **10-inch tablet screenshots**: Optional

**Screenshot Requirements**:
- Minimum 2 phone screenshots required
- Show actual in-game content (no marketing graphics)
- Must accurately represent app functionality

#### Content Rating (IARC)
- [ ] Review [AGE_RATINGS.md](AGE_RATINGS.md) for IARC questionnaire
- [ ] In Google Play Console → **Content rating** → **Start questionnaire**
- [ ] Email address: Your developer email
- [ ] App category: Game
- [ ] Answer questionnaire:
  - **Violence**: YES (fantasy violence, frequent, moderate intensity)
  - **Sexual content**: NO
  - **Language**: NO (chat has profanity filter)
  - **Controlled substances**: NO
  - **Gambling**: NO (cosmetic loot boxes only)
  - **User interaction**: YES (text chat, user-generated content via mods)
  - **Shares location**: NO
  - **In-app purchases**: YES (cosmetic items)
- [ ] Submit questionnaire
- [ ] Verify rating: **ESRB TEEN** or **PEGI 12**

#### Target Audience
- [ ] Navigate to **Target audience and content**
- [ ] Target age group: **13+** (aligned with ESRB TEEN)
- [ ] Store presence: Include app in **Family** category: NO (violence content)

#### Privacy Policy
- [ ] Upload privacy policy URL: `https://your-domain.com/privacy`
- [ ] Ensure accessible and matches [PRIVACY_POLICY.md](PRIVACY_POLICY.md)

### 2.6. Android Build and Test

#### Build .aab
- [ ] Run: `bundle exec fastlane android build`
- [ ] Verify output: `android/app/build/outputs/bundle/release/app-release.aab`
- [ ] Check .aab size: `ls -lh android/app/build/outputs/bundle/release/app-release.aab`
- [ ] Recommended max: 150MB (Google Play limit: 150MB for .aab)

#### Internal Testing Track
- [ ] Run: `bundle exec fastlane android internal`
- [ ] Or manually upload in Google Play Console:
  - [ ] Navigate to **Testing → Internal testing**
  - [ ] Create new release
  - [ ] Upload `app-release.aab`
  - [ ] Enter release notes
  - [ ] Review release → **Save** → **Review release** → **Start rollout to Internal testing**

#### Add Testers
- [ ] Navigate to **Internal testing → Testers**
- [ ] Create email list of internal testers
- [ ] Share opt-in URL with testers
- [ ] Testers install from Google Play Store

#### Pre-Launch Report
- [ ] Google Play automatically tests on ~20 devices
- [ ] Wait 24-48 hours for pre-launch report
- [ ] Review report: **Release → Pre-launch report**
- [ ] Check for:
  - [ ] Crashes (should be zero)
  - [ ] ANRs (Application Not Responding - should be zero)
  - [ ] Security vulnerabilities
  - [ ] Accessibility issues
- [ ] Fix critical issues and re-upload if needed

### 2.7. Android Submission

#### Production Release
- [ ] Verify all store listing metadata complete (100% in Dashboard)
- [ ] Verify content rating received
- [ ] Verify privacy policy URL active
- [ ] Review [FASTLANE_ANDROID.md](FASTLANE_ANDROID.md) production checklist

#### Upload to Production (Staged Rollout)
- [ ] Run: `bundle exec fastlane android release`
- [ ] Or manually:
  - [ ] Navigate to **Production → Create new release**
  - [ ] Upload `app-release.aab`
  - [ ] Release name: Version number (e.g., "1.0.0")
  - [ ] Release notes: What's new in this version
  - [ ] Rollout percentage: **10%** (recommended for first release)
  - [ ] Click **Review release** → **Start rollout to Production**

#### Monitor Rollout
- [ ] Check **Production** dashboard for rollout status
- [ ] Monitor crash reports: **Quality → Android vitals → Crashes & ANRs**
- [ ] Monitor user reviews: **Ratings and reviews**
- [ ] If stable after 24-48 hours, increase rollout:
  - [ ] Run: `bundle exec fastlane android promote`
  - [ ] Or manually: **Production → Manage rollout → Update rollout → 100%**

#### First Release Approval
- [ ] Google Play reviews first release (1-7 days typical)
- [ ] Monitor email for approval or rejection notice
- [ ] If rejected, address issues in **Policy status** and resubmit

---

## Phase 3: Post-Launch

### 3.1. Monitoring

#### iOS App Analytics
- [ ] Monitor in App Store Connect → **Analytics**
- [ ] Track:
  - [ ] Downloads and installations
  - [ ] Crashes and crashes per session
  - [ ] User retention (Day 1, Day 7, Day 30)
  - [ ] In-app purchase conversion (if applicable)

#### Android App Analytics
- [ ] Monitor in Google Play Console → **Statistics**
- [ ] Track:
  - [ ] Installs and uninstalls
  - [ ] Crashes and ANRs (Android vitals)
  - [ ] Ratings and reviews
  - [ ] Acquisition reports (where users found app)

#### Crash Reporting
- [ ] Review TestFlight crash logs (iOS)
- [ ] Review Android vitals → Crashes & ANRs (Android)
- [ ] Prioritize high-frequency crashes
- [ ] Fix and release patch updates as needed

### 3.2. Updates

#### iOS Updates
- [ ] Increment version in Xcode: `Info.plist → CFBundleShortVersionString`
- [ ] Increment build number: `Info.plist → CFBundleVersion`
- [ ] Update release notes in `fastlane/metadata/ios/en-US/release_notes.txt`
- [ ] Run: `bundle exec fastlane ios release`
- [ ] Submit for review in App Store Connect

#### Android Updates
- [ ] Increment `versionCode` in `android/app/build.gradle` (must be higher than previous)
- [ ] Update `versionName` if semantic version changed
- [ ] Update release notes in `fastlane/metadata/android/en-US/changelogs/[versionCode].txt`
- [ ] Run: `bundle exec fastlane android release`
- [ ] Review and publish in Google Play Console

### 3.3. User Feedback

#### Responding to Reviews
- [ ] Monitor App Store reviews daily (first 2 weeks)
- [ ] Respond to negative reviews constructively
- [ ] Thank users for positive feedback
- [ ] Do NOT argue or be defensive in responses

#### Feature Requests
- [ ] Track common feature requests in GitHub issues
- [ ] Prioritize based on frequency and impact
- [ ] Communicate roadmap updates via release notes

### 3.4. Compliance

#### Privacy Policy Updates
- [ ] Review [PRIVACY_POLICY.md](PRIVACY_POLICY.md) for accuracy
- [ ] Update if data collection practices change
- [ ] Notify users via in-app message if material changes

#### Age Rating Updates
- [ ] Re-submit age rating questionnaires if game content changes:
  - New violence intensity (gore added)
  - Language content changes (profanity in dialogue)
  - New monetization (real-money gambling mechanics)
- [ ] Update [AGE_RATINGS.md](AGE_RATINGS.md) if ratings change

---

## Troubleshooting

### iOS Issues

#### "Code signing failed"
- **Solution**: Run `bundle exec fastlane ios sync_signing` to re-sync certificates
- Verify Team ID and provisioning profile in Xcode

#### "Build validation failed"
- **Solution**: Check for missing architectures: `lipo -info Violence.framework/Violence`
- Ensure gomobile built for both arm64 and x86_64 (simulator)

#### "Rejected: Guideline 2.1 - Performance - App Completeness"
- **Solution**: App crashed during review. Check TestFlight crash logs, fix, and resubmit

#### "Rejected: Guideline 4.0 - Design"
- **Solution**: Screenshots or metadata don't match app functionality. Update screenshots to show actual gameplay

### Android Issues

#### "Upload failed: Version code already exists"
- **Solution**: Increment `versionCode` in `android/app/build.gradle`

#### "Pre-launch report shows crashes"
- **Solution**: Check stack traces in pre-launch report. Common issues:
  - Missing permissions in `AndroidManifest.xml`
  - Unsupported API levels (minSdkVersion too low)
  - Native library architecture mismatch

#### "API authentication failed"
- **Solution**: Verify Google Play API JSON key path in `fastlane/.env`
- Check service account has "Release Manager" role in Google Play Console

#### "Rejected: Policy violation"
- **Solution**: Review email for specific policy violation (e.g., content rating mismatch, misleading metadata)
- Common fixes:
  - Correct content rating via IARC questionnaire
  - Remove misleading claims from description
  - Add missing privacy policy disclosures

---

## Security Checklist

### Credentials Protection
- [ ] Never commit `fastlane/.env` to Git (verify in `.gitignore`)
- [ ] Never commit keystore files to Git
- [ ] Never commit Google Play API JSON key to Git
- [ ] Store all passwords in team password manager (1Password, LastPass, etc.)

### Certificate Backup
- [ ] Backup iOS certificates in Match repository (encrypted)
- [ ] Backup Android keystore to secure offline location
- [ ] Document keystore passwords in password manager
- [ ] Test certificate/keystore restoration process

### API Key Rotation
- [ ] Rotate App Store Connect API keys annually
- [ ] Revoke old keys in App Store Connect after rotation
- [ ] Rotate Google Play API keys if compromised

---

## CI/CD Automation (Optional)

### GitHub Actions Setup
- [ ] Create `.github/workflows/ios-release.yml` for iOS automation
- [ ] Create `.github/workflows/android-release.yml` for Android automation
- [ ] Add secrets to GitHub repository settings:
  - iOS: `APP_IDENTIFIER`, `APPLE_ID`, `TEAM_ID`, `MATCH_PASSWORD`, `ASC_KEY_CONTENT`
  - Android: `ANDROID_KEYSTORE_BASE64`, `GOOGLE_PLAY_JSON_BASE64`, keystore passwords
- [ ] Test CI/CD with tag push: `git tag v1.0.0 && git push --tags`

### GitLab CI Setup
- [ ] Create `.gitlab-ci.yml` with iOS and Android deploy jobs
- [ ] Add CI/CD variables in GitLab project settings
- [ ] Test pipeline with tag push

See [FASTLANE_IOS.md](FASTLANE_IOS.md) and [FASTLANE_ANDROID.md](FASTLANE_ANDROID.md) for example CI/CD configurations.

---

## Related Documentation

- [MOBILE_BUILD.md](MOBILE_BUILD.md) - Gomobile build pipeline
- [MOBILE_CONTROLS.md](MOBILE_CONTROLS.md) - Touch controls implementation
- [FASTLANE_IOS.md](FASTLANE_IOS.md) - iOS Fastlane configuration
- [FASTLANE_ANDROID.md](FASTLANE_ANDROID.md) - Android Fastlane configuration
- [AGE_RATINGS.md](AGE_RATINGS.md) - Age rating questionnaire responses
- [PRIVACY_POLICY.md](PRIVACY_POLICY.md) - Privacy policy template
- [Official Fastlane Docs](https://docs.fastlane.tools/)
- [App Store Connect Help](https://help.apple.com/app-store-connect/)
- [Google Play Console Help](https://support.google.com/googleplay/android-developer/)

---

## Revision History

| Date       | Version | Changes                                      |
|------------|---------|----------------------------------------------|
| 2026-03-05 | 1.0     | Initial mobile publishing submission guide   |
