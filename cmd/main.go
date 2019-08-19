package main

import (
	"fmt"

	"github.com/gdamore/tcell"

	"github.com/rivo/tview"
)

func main() {
	Display()
}

func loadBoardsList(lst *tview.TreeView, ib *ImageBoard) {
	ib.FetchCategories()

	root := tview.NewTreeNode("Доски")
	lst.SetRoot(root).SetCurrentNode(root).SetTopLevel(0)

	for _, bcat := range ib.Categories {
		cat := tview.NewTreeNode(bcat).SetExpanded(false)
		root.AddChild(cat)

		for _, b := range ib.BoardsByCategory[bcat] {
			brd := tview.NewTreeNode(fmt.Sprintf("/%v/", b))
			brd.SetReference(b)
			cat.AddChild(brd)
		}
	}

}

func loadThreadsList(boardID string, tl *tview.List, ib *ImageBoard) {
	ib.UpdateBoard(boardID)
	tl.Clear()

	for _, t := range ib.Boards[boardID].ThreadsIndex {

		tl.AddItem(ParseHTML(ib.Boards[boardID].Posts[t].Subject), "", 0, nil)
	}
}

// Display function run application and display interface
func Display() {

	// TUI
	app := tview.NewApplication()
	bs := tview.NewTreeView()
	tl := tview.NewList().ShowSecondaryText(false)
	tv := tview.NewTextView().SetWordWrap(true).SetRegions(true).SetDynamicColors(true)

	bs.SetBorder(true)
	tl.SetBorder(true)
	tv.SetBorder(true)

	flex := tview.NewFlex().AddItem(bs, 0, 2, true).AddItem(tl, 0, 5, false).AddItem(tv, 0, 7, false)
	app.SetRoot(flex, true)
	app.SetFocus(bs)

	ib := ImageBoard{}
	loadBoardsList(bs, &ib)

	var boardID string
	widgetFocus := 0
	widgets := []tview.Primitive{bs, tl, tv}

	bs.SetSelectedFunc(func(node *tview.TreeNode) {
		if node.GetReference() != nil {
			boardID = node.GetReference().(string)
			loadThreadsList(boardID, tl, &ib)
			app.SetFocus(tl)
			widgetFocus = 1
		} else {
			node.SetExpanded(!node.IsExpanded())
		}

	})

	tl.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if boardID != "" {
			thID := ib.Boards[boardID].ThreadsIndex[index]
			ib.UpdateThread(boardID, thID)
			t := ib.RenderThread(boardID, thID)
			_ = t
			tv.SetText(t)
			tv.ScrollToBeginning()
			app.SetFocus(tv)
			widgetFocus = 2

		}
	})

	tl.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if boardID != "" {
			thID := ib.Boards[boardID].ThreadsIndex[index]
			t := ib.RenderThread(boardID, thID)
			_ = t
			tv.SetText(t)
			tv.ScrollToBeginning()
		}
	})

	//panic(nil)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			widgetFocus--
			if widgetFocus < 0 {
				widgetFocus = 0
			}
			app.SetFocus(widgets[widgetFocus])
		case tcell.KeyRight:
			widgetFocus++
			if widgetFocus >= len(widgets) {
				widgetFocus = len(widgets) - 1
			}
			app.SetFocus(widgets[widgetFocus])
		default:
			return event
		}

		return nil
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
