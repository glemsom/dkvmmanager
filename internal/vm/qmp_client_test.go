package vm

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQMPGreetingParsing(t *testing.T) {
	greetingJSON := `{"QMP": {"version": {"qemu": {"major": 8, "minor": 2, "micro": 0}, "package": ""}, "capabilities": []}}`

	var greeting QMPGreeting
	if err := json.Unmarshal([]byte(greetingJSON), &greeting); err != nil {
		t.Fatalf("Failed to parse greeting: %v", err)
	}

	if greeting.QMP.Version.QEMU.Major != 8 {
		t.Errorf("Expected major=8, got %d", greeting.QMP.Version.QEMU.Major)
	}
	if greeting.QMP.Version.QEMU.Minor != 2 {
		t.Errorf("Expected minor=2, got %d", greeting.QMP.Version.QEMU.Minor)
	}
}

func TestQMPCommandMarshaling(t *testing.T) {
	cmd := QMPCommand{
		Execute: "query-status",
		ID:      "cmd-1",
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("Failed to marshal command: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed["execute"] != "query-status" {
		t.Errorf("Expected execute=query-status, got %v", parsed["execute"])
	}
	if parsed["id"] != "cmd-1" {
		t.Errorf("Expected id=cmd-1, got %v", parsed["id"])
	}
}

func TestQMPResponseParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantID   string
		wantErrC string
	}{
		{
			name:    "success response",
			input:   `{"return": {}, "id": "cmd-1"}`,
			wantID:  "cmd-1",
			wantErr: false,
		},
		{
			name:     "error response",
			input:    `{"error": {"class": "CommandNotFound", "desc": "Unknown command"}, "id": "cmd-2"}`,
			wantID:   "cmd-2",
			wantErr:  true,
			wantErrC: "CommandNotFound",
		},
		{
			name:   "status response",
			input:  `{"return": {"status": "running", "singlestep": false, "running": true}, "id": "cmd-3"}`,
			wantID: "cmd-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp QMPResponse
			if err := json.Unmarshal([]byte(tt.input), &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if resp.ID != tt.wantID {
				t.Errorf("Expected ID=%s, got %s", tt.wantID, resp.ID)
			}

			if tt.wantErr && resp.Error == nil {
				t.Error("Expected error in response, got nil")
			}
			if tt.wantErr && resp.Error != nil && resp.Error.Class != tt.wantErrC {
				t.Errorf("Expected error class=%s, got %s", tt.wantErrC, resp.Error.Class)
			}
		})
	}
}

func TestQMPEventParsing(t *testing.T) {
	eventJSON := `{"event": "SHUTDOWN", "data": {"guest": false, "reason": "host-qmp-quit"}, "timestamp": {"seconds": 1711000000, "microseconds": 123456}}`

	var event QMPEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	if event.Event != "SHUTDOWN" {
		t.Errorf("Expected event=SHUTDOWN, got %s", event.Event)
	}
	if event.Timestamp.Seconds != 1711000000 {
		t.Errorf("Expected seconds=1711000000, got %d", event.Timestamp.Seconds)
	}
}

func TestQMPVCPUInfoParsing(t *testing.T) {
	respJSON := `{"return": [{"CPU": 0, "current": true, "halted": false, "qom-path": "/machine/unattached/device[0]", "thread-id": 12345, "props": {"core-id": 0, "die-id": 0, "socket-id": 0, "thread-id": 0}}], "id": "cmd-1"}`

	var resp QMPResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	var cpus []QMPVCPUInfo
	if err := json.Unmarshal(resp.Return, &cpus); err != nil {
		t.Fatalf("Failed to parse vCPU info: %v", err)
	}

	if len(cpus) != 1 {
		t.Fatalf("Expected 1 CPU, got %d", len(cpus))
	}

	if cpus[0].ThreadID != 12345 {
		t.Errorf("Expected thread-id=12345, got %d", cpus[0].ThreadID)
	}
	if cpus[0].Props.CoreID != 0 {
		t.Errorf("Expected core-id=0, got %d", cpus[0].Props.CoreID)
	}
}

