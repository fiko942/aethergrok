//go:build windows

package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type conPtyHandle struct {
	hPC     windows.Handle
	inWrite windows.Handle
	outRead windows.Handle
	process windows.Handle
	thread  windows.Handle
	job     windows.Handle
	pid     int
}

func (c *conPtyHandle) Read(p []byte) (n int, err error) {
	if c.outRead == 0 {
		return 0, io.EOF
	}
	var bytesRead uint32
	err = windows.ReadFile(c.outRead, p, &bytesRead, nil)
	if err != nil {
		return 0, err
	}
	if bytesRead == 0 {
		return 0, io.EOF
	}
	return int(bytesRead), nil
}

func (c *conPtyHandle) Write(p []byte) (n int, err error) {
	if c.inWrite == 0 {
		return 0, io.ErrClosedPipe
	}
	var written uint32
	err = windows.WriteFile(c.inWrite, p, &written, nil)
	if err != nil {
		return 0, err
	}
	return int(written), nil
}

func (c *conPtyHandle) Close() error {
	var firstErr error

	// 1. Terminate all processes in the Job Object if one was created
	if c.job != 0 {
		_ = windows.TerminateJobObject(c.job, 1)
	}

	// 2. Terminate child processes and the root process to unblock I/O
	if c.pid > 0 {
		_ = killChildProcessesOf(c.pid)
	}
	if c.process != 0 {
		_ = windows.TerminateProcess(c.process, 1)
	}

	// 3. Close PseudoConsole to release pipe endpoints
	if c.hPC != 0 {
		windows.ClosePseudoConsole(c.hPC)
		c.hPC = 0
	}

	// 4. Cancel any pending I/O on outRead
	if c.outRead != 0 {
		_ = windows.CancelIoEx(c.outRead, nil)
	}

	// 5. Close pipe handles
	if c.inWrite != 0 {
		if err := windows.CloseHandle(c.inWrite); err != nil && firstErr == nil {
			firstErr = err
		}
		c.inWrite = 0
	}
	if c.outRead != 0 {
		if err := windows.CloseHandle(c.outRead); err != nil && firstErr == nil {
			firstErr = err
		}
		c.outRead = 0
	}

	// 6. Close job handle after killing processes
	if c.job != 0 {
		_ = windows.CloseHandle(c.job)
		c.job = 0
	}

	// 7. Close process and thread handles
	if c.process != 0 {
		_ = windows.CloseHandle(c.process)
		c.process = 0
	}
	if c.thread != 0 {
		_ = windows.CloseHandle(c.thread)
		c.thread = 0
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

	// 1. Try Windows ConPTY (Windows 10 1809+ / Windows 11) for full VT & CLS support
	wrapper, conPtyErr := tryStartConPty(cmd, rows, cols)
	if conPtyErr == nil && wrapper != nil {
		return wrapper, nil
	}

	// 2. Fallback to Anonymous Pipes with enhanced PowerShell setup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	// If powershell, inject ANSI-compatible Clear-Host so CLS clears screen over pipes
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

func createEnvBlock(env []string) *uint16 {
	if len(env) == 0 {
		return nil
	}
	var block []uint16
	for _, e := range env {
		u16, err := windows.UTF16FromString(e)
		if err == nil {
			block = append(block, u16...)
		}
	}
	if len(block) == 0 {
		return nil
	}
	block = append(block, 0)
	return &block[0]
}

func tryStartConPty(cmd *exec.Cmd, rows, cols int) (*osFileWrapper, error) {
	var inRead, inWrite windows.Handle
	var outRead, outWrite windows.Handle

	if err := windows.CreatePipe(&inRead, &inWrite, nil, 0); err != nil {
		return nil, err
	}

	if err := windows.CreatePipe(&outRead, &outWrite, nil, 0); err != nil {
		_ = windows.CloseHandle(inRead)
		_ = windows.CloseHandle(inWrite)
		return nil, err
	}

	coord := windows.Coord{X: int16(cols), Y: int16(rows)}
	var hPC windows.Handle
	if err := windows.CreatePseudoConsole(coord, inRead, outWrite, 0, &hPC); err != nil {
		_ = windows.CloseHandle(inRead)
		_ = windows.CloseHandle(inWrite)
		_ = windows.CloseHandle(outRead)
		_ = windows.CloseHandle(outWrite)
		return nil, err
	}

	// PseudoConsole owns inRead and outWrite now
	_ = windows.CloseHandle(inRead)
	_ = windows.CloseHandle(outWrite)

	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		_ = windows.CloseHandle(inWrite)
		_ = windows.CloseHandle(outRead)
		windows.ClosePseudoConsole(hPC)
		return nil, err
	}
	defer attrList.Delete()

	if err := attrList.Update(windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, unsafe.Pointer(hPC), unsafe.Sizeof(hPC)); err != nil {
		_ = windows.CloseHandle(inWrite)
		_ = windows.CloseHandle(outRead)
		windows.ClosePseudoConsole(hPC)
		return nil, err
	}

	si := new(windows.StartupInfoEx)
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.ProcThreadAttributeList = attrList.List()

	pi := new(windows.ProcessInformation)
	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT)

	var dirPtr *uint16
	if cmd.Dir != "" {
		dirPtr = windows.StringToUTF16Ptr(cmd.Dir)
	}

	envBlock := createEnvBlock(cmd.Env)
	if envBlock != nil {
		flags |= windows.CREATE_UNICODE_ENVIRONMENT
	}

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
	cmdLinePtr := windows.StringToUTF16Ptr(commandLine)

	err = windows.CreateProcess(
		nil,
		cmdLinePtr,
		nil,
		nil,
		false,
		flags,
		envBlock,
		dirPtr,
		&si.StartupInfo,
		pi,
	)
	if err != nil {
		_ = windows.CloseHandle(inWrite)
		_ = windows.CloseHandle(outRead)
		windows.ClosePseudoConsole(hPC)
		return nil, err
	}

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
		_ = windows.AssignProcessToJobObject(jobHandle, pi.Process)
		job = jobHandle
	}

	conPty := &conPtyHandle{
		hPC:     hPC,
		inWrite: inWrite,
		outRead: outRead,
		process: pi.Process,
		thread:  pi.Thread,
		job:     job,
		pid:     int(pi.ProcessId),
	}

	proc, _ := os.FindProcess(int(pi.ProcessId))
	cmd.Process = proc

	wrapper := &osFileWrapper{
		readCloser:  conPty,
		writeCloser: conPty,
		file:        nil,
		hPC:         uintptr(hPC),
	}

	return wrapper, nil
}

