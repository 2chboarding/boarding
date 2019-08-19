package main

// ThreadStatus статус треда
type ThreadStatus int

//Статусы треда
const (
	Unknown ThreadStatus = iota
	Active
	Deleted
	Hidden
)

// PostID уникальный идентификатор поста
type PostID int64

// ThreadPosts хранит список постов в треде
type ThreadPosts []PostID

// ThreadsMap хранилище тредов
type ThreadsMap map[PostID]ThreadStruct

// PostsMap хранилище постов
type PostsMap map[PostID]PostStruct

// ThreadStruct хранит метаинформацию о треде и список постов
type ThreadStruct struct {
	Status  ThreadStatus
	Lasthit int64
	// список номеров постов, начиная с первого
	Posts ThreadPosts
}

// PostStruct хранит необходимую информацию о посте
type PostStruct struct {
	Subject   string
	Name      string
	Comment   string
	Timestamp int64
}

// BoardStruct кеширует треды с разбивкой по доскам
type BoardStruct struct {
	// Название доски
	Name string

	// Индекс тредов
	ThreadsIndex ThreadPosts

	// Threads хранит все треды, ключ номер первого поста
	Threads ThreadsMap
	// Все посты на этой доске, ключ - номер поста
	Posts PostsMap
}

// ImageBoard является корневым хранилищем
type ImageBoard struct {
	// хранит все доски по их ID
	Boards map[string]BoardStruct
	// Хранит категории
	Categories []string
	// Разбивка досок по категориям
	BoardsByCategory map[string][]string
}
