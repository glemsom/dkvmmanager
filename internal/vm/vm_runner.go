// Package vm provides virtual machine management functionality
package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/hugepages"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// Package-level debug mode flag for the vm package
var debugMode bool

// Package-level dry-run mode flag for the vm package
var dryRunMode bool

// SetDebugMode enables or disables debug mode for the vm package
func SetDebugMode(enabled bool) {
	debugMode = enabled
	if debugMode {
		log.Println("[DEBUG] Debug mode enabled for vm package")
	}
}

// SetDryRunMode enables or disables dry-run mode for the vm package
func SetDryRunMode(enabled bool) {
	dryRunMode = enabled
	if dryRunMode {
		log.Println("[DRY-RUN] Dry-run mode enabled for vm package")
	}
}

// VMRunner manages the lifecycle of a running QEMU virtual machine
type VMRunner struct {
	vm                   *models.VM
	cfg                  *config.Config
	cmd                  *exec.Cmd
	cmdProcess           *os.Process // Cached for race-safe access
	qmpClient            *QMPClient
	socketPath           string
	logChan              chan string
	done                 chan struct{}
	mu                   sync.Mutex
	running              bool
	exitErr              error
	startTime            time.Time
	dryRun               bool
	pciPassthroughConfig models.PCIPassthroughConfig
	usbPassthroughConfig models.USBPassthroughConfig
	cpuOptions           models.CPUOptions
	cpuTopology          models.CPUTopology
	hostCPUTopology      models.HostCPUTopology
	vcpuPinning          models.VCPUPinningGlobal
	startStopScript      models.StartStopScript
	memMB                int64       // Memory in MB for VM (dynamically allocated)
	swtpmProcess         *os.Process // swtpm process, if TPM is enabled
}

// NewVMRunner creates a new VM runner for the given VM
func NewVMRunner(vm *models.VM, cfg *config.Config) *VMRunner {
	r := &VMRunner{
		vm:      vm,
		cfg:     cfg,
		logChan: make(chan string, 256),
		done:    make(chan struct{}),
		memMB:   hugepages.DefaultMemoryMB, // default, overridden by Start()
	}
	if dryRunMode {
		r.dryRun = true
	}
	return r
}

// SetDryRun enables or disables dry-run mode on this runner
func (r *VMRunner) SetDryRun(enabled bool) {
	r.dryRun = enabled
}

// SetStartStopScript sets the start/stop script configuration
func (r *VMRunner) SetStartStopScript(cfg models.StartStopScript) {
	r.startStopScript = cfg
}

// SetVCPUPinning sets global vCPU pinning configuration for this run.
func (r *VMRunner) SetVCPUPinning(p models.VCPUPinningGlobal) {
	r.vcpuPinning = p
}

// SetHostCPUTopology sets the detected host CPU topology for topology-aware CPU allocation
func (r *VMRunner) SetHostCPUTopology(topo models.HostCPUTopology) {
	r.hostCPUTopology = topo
}

// LogChan returns a read-only channel that receives QEMU stdout/stderr lines
func (r *VMRunner) LogChan() <-chan string {
	return r.logChan
}

// IsRunning returns true if the VM process is still running
func (r *VMRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// ExitError returns the error from the QEMU process exit, if any
func (r *VMRunner) ExitError() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.exitErr
}

// StartTime returns when the VM was started
func (r *VMRunner) StartTime() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.startTime
}

// QMPClient returns the QMP client, or nil if not yet connected
func (r *VMRunner) QMPClient() *QMPClient {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.qmpClient
}

// VM returns the VM model
func (r *VMRunner) VM() *models.VM {
	return r.vm
}

// Done returns a channel that is closed when the VM exits
func (r *VMRunner) Done() <-chan struct{} {
	return r.done
}

// MemoryMB returns the allocated memory in MB
func (r *VMRunner) MemoryMB() int64 {
	return r.memMB
}

// VCpuCount returns the number of vCPUs allocated to the VM
func (r *VMRunner) VCpuCount() int {
	// First try to get from CPU topology
	if r.cpuTopology.Enabled && len(r.cpuTopology.SelectedCPUs) > 0 {
		return len(r.cpuTopology.SelectedCPUs)
	}
	// Fall back to vCPU pinning mappings count
	if r.vcpuPinning.Enabled && len(r.vcpuPinning.Mappings) > 0 {
		return len(r.vcpuPinning.Mappings)
	}
	// Default: return 0 if not configured
	return 0
}

