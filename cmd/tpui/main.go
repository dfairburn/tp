package main

import (
	"errors"
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
	maxX, maxY := g.Size()

	if err := createView(g, "URL", 0, 0, maxX-1, 2); err != nil {
		return err
	}
	if err := createView(g, "Variable Overrides", 0, 3, int(0.4*float32(maxX)), int(0.7*float32(maxY))); err != nil {
		return err
	}
	if err := createView(g, "Templates", 0, int(0.7*float32(maxY))+1, int(0.4*float32(maxX)), maxY-1); err != nil {
		return err
	}
	if err := createView(g, "Body", int(0.4*float32(maxX))+2, 3, maxX-1, int(0.5*float32(maxY))); err != nil {
		return err
	}
	if err := createView(g, "Response", int(0.4*float32(maxX))+2, int(0.5*float32(maxY))+1, maxX-1, maxY-1); err != nil {
		return err
	}
	return nil
}

func createView(g *gocui.Gui, title string, x, y, width, height int) error {
	v, err := g.SetView(title, x, y, width, height)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = title
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
