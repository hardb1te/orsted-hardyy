package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	mprDLL                 = syscall.NewLazyDLL("mpr.dll")
	procWNetAddConnection2 = mprDLL.NewProc("WNetAddConnection2W")
	procWNetCancelConn2    = mprDLL.NewProc("WNetCancelConnection2W")
)

type NETRESOURCEW struct {
	DwScope       uint32
	DwType        uint32
	DwDisplayType uint32
	DwUsage       uint32
	LpLocalName   *uint16
	LpRemoteName  *uint16
	LpComment     *uint16
	LpProvider    *uint16
}

func AuthenticateRemote(hostname, username, password, domain string) error {
	var fullUsername string
	if domain != "" && domain != "_" {
		fullUsername = domain + `\` + username
	} else {
		fullUsername = username
	}

	remoteName := `\\` + hostname + `\IPC$`

	remoteNamePtr, err := syscall.UTF16PtrFromString(remoteName)
	if err != nil {
		return fmt.Errorf("failed to convert remote name: %w", err)
	}
	usernamePtr, err := syscall.UTF16PtrFromString(fullUsername)
	if err != nil {
		return fmt.Errorf("failed to convert username: %w", err)
	}
	passwordPtr, err := syscall.UTF16PtrFromString(password)
	if err != nil {
		return fmt.Errorf("failed to convert password: %w", err)
	}

	cancelPtr, _ := syscall.UTF16PtrFromString(remoteName)
	procWNetCancelConn2.Call(
		uintptr(unsafe.Pointer(cancelPtr)),
		0,
		1,
	)

	nr := NETRESOURCEW{
		DwType:       0,
		LpRemoteName: remoteNamePtr,
	}

	ret, _, _ := procWNetAddConnection2.Call(
		uintptr(unsafe.Pointer(&nr)),
		uintptr(unsafe.Pointer(passwordPtr)),
		uintptr(unsafe.Pointer(usernamePtr)),
		0,
	)

	if ret != 0 {
		return fmt.Errorf("WNetAddConnection2W failed (error %d)", ret)
	}

	return nil
}

func DisconnectRemote(hostname string) error {
	remoteName := `\\` + hostname + `\IPC$`
	remoteNamePtr, err := syscall.UTF16PtrFromString(remoteName)
	if err != nil {
		return fmt.Errorf("failed to convert remote name: %w", err)
	}

	ret, _, _ := procWNetCancelConn2.Call(
		uintptr(unsafe.Pointer(remoteNamePtr)),
		0,
		1,
	)

	if ret != 0 {
		return fmt.Errorf("WNetCancelConnection2W failed (error %d)", ret)
	}

	return nil
}

func PsExec(hostname string, binPath string, fileData []byte, arguments string, serviceName string, serviceDesc string, username string, password string, domain string) error {
	if username != "" && username != "_" {
		err := AuthenticateRemote(hostname, username, password, domain)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		defer DisconnectRemote(hostname)
	}

	binPath = strings.Split(binPath, ".")[0] + generateVersionID() + ".exe"
	serviceName = serviceName + generateVersionID()
	err := UploadFile(hostname, binPath, fileData)
	if err != nil {
		return err
	}

	err = StartService(hostname, binPath, arguments, serviceName, serviceDesc)
	return err
}

func generateVersionID() string {
	n := rand.Intn(90000) + 10000
	return fmt.Sprintf("v%d", n)
}

func UploadFile(hostname string, binPath string, fileData []byte) error {
	uncPath := fmt.Sprintf(`\\%s\%s`, hostname, binPath)
	uncPath = strings.Replace(uncPath, ":", "$", 1)

	dir := filepath.Dir(uncPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	f, err := os.Create(uncPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func StartService(hostname string, binPath string, arguments string, serviceName string, serviceDesc string) error {
	manager, err := mgr.ConnectRemote(hostname)
	if err != nil {
		return err
	}

	service, err := manager.CreateService(serviceName, binPath, mgr.Config{
		ErrorControl:   mgr.ErrorNormal,
		BinaryPathName: binPath,
		Description:    serviceDesc,
		DisplayName:    serviceName,
		ServiceType:    windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:      mgr.StartManual,
	}, arguments)

	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		return err
	}
	return err
}