// VCPUPinning returns the vCPU pinning configuration
func (r *VMRunner) VCPUPinning() models.VCPUPinningGlobal {
	return r.vcpuPinning
}

// PCIPassthroughDevices returns the PCI passthrough devices
func (r *VMRunner) PCIPassthroughDevices() []models.PCIPassthroughDevice {
	return r.pciPassthroughConfig.Devices
}

// USBPassthroughDevices returns the USB passthrough devices
func (r *VMRunner) USBPassthroughDevices() []models.USBPassthroughDevice {
	return r.usbPassthroughConfig.Devices
}

// ValidateOVMFFiles checks that OVMF_CODE.fd and OVMF_VARS.fd exist in the VM data directory
func (r *VMRunner) ValidateOVMFFiles() error {
	vmDataDir := r.getVMDataDir()

	codePath := filepath.Join(vmDataDir, "OVMF_CODE.fd")
	varsPath := filepath.Join(vmDataDir, "OVMF_VARS.fd")

	if _, err := os.Stat(codePath); err != nil {
		return fmt.Errorf("OVMF_CODE.fd not found in %s: %w", vmDataDir, err)
	}
	if _, err := os.Stat(varsPath); err != nil {
		return fmt.Errorf("OVMF_VARS.fd not found in %s: %w", vmDataDir, err)
	}

	return nil
}

// getVMDataDir returns the data directory for this VM
func (r *VMRunner) getVMDataDir() string {
	return filepath.Join(r.cfg.DataFolder, "vms", r.vm.ID)
}

// waitForSocket polls until a Unix socket file exists or timeout is reached
func (r *VMRunner) waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket: %s", path)
}

// startTPM starts the swtpm process for this VM.
// TPM state in {vmDataDir}/tpm/ persists across VM restarts.
func (r *VMRunner) startTPM(vmDataDir string) error {
	pidFile := "swtpm.pid"
	tpmDir := filepath.Join(vmDataDir, "tpm")
	tpmSock := filepath.Join(vmDataDir, "tpm.sock")
	pidPath := filepath.Join(tpmDir, pidFile)

	// (A) Ensure state directory exists
	if err := os.MkdirAll(tpmDir, 0700); err != nil {
		return fmt.Errorf("failed to create TPM state dir: %w", err)
	}
	log.Printf("[INFO] TPM state dir: %s (persistent)", tpmDir)

	// (B) Remove stale socket (safe, does not affect running process)
	os.Remove(tpmSock)

	// (C) Detect and kill orphaned swtpm from previous run
	if pidData, err := os.ReadFile(pidPath); err == nil && len(pidData) > 0 {
		pid, perr := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if perr == nil && pid > 0 {
			proc, ferr := os.FindProcess(pid)
			if ferr == nil && proc != nil {
				// Check if process is alive (signal 0)
				if err := proc.Signal(syscall.Signal(0)); err == nil {
					log.Printf("[WARN] Orphaned swtpm PID %d detected for VM %s – killing", pid, r.vm.Name)
					_ = proc.Kill()
					_, _ = proc.Wait() // reap
				}
			}
		}
		os.Remove(pidPath) // always remove stale PID file
	}

	// (D) Start swtpm
	cmd := exec.Command(r.cfg.TPMBinary,
		"socket",
		"--tpm2",
		"--ctrl",
		fmt.Sprintf("type=unixio,path=%s", tpmSock),
		"--flags",
		"not-need-init",
		"--tpmstate",
		fmt.Sprintf("dir=%s", tpmDir),
		"--log",
		fmt.Sprintf("level=20,file=%s/swtpm.log", tpmDir),
	)
	if err := cmd.Start(); err != nil {
		// Transient cleanup only – do NOT delete tpmDir (persistent state)
		os.Remove(tpmSock)
		return fmt.Errorf("failed to start swtpm: %w", err)
	}

	// (E) Cache process handle and write PID file
	r.mu.Lock()
	r.swtpmProcess = cmd.Process
	r.mu.Unlock()

	_ = os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
	if debugMode {
		log.Printf("[DEBUG] TPM PID file written: %s", pidPath)
	}

	// (F) Wait for socket to appear
	if err := r.waitForSocket(tpmSock, 5*time.Second); err != nil {
		r.mu.Lock()
		proc := r.swtpmProcess
		r.swtpmProcess = nil
		r.mu.Unlock()
		if proc != nil {
			_ = proc.Kill()
			// Reap the process to avoid zombie
			_, _ = proc.Wait()
		}
		os.Remove(tpmSock)
		os.Remove(pidPath)
		return fmt.Errorf("swtpm socket not ready: %w", err)
	}

	// Verify swtpm is still running (it could crash immediately after creating socket)
	r.mu.Lock()
	proc := r.swtpmProcess
	r.mu.Unlock()
	if proc != nil {
		// Check if process is still alive
		err := proc.Signal(syscall.Signal(0))
		if err != nil {
			// Process died
			r.mu.Lock()
			r.swtpmProcess = nil
			r.mu.Unlock()
			// Try to wait for it to reap
			_, _ = proc.Wait()
			os.Remove(tpmSock)
			os.Remove(pidPath)
			return fmt.Errorf("swtpm process terminated unexpectedly")
		}
	}

	log.Printf("[INFO] TPM started for VM %s (PID: %d, state: %s)", r.vm.Name, cmd.Process.Pid, tpmDir)

	return nil
}

