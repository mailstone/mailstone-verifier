# Changelog

All notable changes to MailStone Verifier are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2026-05-01

### Added

- Initial public release. Version aligned with the MailStone V2 platform release.
- **Hash Calculator** tab — SHA-256 hashing of any local file (Email PDF, attachment, ACK JSON).
- **TimeStamp Decoder** tab — Decode Base64-encoded RFC 3161 TSA tokens; surfaces provider, UTC date, serial number, hash algorithm, and grant status.
- **Merkle Verifier** tab — Reconstruct the Merkle root from a proof JSON and validate that a given hash is included.
- Standalone binaries for macOS (Intel + Apple Silicon), Linux x64, and Windows x64.
- Build tooling: `Makefile`, `setup.sh`, `setup-webkit.sh`.
- Documentation: README, BUILD, SECURITY, CONTRIBUTING.
- MIT License.
