//go:build windows

package terminal

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/UserExistsError/conpty"
	"golang.org/x/sys/windows"
)

type conPtyInstance struct {
	cpty *conpty.ConPty
	pid  int
	job  windows.Handle
}

func (c *conPtyInstance) Read(p []byte) (n int, err error) {
	if c.cpty == nil {
		return 0, fmt.Errorf("pty closed")
	}
	return c.cpty.Read(p)
}

func (c *conPtyInstance) Write(p []byte) (n int, err error) {
	if c.cpty == nil {
		return 0, fmt.Errorf("pty closed")
	}
	return c.cpty.Write(p)
}

func (c *conPtyInstance) Close() error {
	var firstErr error

	// 1. Terminate all processes in Job Object
	if c.job != 0 {
		_ = windows.TerminateJobObject(c.job, 1)
	}

	// 2. Kill child processes
	if c.pid > 0 {
		_ = killChildProcessesOf(c.pid)
	}

	// 3. Close ConPTY instance
	if c.cpty != nil {
		if err := c.cpty.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.cpty = nil
	}

	// 4. Close Job Object
	if c.job != 0 {
		_ = windows.CloseHandle(c.job)
		c.job = 0
	}

	return firstErr
}

func startPty(cmd *exec.Cmd, rows, cols int) (*osFileWrapper, error) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	// 1. Try Windows ConPTY via robust conpty engine
	if conpty.IsConPtyAvailable() {
		commandLine := cmd.Path
		if len(cmd.Args) > 0 {
			var quotedArgs []string
			for _, arg := range cmd.Args {
				if strings.Contains(arg, " ") && !strings.HasPrefix(arg, `"`) {
					quotedArgs = append(quotedArgs, fmt.Sprintf(`"%s"`, arg))
				} else {
					quotedArgs = append(quotedArgs, arg)
				}
			}
			commandLine = strings.Join(quotedArgs, " ")
		}

		var options []conpty.ConPtyOption
		options = append(options, conpty.ConPtyDimensions(cols, rows))
		if cmd.Dir != "" {
			options = append(options, conpty.ConPtyWorkDir(cmd.Dir))
		}
		if len(cmd.Env) > 0 {
			options = append(options, conpty.ConPtyEnv(cmd.Env))
		}

		cpty, err := conpty.Start(commandLine, options...)
		if err == nil && cpty != nil {
			pid := cpty.Pid()

			var job windows.Handle
			jobHandle, jobErr := windows.CreateJobObject(nil, nil)
			if jobErr == nil {
				var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
				info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | windows.JOB_OBJECT_LIMIT_BREAKAWAY_OK
				_, _ = windows.SetInformationJobObject(
					jobHandle,
					windows.JobObjectExtendedLimitInformation,
					uintptr(unsafe.Pointer(&info)),
					uint32(unsafe.Sizeof(info)),
				)

				pHandle, pErr := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(pid))
				if pErr == nil {
					_ = windows.AssignProcessToJobObject(jobHandle, pHandle)
					_ = windows.CloseHandle(pHandle)
				}
				job = jobHandle
			}

			conInst := &conPtyInstance{
				cpty: cpty,
				pid:  pid,
				job:  job,
			}

			return &osFileWrapper{
				readCloser:  conInst,
				writeCloser: conInst,
				file:        nil,
			}, nil
		}
	}

	// 2. Fallback to Standard Anonymous Pipes with enhanced PowerShell setup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if strings.Contains(strings.ToLower(cmd.Path), "powershell") && len(cmd.Args) <= 1 {
		cmd.Args = []string{
			cmd.Path,
			"-NoLogo",
			"-NoExit",
			"-Command",
			"function Clear-Host { [Console]::Write([char]27 + '[2J' + [char]27 + '[H' + [char]27 + '[3J') }; Set-Alias -Name cls -Value Clear-Host -Option AllScope -Force",
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &osFileWrapper{
		readCloser:  stdout,
		writeCloser: stdin,
		file:        nil,
	}, nil
}

func resizePty(file *osFileWrapper, rows, cols int) error {
	if file != nil && file.readCloser != nil {
		if cptyInst, ok := file.readCloser.(*conPtyInstance); ok && cptyInst.cpty != nil {
			return cptyInst.cpty.Resize(cols, rows)
		}
	}
	return nil
}

// interruptProcess sends Ctrl+C and terminates running child processes on Windows
func interruptProcess(cmd *exec.Cmd, ptmx *osFileWrapper) error {
	if ptmx != nil {
		_, _ = ptmx.Write([]byte{3}) // \x03 (Ctrl+C)
	}

	// If using ConPTY, kill child processes of the ConPTY process
	if ptmx != nil && ptmx.readCloser != nil {
		if cptyInst, ok := ptmx.readCloser.(*conPtyInstance); ok && cptyInst.pid > 0 {
			_ = killChildProcessesOf(cptyInst.pid)
		}
	}

	if cmd != nil && cmd.Process != nil && cmd.Process.Pid > 0 {
		_ = killChildProcessesOf(cmd.Process.Pid)
	}

	return nil
}

// killProcessTree terminates the command and its full process tree on Windows
func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid
	if pid > 0 {
		_ = killChildProcessesOf(pid)
		_ = cmd.Process.Kill()
	}

	return nil
}

// killChildProcessesOf traverses the process tree and terminates all descendants of parentPID
func killChildProcessesOf(parentPID int) error {
	if parentPID <= 0 {
		return nil
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		if int(entry.ParentProcessID) == parentPID {
			childPID := int(entry.ProcessID)
			// Recursively terminate grand-children first
			_ = killChildProcessesOf(childPID)
			if hProc, openErr := windows.OpenProcess(windows.PROCESS_TERMINATE, false, entry.ProcessID); openErr == nil {
				_ = windows.TerminateProcess(hProc, 1)
				_ = windows.CloseHandle(hProc)
			}
		}
		err = windows.Process32Next(snapshot, &entry)
	}
	return nil
}
