// Package vm provides virtual machine management functionality
package vm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// QMPGreeting is the initial message sent by QEMU on connect
type QMPGreeting struct {
	QMP struct {
		Version struct {
			QEMU struct {
				Major int `json:"major"`
				Minor int `json:"minor"`
				Micro int `json:"micro"`
			} `json:"qemu"`
			Package string `json:"package"`
		} `json:"version"`
		Capabilities []string `json:"capabilities"`
	} `json:"QMP"`
}

// QMPResponse is a response to a QMP command
type QMPResponse struct {
	Return json.RawMessage `json:"return"`
	Error  *QMPError       `json:"error"`
	ID     string          `json:"id"`
}

// QMPError represents a QMP error response
type QMPError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

// QMPClientInterface is the minimum interface the runner needs from a QMP
// client. The concrete QMPClient satisfies it. The interface exists so the
// runner can be tested with a mock QMP client (same pattern as HostDiscovery).
type QMPClientInterface interface {
	QueryStatus() (string, error)
	QueryCPUs() ([]VCPUInfo, error)
	QueryCPUsFast() ([]QMPVCPUInfo, error)
	QueryBlockStats() ([]QMPBlockDeviceStats, error)
	QueryNetdev() ([]QMPNetDeviceStats, error)
	QueryBalloon() (uint64, error)
	Close() error
	Quit() error
	Events() <-chan QMPEvent
}

// QMPEvent is an asynchronous event from QEMU
type QMPEvent struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Timestamp QMPTimestamp    `json:"timestamp"`
}

// QMPTimestamp represents a QMP event timestamp
type QMPTimestamp struct {
	Seconds      int64 `json:"seconds"`
	Microseconds int   `json:"microseconds"`
}

// QMPCommand is a command to send to QEMU
type QMPCommand struct {
	Execute string      `json:"execute"`
	Args    interface{} `json:"arguments,omitempty"`
	ID      string      `json:"id,omitempty"`
}

// QMPClient communicates with QEMU via the QMP protocol over a Unix socket.
// All operations are serialized through a mutex to prevent concurrent reads
// from the underlying bufio.Reader.
type QMPClient struct {
	conn       net.Conn
	socketPath string
	mu         sync.Mutex
	reader     *bufio.Reader
	nextID     int
	events     chan QMPEvent
}

// NewQMPClient creates a new QMP client and connects to the given Unix socket path
func NewQMPClient(socketPath string) (*QMPClient, error) {
	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to QMP socket %s: %w", socketPath, err)
	}

	client := &QMPClient{
		conn:       conn,
		socketPath: socketPath,
		reader:     bufio.NewReader(conn),
		nextID:     1,
		events:     make(chan QMPEvent, 64),
	}

	return client, nil
}

// Negotiate reads the QMP greeting and sends qmp_capabilities to complete negotiation.
// This must be called before Execute.
func (c *QMPClient) Negotiate() (*QMPGreeting, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set read deadline for greeting
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read greeting (first line from socket)
	greetingLine, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read QMP greeting: %w", err)
	}

	var greeting QMPGreeting
	if err := json.Unmarshal([]byte(greetingLine), &greeting); err != nil {
		return nil, fmt.Errorf("failed to parse QMP greeting: %w (raw: %s)", err, greetingLine)
	}

	// Send qmp_capabilities to complete negotiation
	cmd := QMPCommand{Execute: "qmp_capabilities"}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal qmp_capabilities: %w", err)
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}
	if _, err := c.conn.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send qmp_capabilities: %w", err)
	}

	// Read qmp_capabilities response
	if err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read qmp_capabilities response: %w", err)
	}

	var resp QMPResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse qmp_capabilities response: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("qmp_capabilities error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	return &greeting, nil
}

