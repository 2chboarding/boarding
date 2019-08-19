package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func getJSONStub(filename string) ([]byte, error) {
	log.Printf("Loading %v", filename)

	return ioutil.ReadFile(filename)
}

// GetBoardsCatalog загружает данные с сайта
func GetBoardsCatalog() ([]byte, error) {
	uri := fmt.Sprintf("https://2ch.hk/makaba/mobile.fcgi?task=get_boards")
	return getJSON(uri)
}

// GetBoardsCatalogStub читает данные из файла
func GetBoardsCatalogStub() ([]byte, error) {
	uri := fmt.Sprintf("../../data/boards.json")
	return getJSONStub(uri)
}

// GetThreads load json from given boards containing list of threads (first page)
func GetThreads(boardID string) ([]byte, error) {
	url := fmt.Sprintf("https://2ch.hk/%v/index.json", boardID)
	//url := fmt.Sprintf("https://2ch.hk/%v/catalog.json", boardID)
	return getJSON(url)
}

// GetThreadsStub читает данные из файла
func GetThreadsStub(boardID string) ([]byte, error) {
	uri := "../../data/board_index.json"
	return getJSONStub(uri)
}

// GetThread получает полный тред с номером num с доски boardID
func GetThread(boardID string, num PostID) ([]byte, error) {
	uri := fmt.Sprintf("https://2ch.hk/%v/res/%v.json", boardID, num)
	return getJSON(uri)
}