// mockQMPServer creates a Unix socket server that simulates QMP protocol
func mockQMPServer(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()

	dir := t.TempDir()
	socketPath = filepath.Join(dir, "test-qmp.sock")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock QMP server: %v", err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)

		// Send greeting
		greeting := `{"QMP": {"version": {"qemu": {"major": 8, "minor": 2, "micro": 0}, "package": ""}, "capabilities": []}}` + "\n"
		conn.Write([]byte(greeting))

		// Read commands and respond
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)

			var cmd map[string]interface{}
			if err := json.Unmarshal([]byte(line), &cmd); err != nil {
				continue
			}

			execute, _ := cmd["execute"].(string)
			id, _ := cmd["id"].(string)

			var respData map[string]interface{}
			switch execute {
			case "qmp_capabilities":
				respData = map[string]interface{}{"return": map[string]interface{}{}}
			case "query-status":
				respData = map[string]interface{}{
					"return": map[string]interface{}{
						"status":     "running",
						"singlestep": false,
						"running":    true,
					},
				}
			case "query-cpus-fast":
				respData = map[string]interface{}{
					"return": []interface{}{
						map[string]interface{}{
							"CPU":       0,
							"current":   true,
							"halted":    false,
							"thread-id": 12345,
							"qom-path":  "/machine/...",
						},
					},
				}
			case "query-cpus":
				respData = map[string]interface{}{
					"return": []interface{}{
						map[string]interface{}{
							"CPU":       0,
							"current":   true,
							"halted":    false,
							"thread_id": 12345,
							"qom_path":  "/machine/unattached/device[0]",
						},
						map[string]interface{}{
							"CPU":       1,
							"current":   true,
							"halted":    false,
							"thread_id": 12346,
							"qom_path":  "/machine/unattached/device[1]",
						},
					},
				}
			case "query-balloon":
				respData = map[string]interface{}{
					"return": map[string]interface{}{
						"actual":     2147483648,
						"mem_period": 0,
						"max":        2147483648,
					},
				}
			case "query-blockstats":
				respData = map[string]interface{}{
					"return": []interface{}{
						map[string]interface{}{
							"device":    "drive0",
							"node-name": "#block0",
							"stats": map[string]interface{}{
								"rd_bytes":      4096,
								"wr_bytes":      8192,
								"rd_operations": 40,
								"wr_operations": 80,
							},
							"parent": nil,
						},
						map[string]interface{}{
							"device":    "drive1",
							"node-name": "#block1",
							"stats": map[string]interface{}{
								"rd_bytes":      0,
								"wr_bytes":      0,
								"rd_operations": 0,
								"wr_operations": 0,
							},
							"parent": nil,
						},
					},
				}
			case "quit":
				respData = map[string]interface{}{"return": map[string]interface{}{}}
			default:
				respData = map[string]interface{}{
					"error": map[string]interface{}{
						"class": "CommandNotFound",
						"desc":  "Unknown command: " + execute,
					},
				}
			}

			if id != "" {
				respData["id"] = id
			}

			respBytes, _ := json.Marshal(respData)
			conn.Write(append(respBytes, '\n'))
		}
	}()

	cleanup = func() {
		listener.Close()
		os.RemoveAll(dir)
	}

	return socketPath, cleanup
}

func TestQMPClientNegotiate(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	greeting, err := client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	if greeting.QMP.Version.QEMU.Major != 8 {
		t.Errorf("Expected QEMU major=8, got %d", greeting.QMP.Version.QEMU.Major)
	}
}

func TestQMPClientExecute(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	// Test query-status
	resp, err := client.Execute("query-status", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("Expected no error, got %s: %s", resp.Error.Class, resp.Error.Desc)
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp.Return, &status); err != nil {
		t.Fatalf("Failed to parse status: %v", err)
	}
	if status.Status != "running" {
		t.Errorf("Expected status=running, got %s", status.Status)
	}
}

func TestQMPClientQueryStatus(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	status, err := client.QueryStatus()
	if err != nil {
		t.Fatalf("QueryStatus failed: %v", err)
	}
	if status != "running" {
		t.Errorf("Expected status=running, got %s", status)
	}
}

func TestQMPClientConnectFailure(t *testing.T) {
	_, err := NewQMPClient("/tmp/nonexistent-qmp-test-socket.sock")
	if err == nil {
		t.Error("Expected connection failure for nonexistent socket")
	}
}