func resizePty(file *osFileWrapper, rows, cols int) error {
	if file != nil && file.hPC != 0 {
		return windows.ResizePseudoConsole(windows.Handle(file.hPC), windows.Coord{X: int16(cols), Y: int16(rows)})
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

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return err
	}

	var children []uint32
	for {
		if int(entry.ParentProcessID) == parentPID && int(entry.ProcessID) != parentPID && entry.ProcessID > 0 {
			children = append(children, entry.ProcessID)
		}
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	for _, childPID := range children {
		// Recursively kill child's own descendants first
		_ = killChildProcessesOf(int(childPID))

		// Terminate process tree via taskkill
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(int(childPID)))
		killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = killCmd.Run()

		// Direct TerminateProcess API call
		if h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, childPID); err == nil {
			_ = windows.TerminateProcess(h, 1)
			_ = windows.CloseHandle(h)
		}
	}
	return nil
}

// interruptProcess sends Ctrl+C and terminates any running foreground child processes
func interruptProcess(cmd *exec.Cmd, ptmx *osFileWrapper) error {
	if ptmx != nil {
		// Send standard ETX byte (\x03 / Ctrl+C) into terminal input
		_, _ = ptmx.Write([]byte{3})
	}

	if cmd != nil && cmd.Process != nil && cmd.Process.Pid > 0 {
		pid := cmd.Process.Pid
		_ = windows.GenerateConsoleCtrlEvent(windows.CTRL_C_EVENT, uint32(pid))
		_ = windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(pid))

		// Terminate all descendant processes running under the shell (e.g. Node, React dev server, Vite, Python)
		_ = killChildProcessesOf(pid)
	}

	return nil
}

// killProcessTree terminates the command and its full process tree
func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil || cmd.Process.Pid <= 0 {
		return nil
	}

	pid := cmd.Process.Pid
	// 1. Terminate all descendant child processes
	_ = killChildProcessesOf(pid)

	// 2. Terminate the entire process tree via taskkill
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = killCmd.Run()

	// 3. Fallback to direct process kill
	return cmd.Process.Kill()
}