// cleanupTPM terminates the swtpm process and removes transient runtime files.
// Persistent TPM state in {vmDataDir}/tpm/ is preserved.
func (r *VMRunner) cleanupTPM() {
	r.mu.Lock()
	proc := r.swtpmProcess
	r.swtpmProcess = nil
	r.mu.Unlock()

	if proc == nil {
		return
	}

	vmDataDir := r.getVMDataDir()
	tpmDir := filepath.Join(vmDataDir, "tpm")
	tpmSock := filepath.Join(vmDataDir, "tpm.sock")
	pidPath := filepath.Join(tpmDir, "swtpm.pid")

	log.Printf("[INFO] TPM stopping for VM %s (preserving state in %s)", r.vm.Name, tpmDir)

	// Attempt graceful shutdown via control channel (best-effort, ignores errors)
	r.shutdownTPMControl(tpmDir)

	// Signal termination
	_ = proc.Signal(syscall.SIGTERM)
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = proc.Kill()
		<-done
	}

	// Remove transient runtime files only – DO NOT remove tpmDir (persistent state)
	os.Remove(tpmSock)
	// ctrlSock was removed — now tpm.sock is used for both control and TPM
	os.Remove(pidPath)

	if debugMode {
		log.Printf("[DEBUG] TPM cleaned up for VM %s (state preserved)", r.vm.Name)
	}
}

// shutdownTPMControl attempts a graceful TPM shutdown via the swtpm control channel.
// It is best-effort and ignores errors (SIGTERM fallback handles the rest).
// CMD_SHUTDOWN = 0x00000003 (big-endian 4-byte command).
func (r *VMRunner) shutdownTPMControl(tpmDir string) error {
	ctrlPath := filepath.Join(tpmDir, "tpm.sock")

	conn, err := net.DialTimeout("unix", ctrlPath, 500*time.Millisecond)
	if err != nil {
		if debugMode {
			log.Printf("[DEBUG] TPM control socket not available: %v", err)
		}
		return err
	}
	defer conn.Close()

	// CMD_SHUTDOWN = 0x00000003 (big-endian)
	cmd := []byte{0x00, 0x00, 0x00, 0x03}
	if err := binary.Write(conn, binary.BigEndian, cmd); err != nil {
		if debugMode {
			log.Printf("[DEBUG] Failed to send CMD_SHUTDOWN: %v", err)
		}
		return err
	}

	// Read response (1 byte status)
	var status byte
	if err := binary.Read(conn, binary.BigEndian, &status); err != nil {
		if debugMode {
			log.Printf("[DEBUG] Failed to read CMD_SHUTDOWN response: %v", err)
		}
		return err
	}

	if status != 0 {
		log.Printf("[WARN] CMD_SHUTDOWN returned non-zero status: 0x%02x", status)
		return fmt.Errorf("shutdown failed, status=0x%02x", status)
	}

	if debugMode {
		log.Printf("[DEBUG] TPM control shutdown successful for VM %s", r.vm.Name)
	}
	return nil
}

