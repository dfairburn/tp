package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, _ := g.Size()
	v1, err := g.SetView("side", -1, 0, int(0.2*float32(maxX)), 5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v1, "Hello world 1")
	}

	v2, err := g.SetView("main", int(0.2*float32(maxX)), 0, maxX, 5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v2, "Hello world 2")
	}

	v3, err := g.SetView("cmdline", -1, 0, maxX, 5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v3, "Hello world 3")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
