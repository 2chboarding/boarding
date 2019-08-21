package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func getJSON(url string) ([]byte, error) {
	//log.Printf("Getting %v ...", url)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	//defer log.Printf("Done\n")

	// При ошибке вернем пустой JSON
	// TODO сделать нормальную обработку ошибок
	if resp.StatusCode != 200 {
		return []byte(`{}`), nil
	}

	return ioutil.ReadAll(resp.Body)
}

func getJSONStub(url string) ([]byte, error) {
	//log.Printf("Fake loading %v", url)

	var filename string

	if strings.Contains(url, "get_board") {
		filename = "data/boards.json"
	} else if strings.Contains(url, "index.json") {
		filename = "data/board_index.json"
	} else if strings.Contains(url, "/res/") {
		filename = "data/full_thread.json"
	}

	if filename == "" {
		return []byte{}, errors.New("Invalid filename")
	}
	return ioutil.ReadFile(filename)
}

//var getterFunc = getJSONStub
var getterFunc = getJSON

// GetBoardsCatalog загружает данные с сайта
func GetBoardsCatalog() ([]byte, error) {
	uri := fmt.Sprintf("https://2ch.hk/makaba/mobile.fcgi?task=get_boards")
	return getterFunc(uri)
}

// GetThreads load json from given boards containing list of threads (first page)
func GetThreads(boardID string) ([]byte, error) {
	url := fmt.Sprintf("https://2ch.hk/%v/index.json", boardID)
	//url := fmt.Sprintf("https://2ch.hk/%v/catalog.json", boardID)
	return getterFunc(url)
}

// GetThread получает полный тред с номером num с доски boardID
func GetThread(boardID string, num PostID) ([]byte, error) {
	uri := fmt.Sprintf("https://2ch.hk/%v/res/%v.json", boardID, num)
	return getterFunc(uri)
}