// Start launches the QEMU process and connects QMP
func (r *VMRunner) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("VM %s is already running", r.vm.Name)
	}
	r.mu.Unlock()

	// Check hugepages availability for VM memory
	hugepagesCfg, err := hugepages.NewAutoConfig()
	if err != nil {
		return fmt.Errorf("failed to configure VM memory: %w", err)
	}
	// Store memory size for QEMU args
	r.memMB = hugepagesCfg.MemMB

	result, err := hugepages.Check()
	if err != nil {
		return fmt.Errorf("hugepages check failed: %w", err)
	}

	result.RequiredPages = hugepagesCfg.RequiredPages()
	result.IsSufficient = result.AvailablePages >= result.RequiredPages
	if !result.IsSufficient {
		// Try to allocate hugepages
		result, err = hugepages.Ensure(hugepagesCfg)
		if err != nil || !result.IsSufficient {
			return fmt.Errorf(
				"insufficient hugepages for VM %q: have %d pages (%dMB × 2MB pages), need %d pages (try: echo %d > /proc/sys/vm/nr_hugepages)",
				r.vm.Name,
				result.AvailablePages,
				r.memMB,
				result.RequiredPages,
				result.RequiredPages,
			)
		}
	}

	// Validate OVMF files
	if err := r.ValidateOVMFFiles(); err != nil {
		return err
	}

	// Execute start script (block VM start if it fails)
	if err := r.executeStartScript(); err != nil {
		return fmt.Errorf("start script failed: %w", err)
	}

	vmDataDir := r.getVMDataDir()
	r.socketPath = filepath.Join("/tmp", fmt.Sprintf("dkvm-%s.sock", r.vm.ID))

	// Clean up stale socket
	os.Remove(r.socketPath)

	// Start TPM if enabled
	var qemuStarted bool
	if r.vm.TPMEnabled {
		if err := r.startTPM(vmDataDir); err != nil {
			return fmt.Errorf("failed to start TPM: %w", err)
		}
		// Defer cleanup if we return error before QEMU starts
		qemuStarted = false
		defer func() {
			if !qemuStarted {
				r.cleanupTPM()
			}
		}()
	}

	// Build QEMU arguments
	args := r.buildQEMUArgs(vmDataDir)

	if debugMode || r.dryRun {
		log.Printf("[DEBUG] QEMU command: %s %s", r.cfg.QEMUPath, strings.Join(args, " "))
	}

	if r.dryRun {
		filtered := filterPassthroughArgs(args)
		fullCmd := fmt.Sprintf("%s %s", r.cfg.QEMUPath, strings.Join(args, " "))
		filteredCmd := fmt.Sprintf("%s %s", r.cfg.QEMUPath, strings.Join(filtered, " "))
		log.Printf("[DRY-RUN] Full command: %s", fullCmd)
		log.Printf("[DRY-RUN] Filtered command: %s", filteredCmd)
		r.logChan <- "[DRY-RUN] Full QEMU command:"
		r.logChan <- fullCmd
		r.logChan <- "[DRY-RUN] Filtered QEMU command (no passthrough):"
		r.logChan <- filteredCmd
		return nil
	}

	// Create command
	r.cmd = exec.Command(r.cfg.QEMUPath, args...)

	// Set up stdout pipe
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Set up stderr pipe
	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start QEMU
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start QEMU: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] QEMU process started: PID=%d, socket=%s", r.cmd.Process.Pid, r.socketPath)
	}

	// TPM: mark as started so defer doesn't clean up
	if r.vm.TPMEnabled {
		qemuStarted = true
	}

	r.mu.Lock()
	r.running = true
	r.cmdProcess = r.cmd.Process
	r.startTime = time.Now()
	r.mu.Unlock()

	// Read stdout/stderr in background
	go r.readOutput(stdout, "stdout")
	go r.readOutput(stderr, "stderr")

	// Monitor process exit
	go r.monitorProcess()

	// Start QMP watchdog to diagnose connection issues
	if debugMode {
		go r.qmpWatchdog()
	}

	// Wait for QMP socket and connect
	go r.connectQMP()

	return nil
}