func TestQMPStatusResponseParsing(t *testing.T) {
	resp := QMPResponse{
		Return: json.RawMessage(`{"status": "running", "singlestep": false, "running": true}`),
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp.Return, &status); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if status.Status != "running" {
		t.Errorf("Expected running, got %s", status.Status)
	}
}

// TestQMPClientQueryBalloon tests QueryBalloon against the fake Unix-socket server.
// The fake server returns actual=2147483648; the test asserts the typed
// return value matches.
func TestQMPClientQueryBalloon(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	actual, err := client.QueryBalloon()
	if err != nil {
		t.Fatalf("QueryBalloon failed: %v", err)
	}
	if actual != 2147483648 {
		t.Errorf("Expected actual=2147483648, got %d", actual)
	}
}

// TestQMPClientQueryBalloonGracefulDegradation tests that when the guest has
// no balloon driver, the server returns a "GenericError: Balloon is not
// activated" error and the client returns (0, nil) — no error surfaced.
func TestQMPClientQueryBalloonGracefulDegradation(t *testing.T) {
	// Custom fake server that returns the "not activated" error.
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test-qmp.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock QMP server: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)

		// Greeting
		greeting := `{"QMP": {"version": {"qemu": {"major": 8, "minor": 2, "micro": 0}, "package": ""}, "capabilities": []}}` + "\n"
		conn.Write([]byte(greeting))

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)

			var cmd map[string]interface{}
			if err := json.Unmarshal([]byte(line), &cmd); err != nil {
				continue
			}
			execute, _ := cmd["execute"].(string)
			id, _ := cmd["id"].(string)

			var respData map[string]interface{}
			switch execute {
			case "qmp_capabilities":
				respData = map[string]interface{}{"return": map[string]interface{}{}}
			case "query-balloon":
				// Simulate no-balloon-driver error.
				respData = map[string]interface{}{
					"error": map[string]interface{}{
						"class": "GenericError",
						"desc":  "Balloon is not activated",
					},
				}
			default:
				respData = map[string]interface{}{"return": map[string]interface{}{}}
			}
			if id != "" {
				respData["id"] = id
			}
			respBytes, _ := json.Marshal(respData)
			conn.Write(append(respBytes, '\n'))
		}
	}()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	actual, err := client.QueryBalloon()
	if err != nil {
		t.Fatalf("QueryBalloon should NOT surface 'not activated' as error: %v", err)
	}
	if actual != 0 {
		t.Errorf("Expected actual=0 when no balloon driver, got %d", actual)
	}
}

