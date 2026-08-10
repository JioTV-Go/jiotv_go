package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

var PID_FILE_NAME = ".jiotv_go.pid"

func getPIDPath() string {
	return utils.GetPathPrefix() + PID_FILE_NAME
}

// RunInBackground starts the JioTV Go server as a background process by
// executing the current binary with the provided arguments. It stores the
// process ID in a file in the user's home directory so it can be stopped later.
// Returns any errors encountered while starting the process.
func RunInBackground(args string, configPath string) error {
	if err := config.Cfg.Load(configPath); err != nil {
		return err
	}

	fmt.Println("Starting JioTV Go server in background...")
	pidPath := getPIDPath()

	// Get the path of the current binary executable
	binaryExecutablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// --skip-update-check and --config are global flags: urfave/cli accepts them
	// only before the subcommand, so they must precede "serve".
	cmdArgs := []string{"--skip-update-check"}
	if configPath != "" && !strings.Contains(args, "--config") {
		cmdArgs = append(cmdArgs, "--config", configPath)
	}
	cmdArgs = append(cmdArgs, "serve")
	cmdArgs = append(cmdArgs, strings.Fields(args)...)

	// Run JioTVServer function as a separate process
	cmd := exec.Command(binaryExecutablePath, cmdArgs...)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Store the PID in a file
	pid := cmd.Process.Pid
	// skipcq: GSC-G302
	err = os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	// Surface an immediate crash (bad flags, port in use) instead of reporting
	// a success for a process that is already gone.
	waitResult := make(chan error, 1)
	go func() { waitResult <- cmd.Wait() }()
	select {
	case waitErr := <-waitResult:
		os.Remove(pidPath)
		return fmt.Errorf("server exited immediately after start: %w", waitErr)
	case <-time.After(1 * time.Second):
	}

	fmt.Println("JioTV Go server started successfully in background.")

	return nil
}

// StopBackground stops the background JioTV Go server process that was previously
// started with RunInBackground. It reads the PID from the PID file, sends a kill
// signal to that process, and deletes the PID file. Returns any errors encountered.
func StopBackground(configPath string) error {
	if err := config.Cfg.Load(configPath); err != nil {
		return err
	}

	fmt.Println("Stopping JioTV Go server running in background...")
	pidPath := getPIDPath()

	// Read the PID from the file
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// Convert PID bytes to string and then parse as an integer
	pidStr := strings.TrimSpace(string(pidBytes))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("failed to convert PID to integer: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find JioTV Go process: %w", err)
	}

	// Send a kill signal to the process
	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill JioTV Go process: %w", err)
	}

	// Remove the PID file after successfully killing the process
	err = os.Remove(pidPath)
	if err != nil {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	fmt.Println("JioTV Go server stopped successfully.")
	return nil
}
