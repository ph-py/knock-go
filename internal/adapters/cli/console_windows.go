//go:build windows

package cli

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func init() {
	enableVirtualTerminal()
}

func enableVirtualTerminal() {
	stdout := os.Stdout.Fd()
	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(stdout, uintptr(unsafe.Pointer(&mode)))
	if r1 != 0 {
		mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
		procSetConsoleMode.Call(stdout, uintptr(mode))
	}
}
