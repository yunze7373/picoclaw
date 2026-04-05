# Android / Termux Deployment Guide

> PicoClaw on Android — run your AI assistant on a phone or tablet.

PicoClaw runs natively on Android via [Termux](https://termux.dev/), a terminal emulator that provides a Linux environment without rooting your device.

## Prerequisites

| Requirement | Version |
|------------|---------|
| Android | 7.0+ (API 24+) |
| Termux | 0.118+ (install from [F-Droid](https://f-droid.org/packages/com.termux/), **not** Play Store) |
| Architecture | ARM64 (aarch64) — most modern Android devices |
| Storage | ~50MB free (binary + workspace) |
| RAM | ~10MB base, <50MB with all optional features |

> ⚠️ **Install Termux from F-Droid only.** The Play Store version is outdated and no longer maintained.

## Quick Install (Pre-built Binary)

```bash
# 1. Update Termux packages
pkg update && pkg upgrade -y

# 2. Install dependencies
pkg install -y git wget

# 3. Download PicoClaw ARM64 binary
wget https://picoclaw.io/download/picoclaw-linux-arm64 -O picoclaw
chmod +x picoclaw

# 4. Move to PATH
mv picoclaw $PREFIX/bin/

# 5. Verify installation
picoclaw --version
```

## Build from Source

```bash
# 1. Install Go and build tools
pkg install -y golang git make

# 2. Clone the repository
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# 3. Build (pure Go, no CGO needed for core)
go build -o picoclaw ./cmd/picoclaw

# 4. Install
cp picoclaw $PREFIX/bin/
```

### Building with CGO Extensions (sqlite-vec)

For optional vector search support via `sqlite-vec`, CGO is required:

```bash
# Install C compiler
pkg install -y clang

# Build with CGO enabled
CGO_ENABLED=1 go build -tags cgo_sqlite -o picoclaw ./cmd/picoclaw
```

## Configuration

PicoClaw auto-detects Termux via the `$TERMUX_VERSION` and `$PREFIX` environment variables. Default config paths:

| Item | Path |
|------|------|
| Config file | `~/.config/picoclaw/config.toml` |
| Workspace | `~/.picoclaw/workspace/` |
| Memory DB | `~/.picoclaw/memory.db` |
| Logs | `~/.picoclaw/logs/` |

### Minimal Config Example

Create `~/.config/picoclaw/config.toml`:

```toml
# Minimal Termux configuration
[agents.defaults]
model = "gpt-4o-mini"  # Use a smaller model for resource efficiency
max_tokens = 4096
context_window = 16384
max_tool_iterations = 10

# Memory optimization for Android
[agents.defaults.subturn]
max_concurrent = 2          # Limit parallel SubTurns on mobile
max_depth = 2               # Reduce nesting depth
default_timeout_minutes = 3 # Shorter timeout for mobile

# API key (use environment variable instead for security)
# export OPENAI_API_KEY="sk-..."

[tools.exec]
enabled = true

[tools.read_file]
enabled = true

[tools.write_file]
enabled = true

[tools.spawn]
enabled = true

[tools.subagent]
enabled = true

# Optional: team orchestration
[tools.team_create]
enabled = false  # Enable if you need multi-agent coordination
```

### Resource-Constrained Config

For devices with <1GB RAM available:

```toml
[agents.defaults]
model = "gpt-4o-mini"
max_tokens = 2048
context_window = 8192
max_tool_iterations = 5

[agents.defaults.subturn]
max_concurrent = 1
max_depth = 1
default_token_budget = 5000

# Disable heavy features
[tools.web]
enabled = false

[tools.web_fetch]
enabled = false

[tools.spawn]
enabled = false
```

## Auto-start with Termux:Boot

Install [Termux:Boot](https://f-droid.org/packages/com.termux.boot/) to start PicoClaw on device boot:

```bash
# 1. Install Termux:Boot from F-Droid
# 2. Open Termux:Boot once to initialize

# 3. Create boot script
mkdir -p ~/.termux/boot
cat > ~/.termux/boot/picoclaw.sh << 'EOF'
#!/data/data/com.termux/files/usr/bin/sh
termux-wake-lock
picoclaw serve &
EOF
chmod +x ~/.termux/boot/picoclaw.sh
```

## Process Management

### Using Termux services

```bash
# Start in background
picoclaw serve &

# Check if running
pgrep -f picoclaw

# Stop
pkill -f picoclaw
```

### Using a simple wrapper script

```bash
cat > ~/start-picoclaw.sh << 'EOF'
#!/bin/bash
while true; do
    picoclaw serve 2>&1 | tee -a ~/.picoclaw/logs/picoclaw.log
    echo "PicoClaw exited, restarting in 5s..."
    sleep 5
done
EOF
chmod +x ~/start-picoclaw.sh
```

## Storage Permissions

Termux uses its own internal storage by default. To access shared storage:

```bash
# Grant storage permission (one-time)
termux-setup-storage

# Shared storage will be at ~/storage/
# Example: access Downloads folder
ls ~/storage/downloads/
```

## Troubleshooting

### "Permission denied" when running binary

```bash
chmod +x $PREFIX/bin/picoclaw
```

### "cannot execute binary file"

Verify your device architecture:

```bash
uname -m
# Expected: aarch64 (ARM64)
```

If your device is `armv7l` (32-bit ARM), you need the ARM32 build or must build from source with `GOARCH=arm`.

### SQLite "database is locked"

Ensure only one PicoClaw instance is running:

```bash
pkill -f picoclaw
picoclaw serve
```

### High battery drain

```bash
# Use Termux wake-lock to prevent throttling
termux-wake-lock

# Reduce polling frequency in config
# Use smaller models (gpt-4o-mini instead of gpt-4o)
```

### Out of memory

Reduce resource usage in config (see Resource-Constrained Config above), or:

```bash
# Check memory usage
free -h

# Monitor PicoClaw specifically
ps aux | grep picoclaw
```

## Updating PicoClaw

```bash
# Self-update (if supported)
picoclaw update

# Or manually
wget https://picoclaw.io/download/picoclaw-linux-arm64 -O $PREFIX/bin/picoclaw
chmod +x $PREFIX/bin/picoclaw
```

## Known Limitations

- **No systemd**: Termux doesn't support systemd. Use Termux:Boot or wrapper scripts.
- **Background throttling**: Android may kill background processes. Use `termux-wake-lock`.
- **Filesystem**: Termux's filesystem is case-sensitive (unlike some Android storage).
- **Network**: Some corporate networks may block API calls. Configure `HTTP_PROXY` if needed.

---

*See also: [Configuration Guide](configuration.md) | [Hardware Compatibility](hardware-compatibility.md)*
