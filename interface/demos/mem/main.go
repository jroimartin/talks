package main

import (
	"debug/elf"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"

	"github.com/jroimartin/gocui"
)

const blocksize = 512

var (
	pt        *Ptrace
	startAddr uint64 = math.MaxUint64
	endAddr   uint64
	curAddr   uint64
)

func main() {
	runtime.LockOSThread()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mem cmd args...")
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	if err := initPtrace(cmd, args...); err != nil {
		log.Fatalln(err)
	}

	if err := guiLoop(); err != nil {
		log.Fatalln(err)
	}
}

// initPtrace creates a new process and attaches to it via PTRACE, it also gets
// the memory bounds of the process based on the PT_LOAD program headers.
func initPtrace(cmd string, args ...string) error {
	var err error

	f, err := elf.Open(cmd)
	if err != nil {
		return err
	}
	defer f.Close()

	// Get memory bounds
	for _, p := range f.Progs {
		if p.Type != elf.PT_LOAD {
			continue
		}
		if p.Vaddr < startAddr {
			startAddr = p.Vaddr
		}
		if addr := p.Vaddr + p.Memsz; addr > endAddr {
			endAddr = addr
		}
	}
	curAddr = startAddr

	if pt, err = ForkExec(cmd, args); err != nil {
		log.Fatalln(err)
	}

	return nil
}

// guiLoop initializes de GUI and calls gocui's main loop.
func guiLoop() error {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		return err
	}
	defer g.Close()

	g.SetLayout(layout)
	if err := initKeybindings(g); err != nil {
		return err
	}

	err := g.MainLoop()
	if err != nil && err != gocui.Quit {
		return err
	}

	return nil
}

// layout defines the GUI's layout. In this case, it contains two views: the
// memory dump and a legend with information about keybindings.
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("legend", maxX-22, 0, maxX-1, 4); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		fmt.Fprintln(v, "KEYBINDINGS")
		fmt.Fprintln(v, "← ↑ → ↓: Seek memory")
		fmt.Fprintln(v, "^C: Exit")
	}

	if v, err := g.SetView("mem", 0, 0, maxX-23, maxY-1); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		if err := g.SetCurrentView("mem"); err != nil {
			return err
		}
		if err := seekMem(v, 0); err != nil {
			return err
		}
	}

	return nil
}

// initKeybindings configures the keybindings needed to visualize the process
// memory and close the application.
func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("mem", gocui.KeyArrowLeft, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			seekMem(v, -1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("mem", gocui.KeyArrowRight, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			seekMem(v, +1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("mem", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			seekMem(v, -16)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("mem", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			seekMem(v, +16)
			return nil
		}); err != nil {
		return err
	}

	return nil
}

// quit gets called when the user wants to close the application and breaks the
// gocui's main loop.
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.Quit
}

// seekMem allows to scroll the memory dump. delta is a displacement relative to
// the current memory address.
// START OMIT
func seekMem(v *gocui.View, delta int) error {
	addr := curAddr + uint64(delta)
	if addr < startAddr || addr > endAddr {
		return errors.New("address out of bounds")
	}

	v.Clear()
	fmt.Fprintf(v, "Dump from: 0x%x\n\n", addr)

	dumper := hex.Dumper(v)
	pt.Seek(addr)
	if _, err := io.CopyN(dumper, pt, blocksize); err != nil { // HL
		return err
	}
	curAddr = addr

	return nil
}

// STOP OMIT
