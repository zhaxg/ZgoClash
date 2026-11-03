//go:build windows

package app

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

const proxyBypass = `<local>;localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*`

func setSystemProxy(enable bool, addr string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开系统代理设置失败: %w", err)
	}
	defer key.Close()

	if enable {
		if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
			return err
		}
		if err := key.SetStringValue("ProxyServer", addr); err != nil {
			return err
		}
		if err := key.SetStringValue("ProxyOverride", proxyBypass); err != nil {
			return err
		}
	} else {
		if err := key.SetDWordValue("ProxyEnable", 0); err != nil {
			return err
		}
	}
	refreshInternetSettings()
	return nil
}

func refreshInternetSettings() {
	wininet := syscall.NewLazyDLL("wininet.dll")
	setOption := wininet.NewProc("InternetSetOptionW")
	const (
		internetOptionRefresh         = 37
		internetOptionSettingsChanged = 39
	)
	_, _, _ = setOption.Call(0, uintptr(internetOptionSettingsChanged), 0, 0)
	_, _, _ = setOption.Call(0, uintptr(internetOptionRefresh), 0, 0)
}

func isElevated() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}

func relaunchElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	dir, _ := syscall.UTF16PtrFromString("")
	arg, _ := syscall.UTF16PtrFromString("")
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	const swShownormal = 1
	ret, _, callErr := proc.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(arg)),
		uintptr(unsafe.Pointer(dir)),
		swShownormal,
	)
	if ret <= 32 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return fmt.Errorf("请求管理员权限失败: %w", callErr)
		}
		return errors.New("请求管理员权限失败")
	}
	return nil
}

func tunSupported() bool {
	return isElevated()
}
