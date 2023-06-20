package routes

import (
	"user-service/cmd/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {
	router.POST("/users", controllers.CreateUser())
	router.GET("/users/:userId", controllers.FindById())
	router.GET("/users", controllers.GetUsers())
}
