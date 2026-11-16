package todo

import (
	"database/sql"
	"errors"
	"todo/auth"

	"github.com/lib/pq"
)

var errInvalidParent = errors.New("Parent must be one of your todos and cannot create a cycle")
var errInvalidRank = errors.New("Rank must be a valid sibling position, and comparison targets must be other ranked siblings")
var errCompletedTodo = errors.New("Reopen the todo before ranking it")

type Todo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Ranked      bool   `json:"ranked"`
	ValueRank   *int   `json:"valueRank"`
	TimeRank    *int   `json:"timeRank"`
	ParentId    *int   `json:"parentId"`
	Children    []Todo `json:"children"`
}

const todoColumns = "id, name, description, value_rank, time_rank, parent_id, completed"
const todoOrder = " ORDER BY value_rank::bigint + time_rank ASC NULLS LAST, id"

type scanner interface {
	Scan(...any) error
}

func scanTodo(row scanner) (Todo, error) {
	todo := Todo{Children: []Todo{}}
	err := row.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.ValueRank, &todo.TimeRank, &todo.ParentId, &todo.Completed)
	todo.Ranked = todo.ValueRank != nil && todo.TimeRank != nil
	return todo, err
}

func readTodos(rows *sql.Rows) ([]Todo, error) {
	defer rows.Close()
	todos := []Todo{}
	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func listTodos(db *sql.DB, user auth.User) ([]Todo, error) {
	rows, err := db.Query("SELECT "+todoColumns+" FROM todos WHERE user_id = $1"+todoOrder, user.Id)
	if err != nil {
		return nil, err
	}
	return readTodos(rows)
}

func getTodo(db *sql.DB, user auth.User, id int) (Todo, error) {
	todo, err := scanTodo(db.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	rows, err := db.Query("SELECT "+todoColumns+" FROM todos WHERE parent_id = $1 AND user_id = $2"+todoOrder, id, user.Id)
	if err != nil {
		return Todo{}, err
	}
	todo.Children, err = readTodos(rows)
	return todo, err
}

// Serialize hierarchy and ranking edits per user to prevent cycles and keep ranks consistent.
func beginEdit(db *sql.DB, user auth.User) (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	var id int
	err = tx.QueryRow("SELECT id FROM users WHERE id = $1 FOR UPDATE", user.Id).Scan(&id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}

func validateParent(tx *sql.Tx, user auth.User, id int, parentId *int) error {
	if parentId == nil {
		return nil
	}
	var valid bool
	err := tx.QueryRow(`WITH RECURSIVE ancestors AS (
		SELECT id, parent_id FROM todos WHERE id = $1 AND user_id = $2
		UNION
		SELECT t.id, t.parent_id FROM todos t JOIN ancestors a ON t.id = a.parent_id WHERE t.user_id = $2
	)
	SELECT EXISTS (SELECT 1 FROM ancestors) AND NOT EXISTS (SELECT 1 FROM ancestors WHERE id = $3)`, *parentId, user.Id, id).Scan(&valid)
	if err != nil {
		return err
	}
	if !valid {
		return errInvalidParent
	}
	return nil
}

func addTodo(db *sql.DB, user auth.User, todo Todo) (Todo, error) {
	tx, err := beginEdit(db, user)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()
	err = validateParent(tx, user, 0, todo.ParentId)
	if err != nil {
		return Todo{}, err
	}
	created, err := scanTodo(tx.QueryRow(`INSERT INTO todos (user_id, name, description, parent_id)
		VALUES ($1, $2, $3, $4) RETURNING `+todoColumns, user.Id, todo.Name, todo.Description, todo.ParentId))
	if err != nil {
		return Todo{}, err
	}
	err = compactRanks(tx, user, created.ParentId)
	if err != nil {
		return Todo{}, err
	}
	created, err = scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", created.Id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	return created, tx.Commit()
}

func editTodo(db *sql.DB, user auth.User, id int, data editRequest) (Todo, error) {
	tx, err := beginEdit(db, user)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()
	todo, err := scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	oldParentId := todo.ParentId
	if data.Name != nil {
		todo.Name = *data.Name
	}
	if data.Description != nil {
		todo.Description = *data.Description
	}
	if data.ParentId != nil {
		todo.ParentId = data.parent
		err = validateParent(tx, user, id, todo.ParentId)
		if err != nil {
			return Todo{}, err
		}
	}
	todo, err = scanTodo(tx.QueryRow(`UPDATE todos SET name = $1, description = $2, parent_id = $3,
		value_rank = CASE WHEN parent_id IS DISTINCT FROM $3 THEN NULL ELSE value_rank END,
		time_rank = CASE WHEN parent_id IS DISTINCT FROM $3 THEN NULL ELSE time_rank END
		WHERE id = $4 AND user_id = $5 RETURNING `+todoColumns, todo.Name, todo.Description, todo.ParentId, id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	err = compactRanks(tx, user, oldParentId)
	if err != nil {
		return Todo{}, err
	}
	parentChanged := (oldParentId == nil) != (todo.ParentId == nil) ||
		(oldParentId != nil && todo.ParentId != nil && *oldParentId != *todo.ParentId)
	if parentChanged {
		err = compactRanks(tx, user, todo.ParentId)
		if err != nil {
			return Todo{}, err
		}
	}
	todo, err = scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	return todo, tx.Commit()
}

func deleteTodo(db *sql.DB, user auth.User, id int) error {
	tx, err := beginEdit(db, user)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var parentId *int
	err = tx.QueryRow("DELETE FROM todos WHERE id = $1 AND user_id = $2 RETURNING parent_id", id, user.Id).Scan(&parentId)
	if err != nil {
		return err
	}
	err = compactRanks(tx, user, parentId)
	if err != nil {
		return err
	}
	return tx.Commit()
}

type rankPlacement struct {
	Position *int
	BeforeId *int
	Keep     bool
}

// A nil position and nil comparison target mean append to the ranked siblings.
func placeRank(tx *sql.Tx, user auth.User, todo Todo, column string, placement rankPlacement) (int, error) {
	if placement.Keep && todo.Ranked {
		if column == "value_rank" {
			return *todo.ValueRank, nil
		}
		return *todo.TimeRank, nil
	}
	rows, err := tx.Query(`SELECT id FROM todos WHERE user_id = $1
		AND parent_id IS NOT DISTINCT FROM $2 AND NOT completed AND value_rank IS NOT NULL AND id <> $3
		ORDER BY `+column+`, id`, user.Id, todo.ParentId, todo.Id)
	if err != nil {
		return 0, err
	}
	ids := []int{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	position := len(ids) + 1
	if placement.Position != nil {
		position = *placement.Position
	}
	if placement.BeforeId != nil {
		position = 0
		for index, id := range ids {
			if id == *placement.BeforeId {
				position = index + 1
				break
			}
		}
	}
	if position < 1 || position > len(ids)+1 {
		return 0, errInvalidRank
	}
	ids = append(ids, 0)
	copy(ids[position:], ids[position-1:])
	ids[position-1] = todo.Id
	// Only callers supply column names; never interpolate request values into SQL.
	_, err = tx.Exec(`UPDATE todos SET `+column+` = ordering.position
		FROM unnest($1::integer[]) WITH ORDINALITY AS ordering(id, position)
		WHERE todos.id = ordering.id AND todos.id <> $2 AND todos.user_id = $3`, pq.Array(ids), todo.Id, user.Id)
	return position, err
}

func compactRanks(tx *sql.Tx, user auth.User, parentId *int) error {
	_, err := tx.Exec(`WITH siblings AS (
		SELECT id, value_rank, time_rank, count(*) OVER () AS sibling_count
		FROM todos WHERE user_id = $1 AND parent_id IS NOT DISTINCT FROM $2 AND NOT completed
	), positions AS (
		SELECT id,
			row_number() OVER (ORDER BY value_rank, id) AS value_position,
			row_number() OVER (ORDER BY time_rank, id) AS time_position
		FROM siblings WHERE value_rank IS NOT NULL OR sibling_count = 1
	)
	UPDATE todos SET value_rank = positions.value_position, time_rank = positions.time_position
	FROM positions WHERE todos.id = positions.id`, user.Id, parentId)
	return err
}

func rankTodo(db *sql.DB, user auth.User, id int, value, time rankPlacement) (Todo, error) {
	tx, err := beginEdit(db, user)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()
	todo, err := scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	if todo.Completed {
		return Todo{}, errCompletedTodo
	}
	valueRank, err := placeRank(tx, user, todo, "value_rank", value)
	if err != nil {
		return Todo{}, err
	}
	timeRank, err := placeRank(tx, user, todo, "time_rank", time)
	if err != nil {
		return Todo{}, err
	}
	todo, err = scanTodo(tx.QueryRow(`UPDATE todos SET value_rank = $1, time_rank = $2
		WHERE id = $3 AND user_id = $4 RETURNING `+todoColumns, valueRank, timeRank, id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	return todo, tx.Commit()
}

func unrankTodo(db *sql.DB, user auth.User, id int) (Todo, error) {
	tx, err := beginEdit(db, user)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()
	todo, err := scanTodo(tx.QueryRow(`UPDATE todos SET value_rank = NULL, time_rank = NULL
		WHERE id = $1 AND user_id = $2 RETURNING `+todoColumns, id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	err = compactRanks(tx, user, todo.ParentId)
	if err != nil {
		return Todo{}, err
	}
	todo, err = scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	return todo, tx.Commit()
}

func completeTodo(db *sql.DB, user auth.User, id int, completed bool) (Todo, error) {
	tx, err := beginEdit(db, user)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()
	todo, err := scanTodo(tx.QueryRow(`UPDATE todos SET completed = $1,
		value_rank = CASE WHEN $1 THEN NULL ELSE value_rank END,
		time_rank = CASE WHEN $1 THEN NULL ELSE time_rank END
		WHERE id = $2 AND user_id = $3 RETURNING `+todoColumns, completed, id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	err = compactRanks(tx, user, todo.ParentId)
	if err != nil {
		return Todo{}, err
	}
	todo, err = scanTodo(tx.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = $1 AND user_id = $2", id, user.Id))
	if err != nil {
		return Todo{}, err
	}
	return todo, tx.Commit()
}
