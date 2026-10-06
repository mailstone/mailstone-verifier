# Build Guide

Build, develop, and release MailStone Verifier from source.

---

## Prerequisites

### Go 1.24+

Install from [golang.org/dl](https://golang.org/dl/) and verify:

```bash
go version
```

### Wails CLI v2

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH="$PATH:$(go env GOPATH)/bin"
wails version
```

Make the `PATH` change permanent by adding `export PATH="$PATH:$(go env GOPATH)/bin"` to `~/.bashrc` or `~/.zshrc`.

### Platform dependencies

#### macOS

```bash
xcode-select --install
```

No other dependency.

#### Linux — Ubuntu / Debian / Mint

For 22+ (libwebkit2gtk-4.1):
```bash
sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.1-dev
```

For older releases (libwebkit2gtk-4.0):
```bash
sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
```

If Wails complains about `webkit2gtk-4.0` while you only have `4.1` installed, run `./setup-webkit.sh` to create the compatibility symlink.

#### Linux — Fedora / RHEL

```bash
sudo dnf install -y gtk3-devel webkit2gtk3-devel
```

#### Linux — Arch / Manjaro

```bash
sudo pacman -S gtk3 webkit2gtk
```

#### Windows

- [Microsoft WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
- A C toolchain such as [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) (CGO is required)

### Automated setup (Linux/macOS)

The repo ships with a helper that installs Wails, Go module dependencies and platform packages in one go:

```bash
./setup.sh
```

---

## Build

### Current platform

```bash
make build
```

The binary is written to `build/bin/mailstone-verifier` (`.exe` on Windows).

### All platforms

```bash
make build-all
```

Produces:

- `mailstone-verifier-darwin-amd64` (macOS Intel)
- `mailstone-verifier-darwin-arm64` (macOS Apple Silicon)
- `mailstone-verifier-linux-amd64` (Linux x64)
- `mailstone-verifier-windows-amd64.exe` (Windows x64)

Windows cross-compiles from Linux (`wails build -platform windows/amd64`) — keep `github.com/wailsapp/go-webview2` at the version Wails' own `go.mod` requires, otherwise the Windows frontend fails to compile. macOS cannot be cross-compiled: build it on a Mac, or let `.github/workflows/release.yml` do it on a tag. Cross-compilation otherwise requires the relevant platform toolchains. If a target fails locally, build it on the matching host (or in a Docker container) instead.

### Manual `wails build`

If you would rather skip the Makefile:

```bash
wails build                                                        # current platform
wails build -platform darwin/arm64 -o mailstone-verifier-darwin-arm64
wails build -platform linux/amd64  -o mailstone-verifier-linux-amd64
wails build -platform windows/amd64 -o mailstone-verifier-windows-amd64.exe
wails build -ldflags "-s -w"                                       # strip debug symbols
```

### Clean build artefacts

```bash
make clean
```

---

## Development

Hot-reload mode (devtools open, frontend changes reload automatically):

```bash
make dev
# or
wails dev
```

### Project layout

```
mailstone-verifier/
├── main.go                 # Wails entry point
├── app.go                  # Backend API (CalculateHash, DecodeTimestamp, VerifyMerkle)
├── internal/
│   ├── hasher/             # SHA-256 hasher
│   ├── timestamp/          # RFC 3161 decoder
│   └── merkle/             # Merkle tree verifier
├── frontend/
│   ├── index.html          # UI with 3 tabs
│   ├── style.css
│   ├── app.js
│   └── assets/
├── go.mod / go.sum
├── wails.json              # Wails configuration
├── Makefile
├── setup.sh / setup-webkit.sh
└── README.md / BUILD.md / LICENSE
```

### Tests

```bash
go test ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

---

## Release

1. **Tag the version**

   ```bash
   git tag -a v2.0.0 -m "Release v2.0.0"
   git push origin v2.0.0
   ```

2. **Clean and build everything**

   ```bash
   make clean
   make build-all
   ```

3. **Sign the binaries**
   - macOS: `codesign` (and notarize if you need Gatekeeper-clean distribution)
   - Windows: `signtool`
   - Linux: optional GPG signature alongside the archives

4. **Package**

   ```bash
   cd build/bin
   tar -czf mailstone-verifier-v2.0.0-darwin-amd64.tar.gz  mailstone-verifier-darwin-amd64
   tar -czf mailstone-verifier-v2.0.0-darwin-arm64.tar.gz  mailstone-verifier-darwin-arm64
   tar -czf mailstone-verifier-v2.0.0-linux-amd64.tar.gz   mailstone-verifier-linux-amd64
   zip       mailstone-verifier-v2.0.0-windows-amd64.zip   mailstone-verifier-windows-amd64.exe
   ```

5. **Generate checksums**

   ```bash
   sha256sum *.tar.gz *.zip > SHA256SUMS.txt
   ```

6. **Publish** the archives + `SHA256SUMS.txt` on the GitHub Releases page.

The binaries embed the entire frontend (HTML/CSS/JS) — no runtime dependency is required at install time other than the system WebView (WebKit on Linux/macOS, WebView2 on Windows).

---

## Troubleshooting

### `wails: command not found`

`wails` lives in `$(go env GOPATH)/bin`. Make sure that directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Persist it by appending the same line to `~/.bashrc` / `~/.zshrc`.

### `Package libwebkit2gtk-4.1-dev not found` (older Linux)

Use the 4.0 package on older distributions:

```bash
sudo apt-get install -y libwebkit2gtk-4.0-dev
```

### `Package webkit2gtk-4.0 not found` (newer Linux with 4.1 only)

Run `./setup-webkit.sh` — it creates a `pkg-config` shim that maps `webkit2gtk-4.0` to the installed `4.1` package so Wails can link.

### `undefined reference to ...` (Linux)

Missing build basics:

```bash
sudo apt-get install -y gcc pkg-config
```

### macOS: "App is damaged and can't be opened"

The binary was downloaded from outside a notarized channel and Gatekeeper quarantined it. Strip the quarantine attribute:

```bash
xattr -cr ./mailstone-verifier-darwin-arm64
```

For distribution to other users, prefer signing + notarization rather than asking them to bypass Gatekeeper.

### Cross-compile target fails

Cross-compiling Wails apps requires the target platform's CGO toolchain. If a target fails:

- Build on the matching host, **or**
- Use a CI runner / Docker image that ships the right toolchain (the GitHub-hosted `ubuntu-latest`, `macos-latest`, `windows-latest` runners cover the four targets we ship).
