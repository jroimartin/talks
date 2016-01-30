package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/jroimartin/gocui"
)

const blocksize = 512

var mu sync.Mutex // protects view "net"

func main() {
	runtime.LockOSThread()

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: net addr:port")
		os.Exit(2)
	}
	addr := os.Args[1]

	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetLayout(layout)
	if err := initKeybindings(g); err != nil {
		log.Fatalln(err)
	}

	go listenAndDump(g, addr)

	err := g.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

// listenAndDump listens at addr and handles incoming connections. It is called
// as a goroutine.
func listenAndDump(g *gocui.Gui, addr string) error {
	var err error

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go g.Execute(func(g *gocui.Gui) error {
			return handleConn(g, conn)
		})
	}
	return nil
}

// layout defines the GUI's layout. In this case, it contains two views: the
// memory dump and a legend with information about keybindings.
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("legend", maxX-23, 0, maxX-1, 5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "KEYBINDINGS")
		fmt.Fprintln(v, "↑ ↓: Seek memory")
		fmt.Fprintln(v, "A: Enable autoscroll")
		fmt.Fprintln(v, "^C: Exit")
	}

	if v, err := g.SetView("net", 0, 0, maxX-24, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetCurrentView("net"); err != nil {
			return err
		}
		v.Wrap = true
		v.Autoscroll = true
	}

	return nil
}

// initKeybindings configures the keybindings needed to visualize the received
// data and close the application.
func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("net", 'a', gocui.ModNone, autoscroll); err != nil {
		return err
	}
	if err := g.SetKeybinding("net", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			scrollView(v, -1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("net", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			scrollView(v, 1)
			return nil
		}); err != nil {
		return err
	}
	return nil
}

// handleConn handles connections and writes received data to the view called
// "net". It is called as a goroutine.
func handleConn(g *gocui.Gui, conn net.Conn) error {
	v, err := g.View("net")
	if err != nil {
		return err
	}
	fmt.Fprintf(v, "Connection from %s:\n\n", conn.RemoteAddr())
	dumper := hex.Dumper(v)
	if _, err := io.Copy(dumper, conn); err != nil {
		return err
	}
	fmt.Fprintf(v, "\n\nEnd of connection.\n\n")
	return nil
}

// quit gets called when the user wants to close the application and breaks the
// gocui's main loop.
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// autoscroll enables autoscroll mode.
func autoscroll(g *gocui.Gui, v *gocui.View) error {
	v.Autoscroll = true
	return nil
}

// scrollView implements view's scroll.
func scrollView(v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}
