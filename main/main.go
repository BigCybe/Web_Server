package main

import (
	"github.com/foolin/echo-template"
	"github.com/labstack/echo"
)

func EchoRequest(){
	c := echo.New()

	group := c.Group("/group")
	{
		group.POST("/new", HandlePostNewGroupEcho)
		group.GET("/sort", HandleGetGroupEcho)
		group.GET("/:id", GetGroupById)
		group.GET("/child/:id", GetChildById)
		group.DELETE("/:id", DeleteById)
		group.PUT("/:id", PutGroupbyId)
	}

	task := c.Group("/task")
	{
		task.GET("/", GetTask)
		task.POST("/new", CreateNewTask)
		task.PUT("/:id", PutTaskbyId)
		task.GET("/group/:id", GetTasksbyType)
		task.PUT("/mod/:id", CompletingTask)
		task.GET("/stat/:period", Statistic)
	}

	users := c.Group("/user")
	{
		users.POST("/auth/new", HandlePostNewUserEcho)
		users.POST("/login", HandlePostTokenEcho)
	}

	c.Renderer = echotemplate.Default()

	pages := c.Group("")
	{
		pages.GET("/", MainPage)
		pages.GET("/token", TokenPage)
		pages.GET("/group", GroupPage)
		pages.GET("/task", TaskPage)
		pages.GET("/t", MainPageTest)
	}
	c.Start(":8080")
}

func main() {
	EchoRequest()
}
