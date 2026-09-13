package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ShellcodeGenerator struct {
	peFilePath string
	donutPath  string
}

func NewShellcodeGenerator(peFile string) (*ShellcodeGenerator, error) {
	if _, err := os.Stat(peFile); err != nil {
		return nil, fmt.Errorf("PE file not found: %s", peFile)
	}

	donutPath := "tools/donut"
	if _, err := os.Stat(donutPath); err != nil {
		return nil, fmt.Errorf("donut binary not found at %s", donutPath)
	}

	return &ShellcodeGenerator{
		peFilePath: peFile,
		donutPath:  donutPath,
	}, nil
}

func (sg *ShellcodeGenerator) GenerateShellcode() ([]byte, error) {
	tmpBin := filepath.Join(os.TempDir(), "shellcode.bin")
	defer os.Remove(tmpBin)

	cmd := exec.Command(sg.donutPath, "-i", sg.peFilePath, "-o", tmpBin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("donut conversion failed: %w", err)
	}

	shellcode, err := os.ReadFile(tmpBin)
	if err != nil {
		return nil, fmt.Errorf("failed to read shellcode output: %w", err)
	}

	return shellcode, nil
}

func (sg *ShellcodeGenerator) FormatBase64(shellcode []byte) string {
	return base64.StdEncoding.EncodeToString(shellcode)
}

func (sg *ShellcodeGenerator) FormatHex(shellcode []byte) string {
	return hex.EncodeToString(shellcode)
}

func (sg *ShellcodeGenerator) FormatRaw(shellcode []byte) []byte {
	return shellcode
}

func (sg *ShellcodeGenerator) FormatVBA(shellcode []byte) string {
	var output strings.Builder
	output.WriteString("Dim shellcode() As Byte\n")
	output.WriteString("shellcode = Array(\n")

	for i, b := range shellcode {
		if i > 0 && i%16 == 0 {
			output.WriteString("\n")
		}
		if i > 0 && i%16 != 0 {
			output.WriteString(", ")
		}
		fmt.Fprintf(&output, "&H%02X", b)
	}

	output.WriteString("\n)\n")
	output.WriteString("' Use VBA shellcode runner to execute shellcode array\n")

	return output.String()
}

func (sg *ShellcodeGenerator) FormatCSharp(shellcode []byte) string {
	var output strings.Builder
	output.WriteString("byte[] shellcode = new byte[] {\n")

	for i, b := range shellcode {
		if i > 0 && i%16 == 0 {
			output.WriteString("\n")
		}
		if i > 0 && i%16 != 0 {
			output.WriteString(", ")
		}
		fmt.Fprintf(&output, "0x%02X", b)
	}

	output.WriteString("\n};\n")
	output.WriteString("// Use C# shellcode runner to execute shellcode array\n")

	return output.String()
}

func (sg *ShellcodeGenerator) FormatJavaScript(shellcode []byte) string {
	var output strings.Builder
	output.WriteString("var shellcode = [\n")

	for i, b := range shellcode {
		if i > 0 && i%16 == 0 {
			output.WriteString("\n")
		}
		if i > 0 && i%16 != 0 {
			output.WriteString(", ")
		}
		fmt.Fprintf(&output, "0x%02X", b)
	}

	output.WriteString("\n];\n")
	output.WriteString("// Use JavaScript shellcode runner to execute shellcode array\n")

	return output.String()
}

func (sg *ShellcodeGenerator) GenerateAndFormat(format string) (interface{}, error) {
	shellcode, err := sg.GenerateShellcode()
	if err != nil {
		return nil, err
	}

	switch format {
	case "base64":
		return sg.FormatBase64(shellcode), nil
	case "hex":
		return sg.FormatHex(shellcode), nil
	case "raw":
		return sg.FormatRaw(shellcode), nil
	case "vba":
		return sg.FormatVBA(shellcode), nil
	case "csharp":
		return sg.FormatCSharp(shellcode), nil
	case "javascript":
		return sg.FormatJavaScript(shellcode), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}