// Execute sends a QMP command and waits for the matching response.
// Each command gets a unique ID to disambiguate responses from async events.
// Events encountered while waiting for a response are queued to the Events channel.
func (c *QMPClient) Execute(command string, args interface{}) (*QMPResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := fmt.Sprintf("cmd-%d", c.nextID)
	c.nextID++

	cmd := QMPCommand{
		Execute: command,
		Args:    args,
		ID:      id,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QMP command: %w", err)
	}

	// Write command
	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}
	if _, err := c.conn.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write QMP command: %w", err)
	}

	// Read responses until we find the one matching our ID
	if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read QMP response for %s: %w", id, err)
		}

		// Try to parse as a command response
		var resp QMPResponse
		if err := json.Unmarshal([]byte(line), &resp); err == nil && resp.ID != "" {
			if resp.ID == id {
				return &resp, nil
			}
			// Different ID - unexpected, skip
			continue
		}

		// Try to parse as an event
		var event QMPEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil && event.Event != "" {
			select {
			case c.events <- event:
			default:
				// Channel full, drop event
			}
			continue
		}

		// Unknown message, skip
		continue
	}
}

// Events returns the channel of asynchronous QMP events.
// Events are captured during Execute calls that encounter async events
// while waiting for command responses.
func (c *QMPClient) Events() <-chan QMPEvent {
	return c.events
}

