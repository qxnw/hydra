package pprof

import (
	"errors"
	"os"
	"runtime/pprof"
	"sync/atomic"
)

var (
	cpuProfiling          int32
	cpuProfile            *os.File
	ErrCPUProfileStart    = errors.New("CPU profile already start")
	ErrCPUProfileNotStart = errors.New("CPU profile not start")
)

func StartCPUProfile(file string) error {
	if atomic.CompareAndSwapInt32(&cpuProfiling, 0, 1) {
		cpuProfile, err := os.Create(file)
		if err != nil {
			return err
		}
		return pprof.StartCPUProfile(cpuProfile)
	}
	return nil
}

func StopCPUProfile() {
	if atomic.LoadInt32(&cpuProfiling) == 1 {
		pprof.StopCPUProfile()
		cpuProfile.Close()
		cpuProfile = nil
	}
}
