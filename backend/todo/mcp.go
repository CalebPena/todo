package todo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"todo/auth"
	"todo/database"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type taskIDInput struct {
	ID int `json:"task_id" jsonschema:"Task ID"`
}

type mcpTaskSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Ranked      bool   `json:"ranked"`
	ValueRank   *int   `json:"valueRank"`
	TimeRank    *int   `json:"timeRank"`
	ParentID    *int   `json:"parentId"`
}
type mcpTask struct {
	mcpTaskSummary
	Children []mcpTaskSummary `json:"children"`
}
type taskOutput struct {
	Task mcpTask `json:"task"`
}
type tasksOutput struct {
	Tasks []mcpTaskSummary `json:"tasks"`
}

func summarizeMCPTask(task Todo) mcpTaskSummary {
	return mcpTaskSummary{task.Id, task.Name, task.Description, task.Completed, task.Ranked, task.ValueRank, task.TimeRank, task.ParentId}
}

type listTasksInput struct {
	ParentID  *int   `json:"parent_id,omitempty" jsonschema:"Parent task ID. Omit for root tasks."`
	All       bool   `json:"all,omitempty" jsonschema:"List all levels instead of just one parent."`
	Completed *bool  `json:"completed,omitempty"`
	Ranked    *bool  `json:"ranked,omitempty"`
	Sort      string `json:"sort,omitempty" jsonschema:"mix (default), value, or time"`
}
type createTaskInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ParentID    *int   `json:"parent_id,omitempty"`
}
type updateTaskInput struct {
	ID          int     `json:"task_id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentID    *int    `json:"parent_id,omitempty"`
	MoveToRoot  bool    `json:"move_to_root,omitempty" jsonschema:"Move to root instead of supplying parent_id."`
}
type completeTaskInput struct {
	ID        int  `json:"task_id"`
	Completed bool `json:"completed"`
}
type rankTaskInput struct {
	ID            int  `json:"task_id"`
	ValueBeforeID *int `json:"value_before_id" jsonschema:"Insert before this ranked sibling for value; null appends as least valuable."`
	TimeBeforeID  *int `json:"time_before_id" jsonschema:"Insert before this ranked sibling for time; null appends as most time-consuming."`
}

func validMCPID(id int) bool      { return id > 0 && id <= 2147483647 }
func validMCPParent(id *int) bool { return id == nil || validMCPID(*id) }

func mcpTaskResult(task Todo, err error) (*mcp.CallToolResult, taskOutput, error) {
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err = errors.New("Task not found")
		case errors.Is(err, errInvalidParent), errors.Is(err, errInvalidRank), errors.Is(err, errCompletedTodo):
		default:
			log.Printf("MCP task operation: %v", err)
			err = errors.New("Unable to update or read task")
		}
	}
	children := []mcpTaskSummary{}
	for _, child := range task.Children {
		children = append(children, summarizeMCPTask(child))
	}
	return nil, taskOutput{Task: mcpTask{summarizeMCPTask(task), children}}, err
}

// Stateless requests are authenticated independently, so revocation takes effect on the next request.
func MCPHandler(c *gin.Context) {
	if origin := c.GetHeader("Origin"); origin != "" {
		parsed, err := url.Parse(origin)
		if err != nil || !strings.EqualFold(parsed.Host, c.Request.Host) || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	db, user := database.GetDb(c), auth.GetUser(c)
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return taskMCPServer(db, user) }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	handler.ServeHTTP(c.Writer, c.Request)
}

func taskMCPServer(db *sql.DB, user auth.User) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "Todo", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: "Manage only the authenticated user's tasks. Ranks are relative to active ranked siblings: value 1 is most valuable; time 1 is quickest. Mix sorts by the sum of both ranks. Completing clears ranks. Use get_task/list_tasks to inspect siblings before rank_task. Ask before deleting tasks.",
	})
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}
	mcp.AddTool(server, &mcp.Tool{Name: "list_tasks", Description: "List tasks under a parent, or all tasks, with completion/ranking filters and mix, value, or time ordering.", Annotations: readOnly}, func(_ context.Context, _ *mcp.CallToolRequest, input listTasksInput) (*mcp.CallToolResult, tasksOutput, error) {
		if !validMCPParent(input.ParentID) {
			return nil, tasksOutput{}, errors.New("Invalid parent_id")
		}
		if input.Sort != "" && input.Sort != "mix" && input.Sort != "value" && input.Sort != "time" {
			return nil, tasksOutput{}, errors.New("sort must be mix, value, or time")
		}
		tasks, err := listTodos(db, user)
		if err != nil {
			log.Printf("MCP list tasks: %v", err)
			return nil, tasksOutput{}, errors.New("Unable to list tasks")
		}
		filtered := []Todo{}
		for _, task := range tasks {
			sameParent := task.ParentId == nil && input.ParentID == nil || task.ParentId != nil && input.ParentID != nil && *task.ParentId == *input.ParentID
			if (!input.All && !sameParent) || (input.Completed != nil && task.Completed != *input.Completed) || (input.Ranked != nil && task.Ranked != *input.Ranked) {
				continue
			}
			filtered = append(filtered, task)
		}
		score := func(task Todo) int {
			if !task.Ranked {
				return 1 << 62
			}
			if input.Sort == "time" {
				return *task.TimeRank
			}
			if input.Sort == "value" {
				return *task.ValueRank
			}
			return *task.ValueRank + *task.TimeRank
		}
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].Completed != filtered[j].Completed {
				return !filtered[i].Completed
			}
			return score(filtered[i]) < score(filtered[j])
		})
		summaries := []mcpTaskSummary{}
		for _, task := range filtered {
			summaries = append(summaries, summarizeMCPTask(task))
		}
		return nil, tasksOutput{Tasks: summaries}, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "get_task", Description: "Read a task including its description, ranks, and direct subtasks.", Annotations: readOnly}, func(_ context.Context, _ *mcp.CallToolRequest, input taskIDInput) (*mcp.CallToolResult, taskOutput, error) {
		if !validMCPID(input.ID) {
			return nil, taskOutput{}, errors.New("Invalid task_id")
		}
		return mcpTaskResult(getTodo(db, user, input.ID))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "create_task", Description: "Create a task or subtask. A lone active sibling is automatically ranked first."}, func(_ context.Context, _ *mcp.CallToolRequest, input createTaskInput) (*mcp.CallToolResult, taskOutput, error) {
		input.Name = strings.TrimSpace(input.Name)
		if input.Name == "" || len(input.Name) > 255 || !validMCPParent(input.ParentID) {
			return nil, taskOutput{}, errors.New("Provide a name of 1–255 bytes and a valid parent_id")
		}
		return mcpTaskResult(addTodo(db, user, Todo{Name: input.Name, Description: input.Description, ParentId: input.ParentID}))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "update_task", Description: "Edit a task's name, description, or parent. Moving parents clears its ranks. Omitted fields stay unchanged."}, func(_ context.Context, _ *mcp.CallToolRequest, input updateTaskInput) (*mcp.CallToolResult, taskOutput, error) {
		if !validMCPID(input.ID) || !validMCPParent(input.ParentID) || input.MoveToRoot && input.ParentID != nil {
			return nil, taskOutput{}, errors.New("Invalid task_id or parent placement")
		}
		if input.Name != nil {
			*input.Name = strings.TrimSpace(*input.Name)
			if *input.Name == "" || len(*input.Name) > 255 {
				return nil, taskOutput{}, errors.New("Name must be 1–255 bytes")
			}
		}
		data := editRequest{Name: input.Name, Description: input.Description, parent: input.ParentID}
		if input.ParentID != nil || input.MoveToRoot {
			data.ParentId, _ = json.Marshal(input.ParentID)
		}
		return mcpTaskResult(editTodo(db, user, input.ID, data))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "set_task_completed", Description: "Complete or reopen a task. Completing clears both ranks; reopening does not restore previous rankings."}, func(_ context.Context, _ *mcp.CallToolRequest, input completeTaskInput) (*mcp.CallToolResult, taskOutput, error) {
		if !validMCPID(input.ID) {
			return nil, taskOutput{}, errors.New("Invalid task_id")
		}
		return mcpTaskResult(completeTodo(db, user, input.ID, input.Completed))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "rank_task", Description: "Rank or re-rank an active task in both sibling orderings. Supply both before IDs; null means append. Existing sibling ranks shift without gaps."}, func(_ context.Context, _ *mcp.CallToolRequest, input rankTaskInput) (*mcp.CallToolResult, taskOutput, error) {
		if !validMCPID(input.ID) || !validMCPParent(input.ValueBeforeID) || !validMCPParent(input.TimeBeforeID) {
			return nil, taskOutput{}, errors.New("Invalid task or comparison ID")
		}
		return mcpTaskResult(rankTodo(db, user, input.ID, rankPlacement{BeforeId: input.ValueBeforeID}, rankPlacement{BeforeId: input.TimeBeforeID}))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "unrank_task", Description: "Clear a task's ranks and close gaps. A lone active task remains automatically ranked first."}, func(_ context.Context, _ *mcp.CallToolRequest, input taskIDInput) (*mcp.CallToolResult, taskOutput, error) {
		if !validMCPID(input.ID) {
			return nil, taskOutput{}, errors.New("Invalid task_id")
		}
		return mcpTaskResult(unrankTodo(db, user, input.ID))
	})
	mcp.AddTool(server, &mcp.Tool{Name: "delete_task", Description: "Permanently delete a task and its descendant subtasks. Ask for confirmation first."}, func(_ context.Context, _ *mcp.CallToolRequest, input taskIDInput) (*mcp.CallToolResult, map[string]bool, error) {
		if !validMCPID(input.ID) {
			return nil, nil, errors.New("Invalid task_id")
		}
		err := deleteTodo(db, user, input.ID)
		_, _, err = mcpTaskResult(Todo{}, err)
		return nil, map[string]bool{"deleted": err == nil}, err
	})
	return server
}
