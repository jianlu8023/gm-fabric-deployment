//go:build !windows

package pidfile

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

var (
	ErrPIDExists    = errors.New("PID file already exists with an active process")
	ErrLockFailed   = errors.New("failed to acquire a lock on the PID file")
	ErrPidNotExists = errors.New("PID file not exists")
)

type PidFile struct {
	Pid string
}

var lPid *PidFile

func newPidFile(filename string) *PidFile {
	return &PidFile{Pid: filename}
}

// CreateOrUpdatePIDFile ensures that a PID file exists and contains the current process's PID.
// It attempts to create the PID file if it does not exist, and update it if the process is not active.
func CreateOrUpdatePIDFile(filename string) error {
	lPid = newPidFile(filename)
	pid, err := readPIDValue(lPid.Pid)
	if err == nil {
		active, err := isProcessActive(pid)
		if err != nil {
			return fmt.Errorf("error checking process activity: %w", err)
		}
		if active {
			return ErrPIDExists
		}
	}
	return createPIDFile(lPid.Pid)
}

// createPIDFile creates or updates the PID file with the current process's PID.
func createPIDFile(filename string) error {
	pf, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("error opening PID file: %w", err)
	}
	defer pf.Close()

	if err := syscall.Flock(int(pf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return ErrLockFailed
		}
		return fmt.Errorf("error locking PID file: %w", err)
	}

	pid := os.Getpid()
	fmt.Printf("starting process with PID: %d\n", pid)
	if _, err := pf.Write([]byte(strconv.Itoa(pid))); err != nil {
		return fmt.Errorf("error writing pid to PID file: %w", err)
	}

	return nil
}

// isProcessActive checks whether the process with the provided PID is running.
func isProcessActive(pid int) (bool, error) {
	if pid <= 0 {
		return false, errors.New("invalid process ID")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		// On Unix systems, os.FindProcess always succeeds and returns a process with the given pid, irrespective of whether the process exists.
		return false, nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			// The process does not exist
			return false, nil
		}
		if errors.Is(err, os.ErrProcessDone) {
			return false, nil
		}
		// Some other unexpected error
		return false, fmt.Errorf("error signaling process: %w", err)
	}

	// The process exists and is active
	return true, nil
}

// readPIDValue reads the PID value from the specified PID file.
func readPIDValue(filename string) (int, error) {
	value, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return 0, fmt.Errorf("error reading PID file: %w", err)
		}
		return 0, ErrPidNotExists // PID file does not exist
	}
	pid, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, fmt.Errorf("error parsing PID value: %w", err)
	}
	return pid, nil
}

func ReleasePID() {
	if lPid == nil {
		return
	}
	if err := os.Remove(lPid.Pid); err != nil {
		log.Printf("error removing pid file: %v", err)
	}
}
