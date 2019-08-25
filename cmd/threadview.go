package main

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type TextBlock struct {
	text  string
	width int
}

type TextLine struct {
	blocks []TextBlock
}

type Text struct {
	lines []TextLine
}

// ThreadView display thread
type ThreadView struct {
	*tview.Box
	text     string
	lines    []string
	tlines   Text
	vscroll  int
	color    int
	oldWidth int
}

func (tv *ThreadView) SetText(text string) {
	tv.text = text
	tv.lines = strings.Split(text, "\n")

	_, _, w, _ := tv.Box.GetInnerRect()
	tv.tlines = parseText(text, w)
}

func (tv *ThreadView) ScrollToBeginning() {
	tv.vscroll = 0
}

func parseText(text string, width int) Text {
	var currToken string
	var widthToken int

	var inWord = false
	var tokens []TextBlock
	var t Text

	flushToken := func() {
		tokens = append(tokens, TextBlock{currToken, widthToken})
		currToken = ""
		widthToken = 0
	}

	// split input text to tokens (words)
	for _, rune := range text {

		switch rune {
		case '\n':
			// new line
			flushToken()
			currToken = "\n"
			widthToken = 0
			flushToken()

		case ' ', '\t':
			// whitespace
			if inWord {
				flushToken()
				inWord = false
			}
			currToken += string(rune)
			widthToken++

		default:
			// regular symbol
			if !inWord {
				flushToken()
				inWord = true
			}
			currToken += string(rune)
			widthToken++
		}
	}
	//ioutil.WriteFile("test.dump", []byte(tokens), 0666)

	// split tokens to lines
	var tl TextLine
	var tll int

	flushLine := func() {
		t.lines = append(t.lines, tl)
		tll = 0
		//tl.blocks = tl.blocks[0:0]
		tl.blocks = nil
	}

	for _, tk := range tokens {

		// unconditional flush line when \n occured
		if tk.text == "\n" {
			flushLine()
		} else {
			// if current line too long, flush and start new
			if tll+tk.width > width {
				flushLine()
			}

			// append current token to line and update length
			tl.blocks = append(tl.blocks, tk)
			tll += tk.width
		}
	}

	// flush remains
	flushLine()

	return t
}

func (tv *ThreadView) Draw(screen tcell.Screen) {
	x, y, w, h := tv.Box.GetInnerRect()

	if tv.oldWidth != w {
		// reparse text
		tv.tlines = parseText(tv.text, w)
		tv.color++
		tv.oldWidth = w
	}

	// Clear output area
	/*for xx := 0; xx < w; xx++ {
		for yy := 0; yy < h; yy++ {
			screen.SetContent(x+xx, y+yy, '.', nil, tcell.StyleDefault)
		}
	}*/

	colors := [...]tcell.Color{tcell.ColorRed, tcell.ColorGreen, tcell.ColorBlue}
	tv.color %= len(colors)

	//xoffs := 0
	//yoffs := 0
	/*for _, l := range tv.lines {
		for _, ch := range l {
			curry := y + yoffs - tv.vscroll

			if curry >= y && curry < y+h {
				screen.SetContent(x+xoffs, curry, ch, nil, tcell.StyleDefault.Foreground(colors[tv.color]))
			}
			xoffs++
			if xoffs >= w {
				xoffs = 0
				yoffs++
			}
		}

		yoffs++
		xoffs = 0
	}
	*/

	/*
		for _, l := range tv.tlines.lines {
			for _, b := range l.blocks {
				for _, ch := range b.text {
					curry := y + yoffs - tv.vscroll
					if curry >= y && curry < y+h {
						screen.SetContent(x+xoffs, curry, ch, nil, tcell.StyleDefault.Foreground(colors[tv.color]))
					}
					xoffs++
				}
			}
			yoffs++
			xoffs = 0
		}
	*/

	//
	for yy := 0; yy < h; yy++ {
		lnum := tv.vscroll + yy
		var tl *TextLine
		if lnum >= 0 && lnum < len(tv.tlines.lines) {
			tl = &tv.tlines.lines[lnum]
		}

		xx := 0

		// output text
		if tl != nil {
			for _, b := range tl.blocks {
				for _, ch := range b.text {

					if xx < w {
						screen.SetContent(x+xx-5, y+yy, ch, nil, tcell.StyleDefault)
						xx++
					}
				}
			}
		}

		// fill left space
		for lx := xx; lx < w; lx++ {
			screen.SetContent(x+lx, y+yy, '.', nil, tcell.StyleDefault)
		}

	}
}

func (tv *ThreadView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return tv.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		_, _, _, h := tv.GetInnerRect()

		switch key := event.Key(); key {
		case tcell.KeyDown:
			tv.vscroll++
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			tv.vscroll--
		case tcell.KeyPgDn:
			tv.vscroll += h
		case tcell.KeyPgUp:
			tv.vscroll -= h

		}
	})
}

/*
func (tv *ThreadView) SetRect(x, y, w, h int) {
	tv.Box.SetRect(x, y, w, h)
}
*/
