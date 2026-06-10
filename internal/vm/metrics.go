// Package vm provides virtual machine management functionality
package vm

import (
	"time"
)

// Metrics is a point-in-time value object containing both QMP-derived guest
// metrics and host /proc-derived metrics for the QEMU process. It is the
// boundary object the view consumes. The runner produces it via Snapshot().
//
// VCPUs is implemented in S3. BlockDevices, BalloonBytes, HostRSSBytes, and
// HostCPUJiffies are placeholders for S4 and S5 and are zero-valued in S3.
type Metrics struct {
	Timestamp time.Time
	Status    string // from query-status

	// Guest / QMP
	VCPUs        []VCPUStat   // per-vCPU thread id + total CPU time in ns
	BlockDevices []BlockStat  // per-block r/w bytes, r/w operations (raw counters)
	BalloonBytes uint64       // from query-balloon; 0 if not available

	// Host /proc/<qemu-pid>
	HostRSSBytes     uint64 // VmRSS from /proc/<pid>/status
	HostCPUJiffies   uint64 // utime+stime from /proc/<pid>/stat
}

// VCPUStat holds per-vCPU statistics for a single snapshot.
type VCPUStat struct {
	ThreadID  int   // host thread ID of the vCPU
	CPUTimeNs int64 // total CPU time consumed in nanoseconds (from /proc)
}

// BlockStat holds per-block-device raw counters for a single snapshot.
type BlockStat struct {
	Device  string // e.g. "drive0"
	RDBytes uint64
	WRBytes uint64
	RDOps   uint64
	WROps   uint64
}