// Close closes the QMP connection
func (c *QMPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// QueryStatus returns the current VM status via query-status
func (c *QMPClient) QueryStatus() (string, error) {
	resp, err := c.Execute("query-status", nil)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("query-status error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp.Return, &result); err != nil {
		return "", fmt.Errorf("failed to parse query-status response: %w", err)
	}
	return result.Status, nil
}

// QueryCPUsFast returns vCPU information including thread IDs
func (c *QMPClient) QueryCPUsFast() ([]QMPVCPUInfo, error) {
	resp, err := c.Execute("query-cpus-fast", nil)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("query-cpus-fast error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var cpus []QMPVCPUInfo
	if err := json.Unmarshal(resp.Return, &cpus); err != nil {
		return nil, fmt.Errorf("failed to parse query-cpus-fast response: %w", err)
	}
	return cpus, nil
}

// QueryCPUs returns vCPU information via the QMP query-cpus command.
// This is distinct from QueryCPUsFast: it uses the older query-cpus command
// which returns thread_id (snake_case) and cpu state fields.
// The returned VCPUInfo.ThreadID is used by the runner to read per-thread
// CPU time from /proc/<pid>/task/<tid>/stat for delta-based CPU% computation.
func (c *QMPClient) QueryCPUs() ([]VCPUInfo, error) {
	resp, err := c.Execute("query-cpus", nil)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("query-cpus error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var cpus []VCPUInfo
	if err := json.Unmarshal(resp.Return, &cpus); err != nil {
		return nil, fmt.Errorf("failed to parse query-cpus response: %w", err)
	}
	return cpus, nil
}

// QMPVCPUInfo represents vCPU information from query-cpus-fast
type QMPVCPUInfo struct {
	CPU      int    `json:"CPU"`
	ThreadID int    `json:"thread-id"`
	QOMPath  string `json:"qom-path"`
	Props    struct {
		CoreID   int `json:"core-id"`
		DieID    int `json:"die-id"`
		ThreadID int `json:"thread-id"`
	} `json:"props"`
}

// VCPUInfo represents vCPU information from query-cpus.
// This is the typed return from QMPClient.QueryCPUs, distinct from
// QueryCPUsFast / QMPVCPUInfo. The QMP "query-cpus" wire format uses
// snake_case field names.
type VCPUInfo struct {
	CPU      int    `json:"CPU"`
	Current  bool   `json:"current"`
	Halted   bool   `json:"halted"`
	ThreadID int    `json:"thread_id"`
	QOMPath  string `json:"qom_path"`
}

// BalloonInfo is the typed response from QMP query-balloon.
// "actual" is the current balloon size in bytes; "max" is the upper
// bound the guest is allowed to inflate to.
type BalloonInfo struct {
	Actual    uint64 `json:"actual"`
	MemPeriod int64  `json:"mem_period"`
	Max       uint64 `json:"max"`
}

// QMPBlockDeviceStats is one element of the query-blockstats return
// array. The QMP wire format nests the counters under "stats".
type QMPBlockDeviceStats struct {
	Device   string           `json:"device"`
	NodeName string           `json:"node-name"`
	Stats    QMPBlockDeviceIO `json:"stats"`
}

// QMPBlockDeviceIO is the per-device I/O counter object inside
// query-blockstats. All counters are cumulative since VM start; the
// runner computes B/s and IOPS from deltas across successive snapshots.
type QMPBlockDeviceIO struct {
	RDBytes uint64 `json:"rd_bytes"`
	WRBytes uint64 `json:"wr_bytes"`
	RDOps   uint64 `json:"rd_operations"`
	WROps   uint64 `json:"wr_operations"`
}

// QMPNetDeviceIO is the per-device network counter object inside
// query-netdev. All counters are cumulative since VM start; the
// runner computes B/s and Pps from deltas across successive snapshots.
type QMPNetDeviceIO struct {
	RXBytes   uint64 `json:"rx-bytes"`
	TXBytes   uint64 `json:"tx-bytes"`
	RXPackets uint64 `json:"rx-packets"`
	TXPackets uint64 `json:"tx-packets"`
}

// QMPNetDeviceStats is one element of the query-netdev return
// array. The QMP wire format nests the counters under "stats".
type QMPNetDeviceStats struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Stats QMPNetDeviceIO `json:"stats"`
}

// QueryBalloon returns the current balloon size in bytes via query-balloon.
// The QMP response shape is {"return": {"actual": N, "mem_period": ..., "max": N}}.
//
// Graceful degradation: when the guest has no balloon driver, QEMU replies with
// a "GenericError: Balloon is not activated" error. We treat that specific case
// as a successful no-op (returns 0 with no error) because the absence of a
// balloon is a normal guest configuration, not a failure.
func (c *QMPClient) QueryBalloon() (uint64, error) {
	resp, err := c.Execute("query-balloon", nil)
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		// Graceful degradation: "Balloon is not activated" is a normal
		// configuration, not a failure. Return 0 with no error.
		if resp.Error.Class == "GenericError" &&
			strings.Contains(resp.Error.Desc, "not activated") {
			return 0, nil
		}
		return 0, fmt.Errorf("query-balloon error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var info BalloonInfo
	if err := json.Unmarshal(resp.Return, &info); err != nil {
		return 0, fmt.Errorf("failed to parse query-balloon response: %w", err)
	}
	return info.Actual, nil
}

// QueryBlockStats returns per-block-device I/O counters via query-blockstats.
// Each element carries r/w bytes and r/w operation counts since VM start;
// the runner computes B/s and IOPS from deltas across successive snapshots.
//
// The QMP response shape is {"return": [{"device": "...", "stats": {...}}, ...]}.
func (c *QMPClient) QueryBlockStats() ([]QMPBlockDeviceStats, error) {
	resp, err := c.Execute("query-blockstats", nil)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("query-blockstats error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var stats []QMPBlockDeviceStats
	if err := json.Unmarshal(resp.Return, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse query-blockstats response: %w", err)
	}
	return stats, nil
}

// QueryNetdev returns per-network-device I/O counters via query-netdev.
// Each element carries r/w bytes and r/w packet counts since VM start;
// the runner computes B/s and Pps from deltas across successive snapshots.
func (c *QMPClient) QueryNetdev() ([]QMPNetDeviceStats, error) {
	resp, err := c.Execute("query-netdev", nil)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("query-netdev error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var stats []QMPNetDeviceStats
	if err := json.Unmarshal(resp.Return, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse query-netdev response: %w", err)
	}
	return stats, nil
}

// Cont resumes a paused VM
func (c *QMPClient) Cont() error {
	resp, err := c.Execute("cont", nil)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("cont error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}
	return nil
}

// Stop pauses a VM
func (c *QMPClient) Stop() error {
	resp, err := c.Execute("stop", nil)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("stop error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}
	return nil
}

// Quit shuts down QEMU
func (c *QMPClient) Quit() error {
	resp, err := c.Execute("quit", nil)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("quit error: %s: %s", resp.Error.Class, resp.Error.Desc)
	}
	return nil
}
