package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// TextBlockStyle keeps style and other information
type TextBlockStyle struct {
	style tcell.Style
	tag   string
}

// TextBlock keeps text block (word/token) with single style
type TextBlock struct {
	text  string
	width int

	style TextBlockStyle
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
	links Links
}

func (txt *Text) Dump() {
	fl, err := os.Create("dump.txt")

	s := fmt.Sprint(txt)

	ioutil.WriteFile("dump1.txt", []byte(s), 0666)

	if err != nil {
		panic(err)
	}

	defer fl.Close()

	fl.WriteString(fmt.Sprintf("Text width %v\n", txt.width))

	for i, l := range txt.lines {
		fl.WriteString(
			fmt.Sprintf(" Line of text # %v, has %v blocks and width %v\n", i, len(l.blocks), l.width))

		for j, b := range l.blocks {
			fl.WriteString(
				fmt.Sprintf("  Block %v has width %v and style %x at %v. Text \"%v\"\n", j, b.width, b.style.style, b.style.tag, b.text))
		}
	}

	//ioutil.WriteFile("dump1.txt", txt.(string), 0666)
}

// Link keeps information about link
type Link struct {
	text   string
	url    string
	local  bool
	thread PostID
	post   PostID
}

// Links map of links
type Links map[int]Link

// ThreadView display thread
type ThreadView struct {
	*tview.Box
	text       string
	cachedText Text
	//Post     PostStruct
	vscroll  int
	oldWidth int
}

type PostView struct {
	tview.Box
	PostStruct        // original post
	postID     PostID //
	postText   Text   // Cached post text
}

type ThreadView2 struct {
	*tview.Box
	Posts []PostView
}

func (pv *PostView) Draw(screen tcell.Screen) {
	x, y, w, h := pv.Box.GetInnerRect()

	// if pv.oldWidth != w {
	// 	// reparse text
	// 	tv.UpdateCache(tv.text)
	// 	tv.oldWidth = w
	// }

	//	pv.Box.Draw(screen)

	for yy := 0; yy < h; yy++ {
		lnum := yy
		var tl *TextLine

		if lnum >= 0 && lnum < len(pv.postText.lines) {
			tl = &pv.postText.lines[lnum]
		}

		xx := 0

		// output text
		if tl != nil {
			for _, b := range tl.blocks {
				for _, ch := range b.text {

					if xx < w {
						if xx > w {
							panic("TextLine exceed widget width")
						}
						screen.SetContent(x+xx, y+yy, ch, nil, b.style.style)
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

// SetText set new text to ThreadVew
func (tv *ThreadView) SetText(text string) {
	_, _, w, _ := tv.Box.GetInnerRect()
	tv.text = text
	tv.cachedText.width = w
	tv.UpdateCache(text)

	tv.cachedText.Dump()

	//tv.tlines = parseText2(text, w)
}

// ScrollToBeginning scroll ThreadView to first line
func (tv *ThreadView) ScrollToBeginning() {
	tv.vscroll = 0
}

type tagAttr struct {
	attrName  string
	attrValue string
}

// ParseHTML Parse HTML and emit tokens
func ParseHTML(source string, eventFunc func(tt html.TokenType, token string, attrs []tagAttr)) {
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

	var style []TextBlockStyle

	currStyle := func() TextBlockStyle {
		return style[len(style)-1]
	}

	pushStyle := func(st TextBlockStyle) {
		style = append(style, st)
		//currBlock.style = st
	}

	popStyle := func() {
		if len(style) == 0 {
			panic("Style stack empty")
		}
		style = style[:len(style)-1]
		//currBlock.style = currStyle()
	}

	pushStyle(TextBlockStyle{tcell.StyleDefault, ""})

	flushTextLine := func() {
		txt.lines = append(txt.lines, currLine)
		currLine.blocks = nil
		currLine.width = 0
	}

	flushTextBlock := func() {
		if currLine.width+currBlock.width > txt.width {
			flushTextLine()
		}

		if currBlock.text != "" || currBlock.width != 0 {
			// append current text block to current text line
			currLine.blocks = append(currLine.blocks, currBlock)
			currLine.width += currBlock.width
		}

		// clear current text block
		currBlock.text = ""
		currBlock.width = 0
		currBlock.style = currStyle()
	}

	wrapText := func(text string) {
		var inText bool

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

		flushTextBlock()
	}

	eventFunc := func(tt html.TokenType, token string, attrs []tagAttr) {
		switch tt {
		case html.StartTagToken:
			switch token {
			case "br":
				flushTextBlock()
				flushTextLine()

			case "a":
				cs := currStyle()
				cs.tag = "a"
				cs.style = tcell.StyleDefault.Foreground(tcell.ColorLime)
				pushStyle(cs)
				flushTextBlock()

			case "strong":
				cs := currStyle()
				cs.tag = "strong"
				cs.style = cs.style.Bold(true)
				pushStyle(cs)
				flushTextBlock()

			default:
				flushTextBlock()
			}

		case html.EndTagToken:
			switch token {
			case "a":
				popStyle()
				flushTextBlock()

			case "strong":
				popStyle()
				flushTextBlock()

			default:
				flushTextBlock()
			}

		case html.TextToken:
			wrapText(token)

		default:
			panic(fmt.Sprintf("Unknown TokenType %v", tt))
		}
	}

	ParseHTML(source, eventFunc)

	// flush last text remains
	flushTextBlock()
	flushTextLine()

}

func (tv *ThreadView) UpdateCache(source string) {
	tv.cachedText.lines = nil
	tv.cachedText.NewTextParser(tv.text)
}

func (tv *ThreadView) SetPost(post *PostStruct) {
	//tv.Post = *post
}

// Draw drawing content of ThreadView on screens
func (tv *ThreadView) Draw(screen tcell.Screen) {
	x, y, w, h := tv.Box.GetInnerRect()

	if tv.oldWidth != w {
		// reparse text
		tv.UpdateCache(tv.text)
		tv.oldWidth = w
	}

	//
	for yy := 0; yy < h; yy++ {
		lnum := tv.vscroll + yy
		var tl *TextLine

		if lnum >= 0 && lnum < len(tv.cachedText.lines) {
			tl = &tv.cachedText.lines[lnum]
		}

		xx := 0

		// output text
		if tl != nil {
			for _, b := range tl.blocks {
				for _, ch := range b.text {

					if xx < w {
						if xx > w {
							panic("TextLine exceed widget width")
						}
						screen.SetContent(x+xx, y+yy, ch, nil, b.style.style)
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

	/*pv := &PostView{Box: tview.NewBox()}
	pv.SetRect(x+5, y+5, w-10, h-10)
	//pv.SetBorder(true)
	pv.post = tv.Post
	pv.postText.width = w - 10
	pv.postText.NewTextParser(pv.post.Comment)
	pv.SetRect(x+5, y+5, w-10, len(pv.postText.lines))
	pv.Draw(screen)*/
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
		result += post.Name + "<br><br>"
		result += post.Comment + "<br><br>"
	}

	return result
}
