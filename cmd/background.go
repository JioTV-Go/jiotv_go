package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const PID_FILE_NAME = "/.do_not_delete_jiotv_go.pid"

// RunInBackground starts the JioTV Go server as a background process by
// executing the current binary with the provided arguments. It stores the
// process ID in a file in the user's home directory so it can be stopped later.
// Returns any errors encountered while starting the process.
func RunInBackground(args string) error {
	fmt.Println("Starting JioTV Go server in background...")

	// get user home directory for storing the PID file
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Get the path of the current binary executable
	binaryExecutablePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd_args := strings.Fields(args)
	// Run JioTVServer function as a separate process
	cmd := exec.Command(binaryExecutablePath, append([]string{"serve"}, cmd_args...)...)
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Store the PID in a file
	pid := cmd.Process.Pid
	// skipcq: GSC-G302
	err = os.WriteFile(homePath+PID_FILE_NAME, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return err
	}

	fmt.Println("JioTV Go server started successfully in background.")

	return nil
}


// StopBackground stops the background JioTV Go server process that was previously
// started with RunInBackground. It reads the PID from the PID file, sends a kill
// signal to that process, and deletes the PID file. Returns any errors encountered.
func StopBackground() error {
	fmt.Println("Stopping JioTV Go server running in background...")

	// get user home directory for storing the PID file
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Read the PID from the file
	pidBytes, err := os.ReadFile(homePath + PID_FILE_NAME)
	if err != nil {
		return err
	}

	// Convert PID bytes to string and then parse as an integer
	pidStr := strings.TrimSpace(string(pidBytes))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Kill()
	if err != nil {
		return err
	}

	// Remove the PID file after successfully killing the process
	err = os.Remove(homePath + PID_FILE_NAME)
	if err != nil {
		return err
	}

	fmt.Println("JioTV Go server stopped successfully.")
	return nil
}
