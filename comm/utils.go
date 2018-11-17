// +build !windows

package comm

import (
	"syscall"
)

func GetRlimitFile() uint64 {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(err)
	}
	return rLimit.Cur
}
