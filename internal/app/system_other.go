//go:build !windows

package app

import "errors"

func setSystemProxy(enable bool, addr string) error {
	return errors.New("当前平台暂不支持系统代理")
}

func isElevated() bool {
	return false
}

func relaunchElevated() error {
	return errors.New("当前平台暂不支持")
}

func tunSupported() bool {
	return false
}
