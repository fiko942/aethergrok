package system

// InitAppProcessGroup configures OS kernel process tree constraints (e.g. Win32 Job Object)
// to guarantee that all child processes and subprocesses terminate on application exit.
func InitAppProcessGroup() error {
	return initAppProcessGroup()
}

// CleanupOrphanProcesses terminates any lingering descendant processes of this application.
func CleanupOrphanProcesses() {
	cleanupOrphanProcesses()
}
