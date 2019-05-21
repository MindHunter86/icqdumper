package app

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type (
	appCui struct {
		cui *gocui.Gui
	}
)

func NewAppCui() *appCui {
	return &appCui{}
}

func (m *appCui) Bootstrap() (*appCui, error) {
	var e error
	if m.cui, e = gocui.NewGui(gocui.OutputNormal); e != nil {
		return nil, e
	}
	defer m.cui.Close()

	m.cui.SetManagerFunc(m.layout)

	if e = m.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, m.quit); e != nil {
		return nil, e
	}

	if e = m.cui.MainLoop(); e != nil && e != gocui.ErrQuit {
		return nil, e
	}

	return nil, e
}

func (m *appCui) layout(g *gocui.Gui) (e error) {
	maxX, maxY := g.Size()

	var view *gocui.View
	if view, e = g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		fmt.Println(view, "Hello World!")
	}

	return
}

func (m *appCui) quit(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }
