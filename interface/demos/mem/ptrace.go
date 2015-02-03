package main

import (
	"os"
	"syscall"
)

type Ptrace struct {
	proc *os.Process
	addr uint64
}

func ForkExec(cmd string, argv []string) (*Ptrace, error) {
	var err error

	pt := new(Ptrace)
	attr := &os.ProcAttr{
		Env: os.Environ(),
		Sys: &syscall.SysProcAttr{Ptrace: true},
	}
	args := []string{cmd}
	args = append(args, argv...)
	pt.proc, err = os.StartProcess(cmd, args, attr)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (pt *Ptrace) Read(b []byte) (n int, err error) {
	n, err = syscall.PtracePeekText(pt.proc.Pid, uintptr(pt.addr), b)
	if err != nil {
		return 0, err
	}
	pt.addr += uint64(n)
	return n, nil
}

func (pt *Ptrace) Seek(addr uint64) {
	pt.addr = addr
}
