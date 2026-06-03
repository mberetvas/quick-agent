//go:build windows

package daemon

import "syscall"

func processAlive(pid int) bool {
	const query = syscall.STANDARD_RIGHTS_READ | 0x1000 // PROCESS_QUERY_LIMITED_INFORMATION
	handle, err := syscall.OpenProcess(query, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}
	const stillActive = 259
	return code == stillActive
}