// Stop gracefully shuts down the VM via QMP quit command
func (r *VMRunner) Stop() error {
	r.mu.Lock()
	client := r.qmpClient
	running := r.running
	r.cmdProcess = r.cmd.Process
	r.mu.Unlock()

	if !running {
		return fmt.Errorf("VM %s is not running", r.vm.Name)
	}

	if client != nil {
		if err := client.Quit(); err != nil {
			// If QMP quit fails, kill the process
			r.mu.Lock()
			proc := r.cmdProcess
			r.mu.Unlock()
			if proc != nil {
				proc.Kill()
			}
			return fmt.Errorf("QMP quit failed, killed process: %w", err)
		}
	} else {
		r.mu.Lock()
		proc := r.cmdProcess
		r.mu.Unlock()
		// No QMP client, kill directly
		if proc != nil {
			proc.Kill()
		}
	}

	// Cleanup TPM synchronously
	r.cleanupTPM()

	return nil
}

// ForceStop kills the QEMU process immediately
func (r *VMRunner) ForceStop() error {
	r.mu.Lock()
	if !r.running || r.cmd == nil || r.cmd.Process == nil {
		r.mu.Unlock()
		return fmt.Errorf("VM %s is not running", r.vm.Name)
	}
	proc := r.cmd.Process
	r.mu.Unlock()

	// Kill QEMU
	if err := proc.Kill(); err != nil {
		return err
	}

	// Cleanup TPM synchronously
	r.cleanupTPM()

	return nil
}

// Cleanup removes stale resources (socket files, etc.)
func (r *VMRunner) Cleanup() {
	if r.socketPath != "" {
		os.Remove(r.socketPath)
	}
	// Ensure TPM process is terminated and state cleaned up
	r.cleanupTPM()
}

// readOutput reads lines from a pipe and sends them to the log channel
func (r *VMRunner) readOutput(pipe io.Reader, source string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		prefix := ""
		if source == "stderr" {
			prefix = "[stderr] "
		}
		// Log output when debug mode is enabled
		if debugMode {
			log.Printf("[DEBUG] [%s] %s", source, line)
		}
		select {
		case r.logChan <- prefix + line:
		default:
			// Channel full, drop oldest
			select {
			case <-r.logChan:
			default:
			}
			r.logChan <- prefix + line
		}
	}
	// Scanner ended - check error
	if err := scanner.Err(); err != nil {
		log.Printf("[DEBUG] readOutput(%s) error: %v", source, err)
	}
}

// readScriptOutput reads lines from a script pipe and writes them to the debug log
func (r *VMRunner) readScriptOutput(pipe io.Reader, source string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		if debugMode {
			log.Printf("[DEBUG] %s: %s", source, line)
		}
	}
}

// monitorProcess waits for the QEMU process to exit
func (r *VMRunner) monitorProcess() {
	err := r.cmd.Wait()

	r.mu.Lock()
	r.running = false
	r.exitErr = err
	client := r.qmpClient
	r.qmpClient = nil // Clear to prevent race with Stop()
	r.mu.Unlock()

	// Clean up QMP (outside lock to avoid holding lock during I/O)
	if client != nil {
		client.Close()
	}

	// Clean up socket
	os.Remove(r.socketPath)

	// Execute stop script (non-blocking - don't fail if it errors)
	r.executeStopScript()

	// Cleanup TPM process
	r.cleanupTPM()

	close(r.done)

	if debugMode {
		log.Printf("[DEBUG] QEMU process exited for VM %s: %v", r.vm.Name, err)
	}
}

// qmpWatchdog periodically checks QMP connection status and logs diagnostics
func (r *VMRunner) qmpWatchdog() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		running := r.running
		proc := r.cmdProcess
		r.mu.Unlock()

		if !running {
			return // VM stopped
		}

		// Check if QEMU process is still alive
		if proc != nil {
			if err := proc.Signal(syscall.Signal(0)); err != nil {
				log.Printf("[DEBUG] QMP watchdog: QEMU process no longer alive: %v", err)
				return
			}
		}

		// Check if QMP socket exists
		if _, err := os.Stat(r.socketPath); os.IsNotExist(err) {
			log.Printf("[DEBUG] QMP watchdog: QMP socket not yet created")
			continue
		}

		// Check if QMP client connected
		r.mu.Lock()
		qmpClient := r.qmpClient
		r.mu.Unlock()
		if qmpClient == nil {
			log.Printf("[DEBUG] QMP watchdog: QEMU running but QMP not yet connected (socket exists)")
		} else {
			// QMP connected, watchdog no longer needed
			return
		}
	}
}

