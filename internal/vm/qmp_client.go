// Package vm provides virtual machine management functionality
package vm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
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
