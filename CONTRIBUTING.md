# Contributing to MailStone Verifier

Thanks for taking the time to help. The Verifier exists so anyone can independently audit MailStone proofs — every improvement to its accuracy, clarity, or platform support is welcome.

## Ways to contribute

- **Bug reports** — open a GitHub issue with reproduction steps, your platform, the Verifier version (`mailstone-verifier --version` once available, or the binary filename), and the proof JSON / timestamp blob you were verifying when it broke (sanitised of any private content).
- **Feature requests** — open an issue describing the use case before sending a PR for anything non-trivial.
- **Pull requests** — see below.
- **Security issues** — see [SECURITY.md](SECURITY.md). Do not open a public issue.

## Pull request workflow

1. Fork the repository and create a topic branch:

   ```bash
   git checkout -b feature/short-description
   ```

2. Make your changes. Keep commits focused; squash noise out before opening the PR.

3. Run the test suite:

   ```bash
   go test ./...
   ```

4. Run the app locally to confirm it still launches and the affected tab works end-to-end:

   ```bash
   make dev
   ```

5. Commit with a descriptive message (Conventional Commits style is appreciated but not required):

   ```
   feat(merkle): support odd-leaf-count duplication
   fix(timestamp): handle missing serial number gracefully
   docs(readme): clarify Linux PATH setup
   ```

6. Open a PR against `master`. In the description, explain *what* changed and *why*, and link any related issue.

## Code conventions

- **Go**: standard `gofmt` / `go vet`. No external linter is required, but PRs that fail `go vet` will be asked to clean up.
- **Frontend**: vanilla HTML/CSS/JS, no build step. Keep it that way unless there is a strong reason to introduce a bundler.
- **Comments**: explain the *why* (an invariant, a non-obvious crypto rule, a platform quirk). Don't restate the *what* — well-named identifiers handle that.
- **Tests**: every new parsing path or verification rule should come with a unit test that demonstrates both the success and a representative failure mode.

## Scope discipline

The Verifier is intentionally minimal:

- It runs **offline**.
- It depends only on Go's standard library plus `github.com/digitorus/timestamp`, `github.com/smallstep/pkcs7`, and `github.com/wailsapp/wails/v2`.
- It does not embed MailStone-specific secrets, branding theming, or telemetry.

PRs that add network calls, telemetry, analytics, or proprietary dependencies will be declined. Changes to the verification logic must be backed by the relevant standard (RFC 3161, RFC 6234 for SHA-256, the proof-document spec linked from the README).

## License

By contributing you agree that your changes will be released under the [MIT License](LICENSE) that already covers the project.