// connectQMP waits for the QMP socket to appear, then negotiates
func (r *VMRunner) connectQMP() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		// Check if QEMU process is still alive before proceeding
		r.mu.Lock()
		running := r.running
		proc := r.cmdProcess
		r.mu.Unlock()
		
		if !running || proc == nil {
			if debugMode {
				log.Printf("[DEBUG] QMP connect aborted: VM no longer running or process nil")
			}
			return
		}
		
		// Verify QEMU process is still alive (signal 0)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			log.Printf("[DEBUG] QMP connect aborted: QEMU process died")
			return
		}

		if _, err := os.Stat(r.socketPath); err == nil {
			// Socket exists, try to connect
			time.Sleep(500 * time.Millisecond) // Brief wait for QEMU to be ready

			client, err := NewQMPClient(r.socketPath)
			if err != nil {
				if debugMode {
					log.Printf("[DEBUG] QMP connect attempt %d failed: %v", i+1, err)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			// Negotiate
			greeting, err := client.Negotiate()
			if err != nil {
				client.Close()
				if debugMode {
					log.Printf("[DEBUG] QMP negotiate attempt %d failed: %v", i+1, err)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			if debugMode {
				log.Printf("[DEBUG] QMP connected: QEMU %d.%d.%d",
					greeting.QMP.Version.QEMU.Major,
					greeting.QMP.Version.QEMU.Minor,
					greeting.QMP.Version.QEMU.Micro)
			}

			r.mu.Lock()
			r.qmpClient = client
			r.mu.Unlock()

			r.logChan <- "[QMP] Connected to QEMU monitor"
			if err := r.ApplyVCPUPinning(r.vcpuPinning); err != nil {
				r.logChan <- fmt.Sprintf("[vCPU pinning] WARNING: %v", err)
			}
			return
		}
		time.Sleep(1 * time.Second)
	}

	r.logChan <- "[QMP] WARNING: Failed to connect to QMP after timeout"
}

// filterPassthroughArgs removes PCI/USB passthrough and drive ROM arguments
// from the QEMU argument list. It returns a new slice without mutating the input.
// Removed patterns:
//   - -device vfio-pci,... (PCI passthrough)
//   - -device usb-host,... (USB passthrough)
//   - -drive ...romfile=... pairs (drive with ROM file)
func filterPassthroughArgs(args []string) []string {
	filtered := make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		if args[i] == "-device" && i+1 < len(args) {
			val := args[i+1]
			if strings.Contains(val, "vfio-pci") || strings.Contains(val, "usb-host") {
				i += 2 // skip flag and value
				continue
			}
		}
		if args[i] == "-drive" && i+1 < len(args) {
			if strings.Contains(args[i+1], "romfile=") {
				i += 2 // skip flag and value
				continue
			}
		}
		filtered = append(filtered, args[i])
		i++
	}
	return filtered
}

// executeStartScript executes the start script before QEMU launches
func (r *VMRunner) executeStartScript() error {
	// Skip if no custom script config
	if r.startStopScript.StartScript == "" && r.startStopScript.UseBuiltin {
		// Use builtin but no PCI devices configured - that's ok
		return nil
	}

	if r.startStopScript.UseBuiltin {
		// Generate builtin script from PCI devices
		script, err := GenerateBuiltinScript(r.pciPassthroughConfig.Devices)
		if err != nil {
			return fmt.Errorf("failed to generate builtin script: %w", err)
		}

		// Write to temp file
		scriptPath := fmt.Sprintf("/tmp/dkvm-builtin-%s.sh", r.vm.ID)
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			return fmt.Errorf("failed to write builtin script: %w", err)
		}

		// Execute the script
		cmd := exec.Command("/bin/bash", scriptPath)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to get stderr pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start script: %w", err)
		}

		// Read output in goroutines to avoid blocking
		go r.readScriptOutput(stdout, "start script stdout")
		go r.readScriptOutput(stderr, "start script stderr")

		if err := cmd.Wait(); err != nil {
			if debugMode {
				log.Printf("[DEBUG] start script failed: %v", err)
			}
			r.logChan <- fmt.Sprintf("[start script] failed: %v", err)
			return fmt.Errorf("start script execution failed: %w", err)
		}

		r.logChan <- fmt.Sprintf("[start script] executed builtin script: %s", scriptPath)
		if debugMode {
			log.Printf("[DEBUG] start script executed builtin script: %s", scriptPath)
		}

		// Clean up temp file
		os.Remove(scriptPath)

	} else if r.startStopScript.StartScript != "" {
		// Custom script - pass PCI devices as command-line arguments
		args := []string{"/bin/bash", r.startStopScript.StartScript, "start"}
		for _, dev := range r.pciPassthroughConfig.Devices {
			args = append(args, dev.Address)
		}
		cmd := exec.Command(args[0], args[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to get stderr pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start script: %w", err)
		}

		// Read output in goroutines to avoid blocking
		go r.readScriptOutput(stdout, "start script stdout")
		go r.readScriptOutput(stderr, "start script stderr")

		if err := cmd.Wait(); err != nil {
			if debugMode {
				log.Printf("[DEBUG] start script failed: %v", err)
			}
			r.logChan <- fmt.Sprintf("[start script] failed: %v", err)
			return fmt.Errorf("start script execution failed: %w", err)
		}

		r.logChan <- fmt.Sprintf("[start script] executed: %s", r.startStopScript.StartScript)
		if debugMode {
			log.Printf("[DEBUG] start script executed: %s", r.startStopScript.StartScript)
		}
	}

	return nil
}

