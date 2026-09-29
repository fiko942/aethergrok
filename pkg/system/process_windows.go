//go:build windows

package system

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var globalJob windows.Handle

func initAppProcessGroup() error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | windows.JOB_OBJECT_LIMIT_BREAKAWAY_OK
	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		_ = windows.CloseHandle(job)
		return err
	}

	err = windows.AssignProcessToJobObject(job, windows.CurrentProcess())
	if err != nil {
		_ = windows.CloseHandle(job)
		return err
	}

	globalJob = job
	return nil
}

func cleanupOrphanProcesses() {
	if globalJob != 0 {
		_ = windows.TerminateJobObject(globalJob, 1)
		_ = windows.CloseHandle(globalJob)
		globalJob = 0
	}

	currentPID := os.Getpid()
	if currentPID <= 0 {
		return
	}

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return
	}

	var children []uint32
	for {
		if int(entry.ParentProcessID) == currentPID && int(entry.ProcessID) != currentPID && entry.ProcessID > 0 {
			children = append(children, entry.ProcessID)
		}
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	for _, childPID := range children {
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(int(childPID)))
		killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = killCmd.Run()

		if h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, childPID); err == nil {
			_ = windows.TerminateProcess(h, 1)
			_ = windows.CloseHandle(h)
		}
	}
}
