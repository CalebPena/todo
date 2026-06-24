package todo

import (
	"database/sql"
	"todo/auth"
)

type Todo struct {
	Id             int
	Name           string
	Description    string
	Ranked         bool
	ValueRank int
	TimeRank       int
	ParentId       int
	Children       []Todo
}

func addTodo(db *sql.DB, user auth.User, todo Todo) (int, error) {
	return 0, nil
}

func getTodo(db *sql.DB, user auth.User, todoId int) error {
	return nil
}

func todoDone(db *sql.DB, user auth.User, todoId int) error {
	return nil
}

func todoNotDone(db *sql.DB, user auth.User, todoId int) error {
	return nil

}

func rankTodo(db *sql.DB, user auth.User, todoId int) error {
	return nil
}

func unrankTodo(db *sql.DB, user auth.User, todoId int) error {
	return nil
}
