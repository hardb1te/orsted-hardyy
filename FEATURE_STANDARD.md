# Feature Development Standard

This document defines the standard for adding new features to Orsted C2, ensuring consistency across the codebase and clear documentation for operators and developers.

---

## Feature Development Checklist

Every new feature should follow this checklist to ensure quality and completeness.

### 1. Design & Specification

- [ ] **Write an Ideal State Criteria (ISC) document** listing what the feature must do
  - Problem statement
  - Vision/goal
  - Out of scope items
  - Acceptance criteria (atomic, testable, verifiable)
- [ ] **Identify affected components** (client CLI, beacon binary, server, protobuf, modules)
- [ ] **Assess backward compatibility** — does this break existing code?

### 2. Implementation

#### Client Changes (CLI)
- [ ] Add command or flag to `client/grumblecli/`
- [ ] Follow existing patterns (e.g., `SetGenerateBeaconCommand`, `SetListenerCommands`)
- [ ] Include help text and argument descriptions
- [ ] Add flag validation
- [ ] Command output should be clear and actionable (e.g., `[+] Feature Generated at ...`)

#### Beacon Changes
- [ ] New beacon binaries go in `beacon/main_*.go` with build tags
- [ ] Follow existing beacon patterns:
  - Import profiles, peers, utils, core packages
  - Initialize profile
  - Create peer (HTTP/HTTPS/TCP)
  - Register with server
  - Implement core loop (task retrieval, handling, reporting)
- [ ] Keep binaries minimal — only essential dependencies

#### Server Changes
- [ ] New handlers go in `server/listeners/` (http.go, https.go)
- [ ] Follow HTTP response patterns:
  - Set `Content-Type` header
  - Use protobuf for structured data
  - Return meaningful error responses (4xx/5xx with text)
- [ ] Register handlers in `addHttpHandler()` function
- [ ] Add logging via `utils.PrintDebug()` and `utils.PrintInfo()`
- [ ] Validate all input (beacon ID, query params, request body)

#### Database Changes (if needed)
- [ ] Queries in `server/orsteddb/`
- [ ] Error handling and meaningful messages

### 3. Testing

- [ ] Manual verification of CLI command
- [ ] Verify beacon binary builds and compiles
- [ ] Test server endpoint (curl or test tool)
- [ ] Verify anti-criteria (what must NOT happen)
  - No leaked secrets in logs or payloads
  - No hardcoded addresses
  - Minimal binary size
  - Backward compatibility maintained

### 4. Documentation

#### README.md
- [ ] Add feature to "Features" section
- [ ] Include CLI usage example (commands and flags)
- [ ] Describe what the feature does in 2-3 sentences
- [ ] Link to detailed docs if available

#### CHANGELOG.md
- [ ] Add entry under latest version or unreleased
- [ ] **Added**: List new files/commands
- [ ] **Modified**: List changed files with brief description
- [ ] **Fixed**: Bug fixes (if any)
- [ ] Mention backward compatibility status

#### Code Comments
- [ ] Add minimal comments for non-obvious logic
- [ ] Document why, not what (WHY field should be clear)
- [ ] Example: `// PE header validation prevents loading invalid binaries`

---

## Code Style Guidelines

### File Organization

```
beacon/
  main_<feature>.go         # New beacon variant
client/grumblecli/
  <feature>-generate.go     # New CLI command generator
server/listeners/
  http.go                   # Add handler function
server/
  <feature>db.go           # New database logic (if needed)
```

### Naming Conventions

- **Commands**: kebab-case (`generate stager`, `load module`)
- **Files**: kebab-case or underscore-separated (`stager-generate.go`, `main_stager_http.go`)
- **Functions**: CamelCase, descriptive (`DownloadAgent`, `RegisterBeacon`)
- **Variables**: camelCase for local, PascalCase for exported

### Beacon Binary Pattern

```go
//go:build <feature_tag>
// +build <feature_tag>

package main

import (
    "time"
    "orsted/beacon/core"
    "orsted/beacon/peers"
    "orsted/beacon/utils"
    "orsted/profiles"
)

var (
    // Injected at compile time via ldflags
    Config1 string
    Config2 string
)

func featureCore() {
    _ = profiles.InitialiseProfile()
    // Setup, registration, core loop
}

func main() {
    featureCore()
}
```

### Server Handler Pattern

```go
func NewFeatureHandler(w http.ResponseWriter, r *http.Request) {
    utils.PrintDebug("Handler called")
    
    // Extract and validate input
    param := r.URL.Query().Get("param")
    if param == "" {
        http.Error(w, "param required", http.StatusBadRequest)
        return
    }
    
    // Do work
    result, err := doWork(param)
    if err != nil {
        http.Error(w, "work failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Respond
    w.Header().Set("Content-Type", "application/json")
    w.Write(result)
}

// In addHttpHandler():
mux.HandleFunc("/feature/path/", NewFeatureHandler)
```

---

## Commit Message Format

Follow conventional commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `refactor`, `docs`, `test`  
**Scope**: `stager`, `beacon`, `server`, `cli`

**Examples:**
```
feat(stager): add two-stage stager delivery system

Add minimal stager binaries that fetch and execute full beacons on-demand.
Includes CLI generation, HTTP/HTTPS support, and server download endpoint.

Closes #<issue_number>
```

```
fix(beacon): correct arch flag handling for 32-bit builds
```

---

## Review Checklist

Before submitting a feature:

- [ ] Code compiles without errors
- [ ] All files are in the correct location
- [ ] README.md updated with usage examples
- [ ] CHANGELOG.md updated with changes
- [ ] Commit messages follow conventional format
- [ ] No hardcoded credentials or test data
- [ ] Binary size is reasonable (~50-500KB depending on feature)
- [ ] Error messages are clear and actionable
- [ ] Backward compatibility maintained (or clearly noted)

---

## Examples

### Example 1: Adding a New Beacon Variant

1. Create `beacon/main_feature.go` with feature-specific logic
2. Update `client/grumblecli/beacon-generate.go` to support `generate beacon feature`
3. Add flags for configuration (e.g., `--arch`, `--timeout`)
4. Update README.md with usage and description
5. Add entry to CHANGELOG.md under "Added"

### Example 2: Adding a Server Endpoint

1. Create handler function in `server/listeners/http.go`
2. Register in `addHttpHandler()`
3. Include authentication/validation (beacon ID, etc.)
4. Add to README.md under "Features"
5. Update CHANGELOG.md

---

## Questions?

Refer to existing implementations in the codebase:
- Beacon pattern: `beacon/main_http.go`, `beacon/main_stager_http.go`
- CLI pattern: `client/grumblecli/beacon-generate.go`, `client/grumblecli/stager-generate.go`
- Server pattern: `server/listeners/http.go` (`RegisterBeacon`, `DownloadAgent`)
