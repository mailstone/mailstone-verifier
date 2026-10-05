# Changelog

All notable changes to MailStone Verifier are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2026-10-05

### Added

- **ERE Verification** tab for Registered Electronic Deliveries (Envoi Recommandé Électronique):
  - **Recipient decision signature** — verify the Ed25519 signature printed in the Dossier de preuve against the recipient's public key and the exact signed message (pasted as printed, or rebuilt from `ere_id`, `decision`, `decided_at`). The tool answers true or false; nothing is recomputed by hand.
  - **Event evidence hash** — recompute the Merkle leaf of each event of a delivery (deposit receipt, dispatch, first and later presentations, decision, cancellation, expiry) from the facts printed in the proof, with the platform's exact canonical formats, then verify it in the Merkle tab against that event's proof block.
- **TimeStamp Decoder**: optional "hash the token should cover" field — reports whether the token's imprint equals a Merkle root or a file hash, which is what links a timestamp to a batch.
- **TimeStamp Decoder**: signer certificate (subject, issuer, validity window) and an explicit note that the chain of trust is the reader's to confirm.
- Real-data regression tests: a decision signed on the MailStone platform, platform-anchored leaves for four event types, and a genuine root timestamp token (verified, covered root, tamper detection).

### Changed

- **Merkle Verification**: the hash is optional — with an empty field the block's own leaf is verified, which is how ERE proof blocks (one leaf each) are meant to be used.

### Fixed

- **TimeStamp Decoder** no longer reports a token as decoded without checking it: the RSASSA-PSS path (MailStone TimeStamp) now verifies the PKCS#7 signature against the embedded certificate, like the standard path always did, and says so. A token with no embedded certificate is flagged as unverifiable.

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
