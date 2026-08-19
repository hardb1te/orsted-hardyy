# Changelog

All notable changes to Orsted C2 are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/).

---

## [1.1.0] - 2026-08-20

### Added

#### Two-Stage Stager Delivery System
- **New CLI command**: `generate stager` — produces minimal stagers (~50KB) for Windows x86/x64
- **Stager binaries**: `main_stager_http.go` and `main_stager_https.go` with registration, download, and in-memory execution
- **Server endpoint**: `/agent/download/{beacon_id}` — authenticates beacons and serves full agent binaries on-demand
- **Architecture support**: Stagers support both 32-bit and 64-bit targets via `--arch` flag
- **Proxy support**: HTTP/HTTPS proxy configuration for stager callbacks and downloads
- **Retry logic**: Exponential backoff (2s/4s/8s) for download failures
- **PE validation**: Stagers validate downloaded binaries before execution

### Modified

#### Beacon Generation (`beacon-generate.go`)
- Added `--arch` flag (short: `-r`) with values `32` and `64` (default: `64`)
- Output filenames now include architecture suffix: `main_<type>_<arch>.exe`
- Support for 32-bit cross-compilation: `GOARCH=386`, `CC=i686-w64-mingw32-gcc`
- Existing 64-bit default behavior preserved (backward compatible)

### Documentation

- Added stager usage guide to README.md
- Created FEATURE_STANDARD.md for consistent feature development

---

## [1.0.0] - Earlier

- Original Orsted C2 implementation with beacon registration, task handling, and modular payload system