// TestQMPClientQueryBlockStats tests QueryBlockStats against the fake
// Unix-socket server. Asserts that r/w bytes and op counts are typed correctly.
func TestQMPClientQueryBlockStats(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	stats, err := client.QueryBlockStats()
	if err != nil {
		t.Fatalf("QueryBlockStats failed: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(stats))
	}

	if stats[0].Device != "drive0" {
		t.Errorf("Expected device[0]='drive0', got '%s'", stats[0].Device)
	}
	if stats[0].Stats.RDBytes != 4096 {
		t.Errorf("Expected rd_bytes=4096, got %d", stats[0].Stats.RDBytes)
	}
	if stats[0].Stats.WRBytes != 8192 {
		t.Errorf("Expected wr_bytes=8192, got %d", stats[0].Stats.WRBytes)
	}
	if stats[0].Stats.RDOps != 40 {
		t.Errorf("Expected rd_operations=40, got %d", stats[0].Stats.RDOps)
	}
	if stats[0].Stats.WROps != 80 {
		t.Errorf("Expected wr_operations=80, got %d", stats[0].Stats.WROps)
	}
	if stats[1].Device != "drive1" {
		t.Errorf("Expected device[1]='drive1', got '%s'", stats[1].Device)
	}
}

// TestQMPClientQueryCPUs tests the QueryCPUs method (distinct from QueryCPUsFast)
// against the fake Unix-socket JSON server.
func TestQMPClientQueryCPUs(t *testing.T) {
	socketPath, cleanup := mockQMPServer(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	client, err := NewQMPClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Negotiate()
	if err != nil {
		t.Fatalf("Negotiation failed: %v", err)
	}

	cpus, err := client.QueryCPUs()
	if err != nil {
		t.Fatalf("QueryCPUs failed: %v", err)
	}

	if len(cpus) != 2 {
		t.Fatalf("Expected 2 CPUs, got %d", len(cpus))
	}

	if cpus[0].ThreadID != 12345 {
		t.Errorf("Expected thread_id=12345, got %d", cpus[0].ThreadID)
	}
	if cpus[1].ThreadID != 12346 {
		t.Errorf("Expected thread_id=12346, got %d", cpus[1].ThreadID)
	}
	if cpus[0].CPU != 0 {
		t.Errorf("Expected CPU index 0, got %d", cpus[0].CPU)
	}
}

// TestQMPBalloonResponseParsing tests parsing the QMP query-balloon response format.
// The wire format is {"return": {"actual": <bytes>, "mem_period": <seconds>, "max": <bytes>}}.
func TestQMPBalloonResponseParsing(t *testing.T) {
	respJSON := `{"return": {"actual": 2147483648, "mem_period": 0, "max": 2147483648}, "id": "cmd-1"}`

	var resp QMPResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	var balloon BalloonInfo
	if err := json.Unmarshal(resp.Return, &balloon); err != nil {
		t.Fatalf("Failed to parse balloon info: %v", err)
	}

	if balloon.Actual != 2147483648 {
		t.Errorf("Expected actual=2147483648, got %d", balloon.Actual)
	}
	if balloon.Max != 2147483648 {
		t.Errorf("Expected max=2147483648, got %d", balloon.Max)
	}
}

// TestQMPBlockStatsResponseParsing tests parsing the QMP query-blockstats
// response format. The wire format is an array of devices, each with a
// "device" name and a "stats" sub-object containing rd_bytes/wr_bytes/rd_operations/wr_operations.
func TestQMPBlockStatsResponseParsing(t *testing.T) {
	respJSON := `{"return": [
		{
			"device": "drive0",
			"node-name": "#block0",
			"stats": {
				"rd_bytes": 1024,
				"wr_bytes": 2048,
				"rd_operations": 10,
				"wr_operations": 20
			},
			"parent": null
		},
		{
			"device": "drive1",
			"node-name": "#block1",
			"stats": {
				"rd_bytes": 0,
				"wr_bytes": 0,
				"rd_operations": 0,
				"wr_operations": 0
			},
			"parent": null
		}
	], "id": "cmd-1"}`

	var resp QMPResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	var stats []QMPBlockDeviceStats
	if err := json.Unmarshal(resp.Return, &stats); err != nil {
		t.Fatalf("Failed to parse block stats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(stats))
	}

	if stats[0].Device != "drive0" {
		t.Errorf("Expected device[0]='drive0', got '%s'", stats[0].Device)
	}
	if stats[0].Stats.RDBytes != 1024 {
		t.Errorf("Expected rd_bytes=1024, got %d", stats[0].Stats.RDBytes)
	}
	if stats[0].Stats.WRBytes != 2048 {
		t.Errorf("Expected wr_bytes=2048, got %d", stats[0].Stats.WRBytes)
	}
	if stats[0].Stats.RDOps != 10 {
		t.Errorf("Expected rd_operations=10, got %d", stats[0].Stats.RDOps)
	}
	if stats[0].Stats.WROps != 20 {
		t.Errorf("Expected wr_operations=20, got %d", stats[0].Stats.WROps)
	}
}

// TestQMPQueryCPUsInfoParsing tests parsing the QMP query-cpus response format
func TestQMPQueryCPUsInfoParsing(t *testing.T) {
	respJSON := `{"return": [{"CPU": 0, "current": true, "halted": false, "thread_id": 12345, "qom_path": "/machine/unattached/device[0]"}], "id": "cmd-1"}`

	var resp QMPResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	var cpus []VCPUInfo
	if err := json.Unmarshal(resp.Return, &cpus); err != nil {
		t.Fatalf("Failed to parse vCPU info: %v", err)
	}

	if len(cpus) != 1 {
		t.Fatalf("Expected 1 CPU, got %d", len(cpus))
	}

	if cpus[0].ThreadID != 12345 {
		t.Errorf("Expected thread_id=12345, got %d", cpus[0].ThreadID)
	}
	if cpus[0].CPU != 0 {
		t.Errorf("Expected CPU=0, got %d", cpus[0].CPU)
	}
}
