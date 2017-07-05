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

//StartCPUProfile 生成CPU性能监控文件
func StartCPUProfile() error {
	if atomic.CompareAndSwapInt32(&cpuProfiling, 0, 1) {
		cpuProfile, err := os.Create("./cpu.pprof")
		if err != nil {
			return err
		}
		return pprof.StartCPUProfile(cpuProfile)
	}
	return nil
}

//StopCPUProfile 停止生成CPU监控文件
func StopCPUProfile() {
	if atomic.LoadInt32(&cpuProfiling) == 1 {
		pprof.StopCPUProfile()
		cpuProfile.Close()
		cpuProfile = nil
	}
}
