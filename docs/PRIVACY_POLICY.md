# Privacy Policy for VIOLENCE

**Effective Date:** March 5, 2026  
**Last Updated:** March 5, 2026

## Introduction

VIOLENCE ("we," "our," or "the app") is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our mobile application and related services.

## Information We Collect

### Information You Provide

- **Account Information:** If you create an account for cloud save synchronization, we collect your chosen username and password (stored as a cryptographic hash).
- **Cloud Save Data:** Game progress, settings, and save files that you choose to sync to your configured cloud storage provider (S3, WebDAV, or custom backend).
- **Mod Marketplace:** If you upload mods, we collect your author name, mod metadata, and WASM files.

### Automatically Collected Information

- **Gameplay Telemetry:** Anonymous gameplay statistics including genre selection, level completion rates, and performance metrics (opt-in only).
- **Multiplayer Data:** Player username, match statistics, and leaderboard rankings when participating in online play.
- **Device Information:** Device type, operating system version, screen resolution, and app version for crash reporting and compatibility optimization.
- **Network Information:** IP address for multiplayer matchmaking and federation hub discovery (not stored long-term).

### Information We Do NOT Collect

- We do not collect personally identifiable information (PII) beyond what you explicitly provide.
- We do not track your location.
- We do not access your camera, microphone, contacts, or other device sensors beyond game controls.
- We do not use third-party advertising or analytics SDKs.

## How We Use Your Information

We use collected information for:

1. **Game Functionality:** Providing core features like save synchronization, multiplayer matchmaking, and mod distribution.
2. **Service Improvement:** Analyzing anonymous gameplay data to improve balance, performance, and user experience.
3. **Community Features:** Enabling leaderboards, achievements, and replay sharing.
4. **Security:** Detecting cheating, preventing abuse, and enforcing fair play policies.
5. **Technical Support:** Diagnosing crashes and technical issues you report.

## Data Storage and Security

### Cloud Saves

- Cloud save data is stored in **your configured storage provider** (AWS S3, Backblaze B2, MinIO, Nextcloud, etc.).
- We provide **AES-256-GCM encryption** for all cloud-stored saves, with keys derived from your password using PBKDF2 (100,000 iterations).
- We do not have access to your cloud storage credentials or decryption keys.

### Mod Marketplace

- Mod metadata is stored in our registry database with SHA256 checksums for integrity verification.
- WASM mod files are virus-scanned (placeholder heuristic) before acceptance.
- Author names and upload timestamps are publicly visible.

### Multiplayer Data

- Match statistics and leaderboard entries are stored on our federation hub servers.
- Anti-cheat data (replay checksums, timing heuristics) is retained for 90 days.
- IP addresses for matchmaking are not stored beyond active sessions.

### Security Measures

We implement industry-standard security practices:

- **Encryption in Transit:** All network communication uses TLS 1.3.
- **Encryption at Rest:** Cloud saves use AES-256-GCM with user-controlled keys.
- **Access Controls:** Database access is restricted to authorized personnel only.
- **Regular Audits:** Security reviews and dependency updates are performed quarterly.

## Data Sharing and Disclosure

We **do not sell or rent** your information to third parties. We may share data only in these circumstances:

1. **With Your Consent:** When you explicitly authorize sharing (e.g., publishing replay files).
2. **Service Providers:** Cloud storage providers you configure (AWS, Backblaze, Nextcloud, etc.) receive only encrypted save data.
3. **Legal Obligations:** If required by law, court order, or government regulation.
4. **Safety and Security:** To prevent fraud, abuse, or violations of our Terms of Service.

## Your Rights and Choices

### Data Access and Deletion

- **Cloud Saves:** You have full control via your cloud provider's interface. Deleting files from your S3 bucket or WebDAV folder immediately removes them.
- **Mod Marketplace:** Contact us at privacy@opd.ai to request removal of uploaded mods.
- **Leaderboards:** You may opt out of public leaderboard display in the app settings.
- **Telemetry:** Disable anonymous gameplay statistics in Settings > Privacy.

### Account Deletion

To delete your account and associated data:

1. Remove cloud save sync credentials from the app settings.
2. Delete local save files via the in-app save manager.
3. Email privacy@opd.ai to request removal of leaderboard entries and mod uploads.

We will process deletion requests within 30 days.

## Children's Privacy

VIOLENCE is rated for ages 12+ (ESRB: T for Teen) due to fantasy/sci-fi violence. We do not knowingly collect information from children under 13 without parental consent. If you believe we have inadvertently collected such data, contact us immediately.

## International Data Transfers

If you are located outside the United States, your information may be transferred to and processed in the U.S. or other countries where our servers are located. We ensure appropriate safeguards are in place for cross-border data transfers.

## Third-Party Services

### Cloud Storage Providers

When you configure cloud save sync, you directly interact with third-party providers:

- **AWS S3:** [AWS Privacy Policy](https://aws.amazon.com/privacy/)
- **Backblaze B2:** [Backblaze Privacy Policy](https://www.backblaze.com/company/privacy.html)
- **Nextcloud/ownCloud:** Self-hosted or third-party instances (check your provider's policy)

We are not responsible for the privacy practices of these third parties.

### Federation Hubs

Third-party federation hubs (multiplayer servers run by the community) may collect additional data. Review each hub's privacy policy before connecting.

## Changes to This Privacy Policy

We may update this policy periodically. Changes will be posted in-app and on our website with an updated "Last Updated" date. Continued use after changes constitutes acceptance.

## Contact Us

For privacy-related questions, concerns, or data requests:

- **Email:** privacy@opd.ai
- **GitHub Issues:** https://github.com/opd-ai/violence/issues
- **Mailing Address:**  
  OPD.AI  
  Attn: Privacy Officer  
  [Your Address]

## Compliance

This Privacy Policy complies with:

- California Consumer Privacy Act (CCPA)
- General Data Protection Regulation (GDPR)
- Children's Online Privacy Protection Act (COPPA)
- Apple App Store Guidelines
- Google Play Store Policies

## Your California Privacy Rights

California residents have additional rights under the CCPA:

- **Right to Know:** Request disclosure of data collected, sources, purposes, and third-party sharing.
- **Right to Delete:** Request deletion of personal information (subject to legal exceptions).
- **Right to Opt-Out:** Opt out of data "sales" (we do not sell data).
- **Non-Discrimination:** We will not discriminate against you for exercising these rights.

To exercise these rights, email privacy@opd.ai with "CCPA Request" in the subject line.

## Data Retention

- **Cloud Saves:** Retained indefinitely in your cloud storage until you delete them.
- **Mod Uploads:** Retained indefinitely unless you request removal.
- **Leaderboard Data:** Retained for the current season (90 days) plus historical archives.
- **Crash Reports:** Retained for 12 months.
- **Multiplayer Sessions:** IP addresses discarded immediately after session ends; match statistics retained for 90 days.

---

**By using VIOLENCE, you acknowledge that you have read and understood this Privacy Policy.**