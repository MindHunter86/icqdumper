package app

import (
	"fmt"
	"io"

	"github.com/jroimartin/gocui"
)

type (
	AppCui struct {
		cui     *gocui.Gui
		buffer1 io.Writer
	}
)

func NewAppCui() *AppCui {
	return &AppCui{}
}

func (m *AppCui) Bootstrap() (*AppCui, error) {
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

func (m *AppCui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if view, e := g.SetView("hello", 1, 1, maxX-1, maxY-1); e != nil {
		if e != gocui.ErrUnknownView {
			return e
		}

		m.buffer1 = view
		fmt.Fprintln(view, "Hello Worlddddddddddddddddd!")
		fmt.Fprintln(view, maxX, maxY)
	}

	return nil
}

func (m *AppCui) quit(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }
func (m *AppCui) GetBuffer() io.Writer                   { return m.buffer1 }
func (m *AppCui) Destroy()                               { m.cui.Close() }