// executeStopScript executes the stop script after QEMU exits (non-blocking)
func (r *VMRunner) executeStopScript() {
	// Skip if no custom script config
	if r.startStopScript.StopScript == "" && r.startStopScript.UseBuiltin {
		return
	}

	if r.startStopScript.UseBuiltin {
		// Generate builtin stop script
		script := GenerateBuiltinStopScript()

		// Write to temp file
		scriptPath := fmt.Sprintf("/tmp/dkvm-builtin-stop-%s.sh", r.vm.ID)
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			log.Printf("[stop script] WARNING: failed to write builtin script: %v", err)
			return
		}

		// Execute the script
		cmd := exec.Command("/bin/bash", scriptPath)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("[stop script] WARNING: failed to get stdout pipe: %v", err)
			os.Remove(scriptPath)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("[stop script] WARNING: failed to get stderr pipe: %v", err)
			os.Remove(scriptPath)
			return
		}
		if err := cmd.Start(); err != nil {
			log.Printf("[stop script] WARNING: failed to start script: %v", err)
			os.Remove(scriptPath)
			return
		}

		// Read output in goroutines to avoid blocking
		go r.readScriptOutput(stdout, "stop script stdout")
		go r.readScriptOutput(stderr, "stop script stderr")

		if err := cmd.Wait(); err != nil {
			if debugMode {
				log.Printf("[DEBUG] stop script failed: %v", err)
			}
			r.logChan <- fmt.Sprintf("[stop script] WARNING: %v", err)
		}

		// Clean up temp file
		os.Remove(scriptPath)

	} else if r.startStopScript.StopScript != "" {
		// Custom stop script (non-blocking) - pass PCI devices as CLI args
		args := []string{"/bin/bash", r.startStopScript.StopScript, "stop"}
		for _, dev := range r.pciPassthroughConfig.Devices {
			args = append(args, dev.Address)
		}
		cmd := exec.Command(args[0], args[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			if debugMode {
				log.Printf("[DEBUG] stop script stdout pipe error: %v", err)
			}
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			if debugMode {
				log.Printf("[DEBUG] stop script stderr pipe error: %v", err)
			}
			return
		}
		if err := cmd.Start(); err != nil {
			if debugMode {
				log.Printf("[DEBUG] stop script start error: %v", err)
			}
			return
		}

		// Read output in goroutines to avoid blocking
		go r.readScriptOutput(stdout, "stop script stdout")
		go r.readScriptOutput(stderr, "stop script stderr")

		if err := cmd.Wait(); err != nil {
			if debugMode {
				log.Printf("[DEBUG] stop script failed: %v", err)
			}
			r.logChan <- fmt.Sprintf("[stop script] WARNING: %v", err)
		}
	}
}
