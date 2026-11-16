package todo

import "github.com/gin-gonic/gin"

func Router(group *gin.RouterGroup) {
	group.GET("", listHandler)
	group.POST("", createHandler)
	group.GET("/:id", getHandler)
	group.PATCH("/:id", editHandler)
	group.DELETE("/:id", deleteHandler)
	group.PUT("/:id/rank", rankHandler)
	group.DELETE("/:id/rank", unrankHandler)
	group.PUT("/:id/complete", completionHandler(true))
	group.DELETE("/:id/complete", completionHandler(false))
}
