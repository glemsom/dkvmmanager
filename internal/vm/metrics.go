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
	VCPUs        []VCPUStat  // per-vCPU thread id + total CPU time in ns
	BlockDevices []BlockStat // per-block r/w bytes, r/w operations (raw counters)
	NetDevices   []NetStat   // per-netdev r/w bytes, r/w packets (derived rates)
	BalloonBytes uint64      // from query-balloon; 0 if not available

	// Host /proc/<qemu-pid>
	HostRSSBytes   uint64 // VmRSS from /proc/<pid>/status
	HostCPUJiffies uint64 // utime+stime from /proc/<pid>/stat
}

// VCPUStat holds per-vCPU statistics for a single snapshot.
type VCPUStat struct {
	ThreadID  int   // host thread ID of the vCPU
	CPUTimeNs int64 // total CPU time consumed in nanoseconds (from /proc)
}

// BlockStat holds per-block-device statistics for a single snapshot.
// Raw counters (RDBytes/WRBytes/RDOps/WROps) come from QMP query-blockstats.
// Derived rates (RDBps/WRBps/RDIOPS/WRIOPS) are computed from deltas against
// the previous snapshot by the runner; the view consumes the already-derived
// numbers directly. On a cold snapshot the rate fields are zero.
type BlockStat struct {
	Device  string // e.g. "drive0"
	RDBytes uint64 // cumulative r/w bytes since VM start
	WRBytes uint64
	RDOps   uint64 // cumulative r/w operation counts since VM start
	WROps   uint64

	// Derived (per-second, populated from delta math in Snapshot()).
	// Cold snapshot leaves these at zero. Negative or wrapped counters
	// (QEMU counter reset) also leave these at zero rather than going negative.
	RDBps  uint64 // bytes read per second
	WRBps  uint64 // bytes written per second
	RDIOPS uint64 // read operations per second
	WRIOPS uint64 // write operations per second
}

// NetStat holds per-network-device statistics for a single snapshot.
// Raw counters (RXBytes/TXBytes/RXPackets/TXPackets) come from QMP query-netdev.
// Derived rates (RXBps/TXBps/RXPps/TXPps) are computed from deltas against
// the previous snapshot by the runner; the view consumes the already-derived
// numbers directly. On a cold snapshot the rate fields are zero.
type NetStat struct {
	Device string // e.g. "hostnet0"
	RXBytes uint64 // cumulative RX bytes since VM start
	TXBytes uint64 // cumulative TX bytes since VM start
	RXPackets uint64 // cumulative RX packets since VM start
	TXPackets uint64 // cumulative TX packets since VM start

	// Derived (per-second, populated from delta math in Snapshot()).
	// Cold snapshot leaves these at zero. Negative or wrapped counters
	// (QEMU counter reset) also leave these at zero rather than going negative.
	RXBps  uint64 // bytes received per second
	TXBps  uint64 // bytes transmitted per second
	RXPps  uint64 // packets received per second
	TXPps  uint64 // packets transmitted per second
}
