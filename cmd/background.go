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
var PID_FILE_PATH string

func readPIDPath() {
	PID_FILE_PATH = utils.GetPathPrefix() + PID_FILE_NAME
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
	readPIDPath()

	// Get the path of the current binary executable
	binaryExecutablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmdArgs := strings.Fields(args)
	cmdArgs = append(cmdArgs, "--skip-update-check")

	// Add current config path to args if not already explicitly provided
	if !strings.Contains(args, "--config") {
		cmdArgs = append(cmdArgs, "--config", configPath)
	}

	// Run JioTVServer function as a separate process
	cmd := exec.Command(binaryExecutablePath, append([]string{"serve"}, cmdArgs...)...)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Store the PID in a file
	pid := cmd.Process.Pid
	// skipcq: GSC-G302
	err = os.WriteFile(PID_FILE_PATH, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Wait for 1 second to allow the server to start
	time.Sleep(1 * time.Second)

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
	readPIDPath()

	// Read the PID from the file
	pidBytes, err := os.ReadFile(PID_FILE_PATH)
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
	err = os.Remove(PID_FILE_PATH)
	if err != nil {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	fmt.Println("JioTV Go server stopped successfully.")
	return nil
}
