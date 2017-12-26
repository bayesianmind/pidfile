package pidfile

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	ErrProcessRunning = errors.New("process is running")
	ErrFileStale      = errors.New("pidfile exists but process is not running")
	ErrFileInvalid    = errors.New("pidfile has invalid contents")
)

// Remove a pidfile
func Remove(filename string) error {
	return os.RemoveAll(filename)
}

// IsRunning returns true if the pidfile exists and is running
func IsRunning(filename string) (bool, error) {
	// Check for existing pid
	pid, err := pidfileContents(filename)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}

	return pidIsRunning(pid), nil
}

// Write writes a pidfile, returning an error
// if the process is already running or pidfile is orphaned
func Write(filename string) error {
	return WriteControl(filename, os.Getpid(), false)
}

func WriteControl(filename string, pid int, overwrite bool) error {
	// Check for existing pid
	oldpid, err := pidfileContents(filename)
	if err != nil && !isNotFoundErr(err) {
		return err
	}

	// We have a pid
	if err == nil {
		if pidIsRunning(oldpid) {
			return ErrProcessRunning
		}
		if !overwrite {
			return ErrFileStale
		}
	}

	// We're clear to (over)write the file
	err = ioutil.WriteFile(filename, []byte(fmt.Sprintf("%d\n", pid)), 0644)
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		fmt.Printf("got unexpected not found error type=%t val:%+v\n", err, err)
	}
	return err
}
func isNotFoundErr(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	// on some platforms we get this back from ioutil.ReadFile()
	return strings.Contains(err.Error(), "no such file or directory")
}

func pidfileContents(filename string) (int, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, ErrFileInvalid
	}

	return pid, nil
}

func pidIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	switch {
	case err == nil:
		return true
	case err.Error() == "no such process",
		err.Error() == "os: process already finished",
		err.Error() == "operation not permitted":
		return false
	default:
		return true
	}
}
