package terminal

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// conPTY is a thin wrapper over the Windows ConPTY (pseudoconsole) API. It owns
// the pseudoconsole handle, the retained ends of the input/output pipes, the
// process/thread attribute list, and the spawned child process.
//
// A ConPTY gives the child (here `wsl.exe`) a real TTY. WSL2 in turn allocates a
// genuine Linux pty inside the distro, so isatty() is true, Ctrl-C is delivered
// as SIGINT, and \r-based progress output (curl, dnf) renders correctly.
type conPTY struct {
	hpc    windows.Handle // the pseudoconsole
	inW    *os.File       // our end: write bytes here -> child stdin
	outR   *os.File       // our end: read bytes here <- child stdout/stderr
	attrs  *windows.ProcThreadAttributeListContainer
	procPI windows.ProcessInformation
}

// newConPTY creates a pseudoconsole of the given size and returns a wrapper with
// the caller-side pipe ends ready for I/O. The child-side pipe ends are handed to
// the pseudoconsole and closed locally afterwards.
func newConPTY(cols, rows int16) (*conPTY, error) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	// Pipe 1: child reads its stdin from childInR; we write to inW.
	var childInR, inW windows.Handle
	if err := windows.CreatePipe(&childInR, &inW, nil, 0); err != nil {
		return nil, fmt.Errorf("create input pipe: %w", err)
	}

	// Pipe 2: child writes its stdout/stderr to childOutW; we read from outR.
	var outR, childOutW windows.Handle
	if err := windows.CreatePipe(&outR, &childOutW, nil, 0); err != nil {
		windows.CloseHandle(childInR)
		windows.CloseHandle(inW)
		return nil, fmt.Errorf("create output pipe: %w", err)
	}

	var hpc windows.Handle
	size := windows.Coord{X: cols, Y: rows}
	if err := windows.CreatePseudoConsole(size, childInR, childOutW, 0, &hpc); err != nil {
		windows.CloseHandle(childInR)
		windows.CloseHandle(inW)
		windows.CloseHandle(outR)
		windows.CloseHandle(childOutW)
		return nil, fmt.Errorf("create pseudoconsole: %w", err)
	}

	// The pseudoconsole has duplicated the child-side handles; we no longer need them.
	windows.CloseHandle(childInR)
	windows.CloseHandle(childOutW)

	return &conPTY{
		hpc:  hpc,
		inW:  os.NewFile(uintptr(inW), "conpty-in"),
		outR: os.NewFile(uintptr(outR), "conpty-out"),
	}, nil
}

// spawn launches commandLine attached to the pseudoconsole. commandLine must be
// a single already-composed command string (see windows.ComposeCommandLine).
func (c *conPTY) spawn(commandLine string) error {
	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return fmt.Errorf("attribute list: %w", err)
	}

	// For PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE the attribute *value* is the HPCON
	// handle itself (not a pointer to it), with size == sizeof(handle). Passing
	// the handle bits as unsafe.Pointer is the documented Win32 contract; `go vet`
	// flags the uintptr->Pointer conversion but it is correct and intentional here.
	if err := attrs.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(c.hpc),
		unsafe.Sizeof(c.hpc),
	); err != nil {
		attrs.Delete()
		return fmt.Errorf("attach pseudoconsole: %w", err)
	}

	si := &windows.StartupInfoEx{
		StartupInfo:             windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfoEx{}))},
		ProcThreadAttributeList: attrs.List(),
	}

	cmd16, err := windows.UTF16PtrFromString(commandLine)
	if err != nil {
		attrs.Delete()
		return fmt.Errorf("encode command: %w", err)
	}

	var pi windows.ProcessInformation
	// inheritHandles=false: the pseudoconsole plumbs the std handles via the
	// attribute list, so no handle inheritance is needed. EXTENDED_STARTUPINFO_PRESENT
	// tells CreateProcess to read the StartupInfoEx attribute list.
	err = windows.CreateProcess(
		nil,
		cmd16,
		nil,
		nil,
		false,
		windows.EXTENDED_STARTUPINFO_PRESENT|windows.CREATE_UNICODE_ENVIRONMENT,
		nil,
		nil,
		&si.StartupInfo,
		&pi,
	)
	if err != nil {
		attrs.Delete()
		return fmt.Errorf("create process: %w", err)
	}

	c.attrs = attrs
	c.procPI = pi
	return nil
}

// resize changes the pseudoconsole viewport. Safe to call after spawn.
func (c *conPTY) resize(cols, rows int16) error {
	if cols <= 0 || rows <= 0 {
		return nil
	}
	return windows.ResizePseudoConsole(c.hpc, windows.Coord{X: cols, Y: rows})
}

// close tears everything down: input pipe, pseudoconsole, child process, output
// pipe, attribute list, and the process/thread handles. Idempotent-ish; intended
// to be called once.
func (c *conPTY) close() {
	if c.inW != nil {
		c.inW.Close()
	}
	if c.hpc != 0 {
		windows.ClosePseudoConsole(c.hpc)
		c.hpc = 0
	}
	if c.procPI.Process != 0 {
		windows.TerminateProcess(c.procPI.Process, 0)
	}
	if c.outR != nil {
		c.outR.Close()
	}
	if c.attrs != nil {
		c.attrs.Delete()
		c.attrs = nil
	}
	if c.procPI.Thread != 0 {
		windows.CloseHandle(c.procPI.Thread)
		c.procPI.Thread = 0
	}
	if c.procPI.Process != 0 {
		windows.CloseHandle(c.procPI.Process)
		c.procPI.Process = 0
	}
}
