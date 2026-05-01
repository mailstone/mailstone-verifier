# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in MailStone Verifier, **please do not open a public GitHub issue**. Public disclosure before a fix is available puts every user of the tool at risk.

Instead, email **support@mailstone.fr** with:

- A clear description of the issue
- Steps to reproduce (or a proof-of-concept)
- The affected version(s) and platform(s)
- Your assessment of the impact
- Whether you wish to be credited (and how) once the issue is disclosed

We will acknowledge your report within **5 business days**, keep you updated as we investigate, and credit you in the release notes once the fix ships unless you prefer otherwise.

## Scope

This policy covers the MailStone Verifier desktop application — the source in this repository. The MailStone platform itself (API, ingestion, anchoring services) is out of scope for this repo; please refer to the contact details on [mailstone.io](https://mailstone.io) for platform issues.

## What MailStone Verifier does — and does not — protect

The Verifier is an **offline cryptographic auditor**. It computes SHA-256 hashes, parses RFC 3161 timestamps, and reconstructs Merkle proofs locally on your machine. It does not connect to any MailStone server, does not transmit your files anywhere, and does not require credentials.

- **In scope**: parsing bugs that mis-validate a forged proof, dependency vulnerabilities, code execution from a crafted input, supply-chain integrity of the published binaries.
- **Out of scope**: vulnerabilities in third-party blockchain explorers or TSA authorities, weaknesses in your operating system's WebView component, social-engineering attacks against MailStone customers.

## Supported Versions

Only the latest minor release line receives security updates. Older versions should be upgraded.
