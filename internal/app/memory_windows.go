//go:build windows

package app

import (
	"syscall"
	"unsafe"
)

var (
	psapi             = syscall.NewLazyDLL("psapi.dll")
	getProcessMemory  = psapi.NewProc("GetProcessMemoryInfo")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	getCurrentProcess = kernel32.NewProc("GetCurrentProcess")
)

type processMemoryCounters struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// processMemory reports the working set of this process in bytes. The core
// runs in-process, so this covers app and core together; the split the memory
// card shows comes from the Go runtime's own accounting.
func processMemory() uint64 {
	handle, _, _ := getCurrentProcess.Call()
	var counters processMemoryCounters
	counters.cb = uint32(unsafe.Sizeof(counters))
	ret, _, _ := getProcessMemory.Call(
		handle,
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.cb),
	)
	if ret == 0 {
		return 0
	}
	return uint64(counters.WorkingSetSize)
}
