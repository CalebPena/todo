package todo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"todo/auth"
	"todo/database"
	"todo/requests"

	"github.com/gin-gonic/gin"
)

type editRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	ParentId    json.RawMessage `json:"parentId"`
	parent      *int
}

func bindEdit(c *gin.Context, create bool) (editRequest, bool) {
	var data editRequest
	err := c.ShouldBindJSON(&data)
	if err != nil {
		requests.InvalidJson(c)
		return data, false
	}
	if create && data.Name == nil {
		c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper("Name is required"))
		return data, false
	}
	if data.Name != nil {
		*data.Name = strings.TrimSpace(*data.Name)
		if *data.Name == "" || len(*data.Name) > 255 {
			c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper("Name must be between 1 and 255 bytes"))
			return data, false
		}
	}
	if data.ParentId != nil {
		err = json.Unmarshal(data.ParentId, &data.parent)
		if err != nil || (data.parent != nil && (*data.parent < 1 || *data.parent > 2147483647)) {
			c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper("Parent ID must be an integer between 1 and 2147483647 or null"))
			return data, false
		}
	}
	return data, true
}

func todoId(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 || id > 2147483647 {
		c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper("Invalid todo ID"))
		return 0, false
	}
	return id, true
}

func todoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, requests.ErrorResponseWrapper("Todo not found"))
	case errors.Is(err, errInvalidParent):
		c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper(err.Error()))
	case errors.Is(err, errInvalidRank):
		c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper(err.Error()))
	case errors.Is(err, errCompletedTodo):
		c.JSON(http.StatusBadRequest, requests.ErrorResponseWrapper(err.Error()))
	default:
		requests.InternalError(c, err)
	}
}

func listHandler(c *gin.Context) {
	todos, err := listTodos(database.GetDb(c), auth.GetUser(c))
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todos": todos}))
}

func getHandler(c *gin.Context) {
	id, ok := todoId(c)
	if !ok {
		return
	}
	todo, err := getTodo(database.GetDb(c), auth.GetUser(c), id)
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todo": todo}))
}

func createHandler(c *gin.Context) {
	data, ok := bindEdit(c, true)
	if !ok {
		return
	}
	todo := Todo{Name: *data.Name, ParentId: data.parent}
	if data.Description != nil {
		todo.Description = *data.Description
	}
	created, err := addTodo(database.GetDb(c), auth.GetUser(c), todo)
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, requests.ResponseWrapper(gin.H{"todo": created}))
}

func editHandler(c *gin.Context) {
	id, ok := todoId(c)
	if !ok {
		return
	}
	data, ok := bindEdit(c, false)
	if !ok {
		return
	}
	todo, err := editTodo(database.GetDb(c), auth.GetUser(c), id, data)
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todo": todo}))
}

func deleteHandler(c *gin.Context) {
	id, ok := todoId(c)
	if !ok {
		return
	}
	err := deleteTodo(database.GetDb(c), auth.GetUser(c), id)
	if err != nil {
		todoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func rankHandler(c *gin.Context) {
	var data struct {
		ValueRank     *int            `json:"valueRank"`
		TimeRank      *int            `json:"timeRank"`
		ValueBeforeId json.RawMessage `json:"valueBeforeId"`
		TimeBeforeId  json.RawMessage `json:"timeBeforeId"`
	}
	id, ok := todoId(c)
	if !ok {
		return
	}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		requests.InvalidJson(c)
		return
	}
	value, err := parsePlacement(data.ValueRank, data.ValueBeforeId)
	if err != nil {
		todoError(c, err)
		return
	}
	time, err := parsePlacement(data.TimeRank, data.TimeBeforeId)
	if err != nil {
		todoError(c, err)
		return
	}
	if value.Keep && time.Keep {
		todoError(c, errInvalidRank)
		return
	}
	todo, err := rankTodo(database.GetDb(c), auth.GetUser(c), id, value, time)
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todo": todo}))
}

func unrankHandler(c *gin.Context) {
	id, ok := todoId(c)
	if !ok {
		return
	}
	todo, err := unrankTodo(database.GetDb(c), auth.GetUser(c), id)
	if err != nil {
		todoError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todo": todo}))
}

func parsePlacement(position *int, before json.RawMessage) (rankPlacement, error) {
	placement := rankPlacement{Position: position}
	if position == nil && before == nil {
		placement.Keep = true
		return placement, nil
	}
	if position != nil && before != nil {
		return placement, errInvalidRank
	}
	if position != nil && (*position < 1 || *position > 2147483647) {
		return placement, errInvalidRank
	}
	if before != nil {
		err := json.Unmarshal(before, &placement.BeforeId)
		if err != nil || (placement.BeforeId != nil && (*placement.BeforeId < 1 || *placement.BeforeId > 2147483647)) {
			return placement, errInvalidRank
		}
	}
	return placement, nil
}

func completionHandler(completed bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := todoId(c)
		if !ok {
			return
		}
		todo, err := completeTodo(database.GetDb(c), auth.GetUser(c), id, completed)
		if err != nil {
			todoError(c, err)
			return
		}
		c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"todo": todo}))
	}
}
