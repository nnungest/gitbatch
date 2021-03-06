package gui

import (
	"fmt"
	"strings"

	"github.com/isacikgoz/gitbatch/core/command"
	"github.com/isacikgoz/gitbatch/core/git"
	"github.com/jroimartin/gocui"
)

func (gui *Gui) initStashedView(r *git.Repository) error {
	v, err := gui.g.View(stashViewFeature.Name)
	if err != nil {
		return err
	}
	v.Clear()
	st := r.Stasheds
	for _, s := range st {
		fmt.Fprintf(v, " %d %s: %s\n", s.StashID, cyan.Sprint(s.BranchName), s.Description)
	}
	if len(st) > 0 {
		adjustAnchor(0, len(st), v)
	}
	return nil
}

// moves the cursor downwards for the main view and if it goes to bottom it
// prevents from going further
func (gui *Gui) stashCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		_, oy := v.Origin()
		ly := len(v.BufferLines()) - 1

		// if we are at the end we just return
		if cy+oy == ly-1 || ly < 0 {
			return nil
		}
		v.EditDelete(true)
		pos := cy + oy + 1
		adjustAnchor(pos, ly, v)
		gui.renderStashDiff(pos)
	}
	return nil
}

// moves the cursor upwards for the main view
func (gui *Gui) stashCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		ly := len(v.BufferLines()) - 1
		// if we are at the end we just return
		if ly < 0 {
			return nil
		}
		v.EditDelete(true)
		pos := cy + oy - 1
		adjustAnchor(pos, ly, v)
		if pos >= 0 {
			gui.renderStashDiff(cy + oy - 1)
		}
	}
	return nil
}

func (gui *Gui) renderStashDiff(ix int) error {
	r := gui.getSelectedRepository()
	st := r.Stasheds
	if len(st) <= 0 {
		return nil
	}
	v, err := gui.g.View(dynamicViewFeature.Name)
	if err != nil {
		return err
	}
	v.Title = string(StashDiffMode)
	if err := gui.updateDynamicKeybindings(); err != nil {
		return err
	}
	v.Clear()
	d, err := command.StashDiff(r, st[ix].StashID)
	if err != nil {
		return err
	}
	s := colorizeDiff(d)
	fmt.Fprintf(v, strings.Join(s, "\n"))
	return nil
}

func (gui *Gui) stashDiff(g *gocui.Gui, v *gocui.View) error {

	_, oy := v.Origin()
	_, cy := v.Cursor()

	return gui.renderStashDiff(oy + cy)
}

func (gui *Gui) stashPop(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	_, oy := v.Origin()
	_, cy := v.Cursor()
	if len(r.Stasheds) <= 0 {
		return nil
	}

	// since the pop is a func of stashed item, we need to refresh entity here
	// defer r.Refresh()
	stashedItem := r.Stasheds[oy+cy]
	output, err := stashedItem.Pop()
	if err != nil {
		if err = gui.openErrorView(g, output,
			"You should manually resolve this issue",
			stashViewFeature.Name); err != nil {
			return err
		}
	}
	r.Refresh()
	if err := gui.focusToRepository(g, v); err != nil {
		return err
	}
	if err := gui.initStashedView(r); err != nil {
		return err
	}
	return nil
}
