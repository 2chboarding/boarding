package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// Описание получаемой структуры JSON, указаны только нужные поля
// _board хранит информацию о доске
type _board struct {
	Name string `json:"name"` // Название доски - Наука, Космос и тд
	ID   string `json:"id"`   // идентификатор доски - sci, spc и тд, без слешей
	//Category string `json:"category"` // Категория
}

// _boardsCatalog хранит список категорий и досок им соотвествующих, ключ - категория - Тематика, Игры и тд
type _boardsCatalog map[string][]_board // Boards by categories

// FetchCategories получает и парсит каталог досок, полученные данные заносятся в ImageBoard,
// при этом все имеющиеся данные будут удалены
func (ib *ImageBoard) FetchCategories() {
	data, err := GetBoardsCatalog()

	if err != nil {
		panic(err)
	}

	var bc _boardsCatalog
	if err := json.Unmarshal(data, &bc); err != nil {
		panic(err)
	}

	// Инициализация
	ib.Categories = make([]string, 0, len(bc))
	ib.BoardsByCategory = make(map[string][]string)
	ib.Boards = make(map[string]BoardStruct)

	for cat, boards := range bc {
		ib.Categories = append(ib.Categories, cat)
		ib.BoardsByCategory[cat] = make([]string, 0, len(boards))

		for _, br := range boards {
			var b BoardStruct

			b.Name = br.Name
			b.Posts = make(PostsMap)
			b.Threads = make(ThreadsMap)

			ib.Boards[br.ID] = b

			ib.BoardsByCategory[cat] = append(ib.BoardsByCategory[cat], br.ID)
		}
	}

	sort.Strings(ib.Categories)

	return // boardCatalog
}

// структура отдельного поста
type _post struct {
	Num       json.Number `json:"num"`
	Comment   string      `json:"comment"`
	Name      string      `json:"name"`
	Subject   string      `json:"subject"`
	Timestamp int64       `json:"timestamp"`
}

// структура треда
type _thread struct {
	Board     string `json:"Board"`
	PostCount int    `json:"posts_count"`
	Threads   []struct {
		Posts []_post `json:"posts"`
	} `json:"threads"`
}

// UpdateBoard Обновляет данные по указанной доске, пропавшие, удаленные, обновленный треды будут
// помечены соответствующим образом
// TODO обнаружение пропавших, новых и тп это в планах, пока просто загружаются новые данные
func (ib *ImageBoard) UpdateBoard(ID string) {
	if ib.Boards == nil {
		panic("ib.Boards uninitialized")
	}

	data, err := GetThreads(ID)
	if err != nil {
		panic(err)
	}

	var t _thread
	if ID == "fur" {
		var m interface{}
		json.Unmarshal(data, m)
		//panic("STOP")
	}
	if err := json.Unmarshal(data, &t); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			ioutil.WriteFile("invalid_json.json", data, 0666)
			errMsg := fmt.Sprintf("syntax error at byte offset %d", e.Offset)
			ioutil.WriteFile("invalid_json.txt", []byte(errMsg), 0666)

		}
		panic(err)
	}

	// номера тредов (первых постов)

	//ib.Boards[ID].Threads := make([]ThreadStruct, 0, len(t.Threads))
	tempThreadIndex := make(ThreadPosts, 0, len(t.Threads))

	for _, th := range t.Threads {
		//fmt.Println(th.Posts[0].Num)

		if len(th.Posts) == 0 {
			continue
		}

		thNum, err := th.Posts[0].Num.Int64()
		if err != nil {
			panic(err)
		}
		var tempThread ThreadStruct
		tempThread.Posts = make(ThreadPosts, 0, len(th.Posts))
		tempThreadIndex = append(tempThreadIndex, PostID(thNum))

		for _, ps := range th.Posts {
			num, err := ps.Num.Int64()
			if err != nil {
				panic(err)
			}

			tempThread.Posts = append(tempThread.Posts, PostID(num))
			ib.updatePost(ID, ps)
		}
		ib.Boards[ID].Threads[PostID(thNum)] = tempThread

	}
	tempBoard := ib.Boards[ID]
	tempBoard.ThreadsIndex = tempThreadIndex
	ib.Boards[ID] = tempBoard
}

// UpdateThread обновляет данные указанного треда
func (ib *ImageBoard) UpdateThread(ID string, num PostID) {
	data, err := GetThread(ID, num)

	if err != nil {
		panic(err)
	}

	var t _thread
	if err := json.Unmarshal(data, &t); err != nil {
		panic(err)
	}

	thNum, err := t.Threads[0].Posts[0].Num.Int64()
	if err != nil {
		panic(err)
	}

	var tempThread ThreadStruct
	tempThread.Posts = make(ThreadPosts, 0, len(t.Threads[0].Posts))

	for _, ps := range t.Threads[0].Posts {
		num, err := ps.Num.Int64()
		if err != nil {
			panic(err)
		}

		tempThread.Posts = append(tempThread.Posts, PostID(num))
		ib.updatePost(ID, ps)
	}
	ib.Boards[ID].Threads[PostID(thNum)] = tempThread
}

func (ib *ImageBoard) updatePost(ID string, p _post) {
	num, err := p.Num.Int64()
	if err != nil {
		panic(err)
	}

	if _, ok := ib.Boards[ID].Posts[PostID(num)]; ok {
		//fmt.Printf("Post %v exists\n", num)
	}

	ib.Boards[ID].Posts[PostID(num)] = PostStruct{
		Subject:   p.Subject,
		Name:      p.Name,
		Comment:   p.Comment,
		Timestamp: p.Timestamp,
	}
}

// ParseHTML преобразует html теги в форматирование tview.TextView
func ParseHTML(source string) string {
	result := ""
	t := html.NewTokenizer(strings.NewReader(source))

	for {
		tt := t.Next()

		switch tt {
		case html.ErrorToken:
			err := t.Err()
			if err == io.EOF {
				return result
			}

			panic(err)

		case html.StartTagToken:
			tn, hasAttr := t.TagName()
			tok := string(tn)
			_ = hasAttr
			switch tok {
			case "br":
				result += "\n"

			case "a":
				result += `[#00ff00]["1"]`

			case "strong":
				result += `[::b]`

			default:
				/*result += "<" + tok

				if hasAttr {
					for {
						key, value, more := t.TagAttr()
						_ = key
						_ = value
						_ = more
						result += " " + string(key) + ":" + string(value)
						if !more {
							break
						}
					}
				}
				result += ">"*/
			}

		case html.EndTagToken:
			tn, _ := t.TagName()
			tok := string(tn)

			switch tok {
			case "a":
				result += `[""][white]`

			case "strong":
				result += `[::-]`

			default:
				/*result += "</" + tok + ">"*/
			}

		case html.TextToken:
			result += string(t.Text())
		}
	}

	//return result
}

func renderPost(post *PostStruct) string {
	s := ParseHTML(post.Comment)
	//fmt.Println(s)
	return s
}

// RenderThread рендерит указанный тред в виде строки для отображения
func (ib *ImageBoard) RenderThread(boardID string, threadID PostID) string {
	var result string
	for _, postID := range ib.Boards[boardID].Threads[threadID].Posts {
		post := ib.Boards[boardID].Posts[postID]
		result += "[yellow]" + ParseHTML(post.Name) + "[white]" + "\n"
		result += ParseHTML(post.Comment) + "\n\n"
	}

	return result
}
