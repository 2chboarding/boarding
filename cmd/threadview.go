package main

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// TextBlock keeps text block (word/token) with single style
type TextBlock struct {
	text  string
	width int
	style tcell.Style
	ref   int // url index
}

// TextLine keeps single text line with width not more that specified
type TextLine struct {
	blocks []TextBlock
	width  int
}

// Text keeps full text splitted by lines and blocks
type Text struct {
	width int
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

// SetText set new text to ThreadVew
func (tv *ThreadView) SetText(text string) {
	tv.text = text
	tv.lines = strings.Split(text, "\n")

	_, _, w, _ := tv.Box.GetInnerRect()
	tv.tlines = parseText2(text, w)
}

// ScrollToBeginning scroll ThreadView to first line
func (tv *ThreadView) ScrollToBeginning() {
	tv.vscroll = 0
}

type tagAttr struct {
	attrName  string
	attrValue string
}

// ParseHTML2 Parse HTML and emit tokens
func ParseHTML2(source string, eventFunc func(tt html.TokenType, token string, attrs []tagAttr)) {
	tokenizer := html.NewTokenizer(strings.NewReader(source))

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return
			}
			// unknown error
			panic(tokenizer.Err())

		case html.StartTagToken, html.EndTagToken:
			tagName, hasAttrs := tokenizer.TagName()
			tagNameStr := string(tagName)

			if tokenType == html.StartTagToken {
				var attrs []tagAttr

				if hasAttrs {
					for {
						key, value, more := tokenizer.TagAttr()

						attrs = append(attrs, tagAttr{string(key), string(value)})
						if !more {
							break
						}
					}
				}

				eventFunc(tokenType, tagNameStr, attrs)

			} else {
				// html.EndTagToken
				if hasAttrs {
					panic("EndTag has attributes")
				}
				//emitCloseTag(tagNameStr)
				eventFunc(tokenType, tagNameStr, nil)
			}

		case html.TextToken:
			eventFunc(tokenType, string(tokenizer.Text()), nil)
		}
	}
}

// NewTextParser parse HTML from source to Text
func (txt *Text) NewTextParser(source string) {
	var currLine TextLine
	var currBlock TextBlock

	flushLine := func() {
		txt.lines = append(txt.lines, currLine)
		currLine.blocks = nil
		currLine.width = 0
	}

	wrapText := func(text string) {
		var inText bool

		flushTextBlock := func() {
			if currLine.width+currBlock.width > txt.width {
				flushLine()
			}

			if currBlock.text != "" {
				// append current text block to current text line
				currLine.blocks = append(currLine.blocks, currBlock)
				currLine.width += currBlock.width
				// clear current text block
				currBlock.text = ""
				currBlock.width = 0
			}
		}

		for _, ch := range text {
			switch ch {
			case ' ':
				if inText {
					flushTextBlock()
				}
				inText = false

			default:
				if !inText {
					flushTextBlock()
				}
				inText = true
			}

			currBlock.text += string(ch)
			currBlock.width++
		}

		// flush last text block
		flushTextBlock()
	}

	eventFunc := func(tt html.TokenType, token string, attrs []tagAttr) {
		switch tt {
		case html.StartTagToken:
			switch token {
			case "br":
				flushLine()

			case "a":
			}

		case html.EndTagToken:

		case html.TextToken:
			wrapText(token)

		default:
			panic(fmt.Sprintf("Unknown TokenType %v", tt))
		}
	}

	ParseHTML2(source, eventFunc)
}

func parseText2(source string, width int) Text {
	var txt Text
	txt.width = width
	txt.NewTextParser(source)

	return txt
}

// Draw drawing content of ThreadView on screens
func (tv *ThreadView) Draw(screen tcell.Screen) {
	x, y, w, h := tv.Box.GetInnerRect()

	if tv.oldWidth != w {
		// reparse text
		tv.tlines = parseText2(tv.text, w)
		tv.color++
		tv.oldWidth = w
	}

	colors := []tcell.Color{tcell.ColorRed, tcell.ColorGreen, tcell.ColorBlue}
	tv.color %= len(colors)

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
						screen.SetContent(x+xx, y+yy, ch, nil, b.style)
						xx++
					}
				}
			}
		}

		// fill left space
		for lx := xx; lx < w; lx++ {
			screen.SetContent(x+lx, y+yy, ' ', nil, tcell.StyleDefault)
		}

	}
}

// InputHandler handle input
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

// RenderThread join all posts text to one big
func (ib *ImageBoard) RenderThread(boardID string, threadID PostID) string {

	var result string
	for _, postID := range ib.Boards[boardID].Threads[threadID].Posts {
		post := ib.Boards[boardID].Posts[postID]
		result += post.Name + "<br>"
		result += post.Comment + "<br><br>"
	}

	return result
}
