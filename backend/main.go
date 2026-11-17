package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"todo/auth"
	"todo/database"
	"todo/todo"
	"todo/util"
)

func init() {
	if !util.IsDebug {
		gin.SetMode(gin.ReleaseMode)
	}
}

var i int = 0

func Hello(c *gin.Context) {
	debugStr := "off"
	if util.IsDebug {
		debugStr = "on"
	}

	user := auth.GetUser(c)

	fmt.Fprintf(c.Writer, "Hello %d. Debug mode is %s. Your email is %s", i, debugStr, user.Email)
	i++
}

func main() {
	db, err := database.InitDb()

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(database.DbMiddleware(db))
	r.GET("/healthz", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.Status(503)
			return
		}
		c.Status(200)
	})

	auth.Router(r.Group("/auth"))
	r.Any("/mcp", auth.MCPAuthMiddleware, todo.MCPHandler)

	p := r.Group("/")
	p.Use(auth.AuthMiddleware)

	p.GET("/ping", Hello)
	auth.MCPTokenRoutes(p.Group("/mcp-tokens"))
	todo.Router(p.Group("/todos"))

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("BACKEND_PORT")
	}
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
