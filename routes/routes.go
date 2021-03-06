package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/hello-world/cache"
	"github.com/sing3demons/hello-world/database"
	"github.com/sing3demons/hello-world/handler"
	"github.com/sing3demons/hello-world/repository"
)

func Serve(app *fiber.App) {
	db := database.GetDB()
	cache := cache.NewCacher(cache.NewCacherConfig())
	v1 := app.Group("/api")

	todoGroup := v1.Group("/todo")
	todoRepository := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepository, cache)
	{
		todoGroup.Get("", todoHandler.FindTodos)
		todoGroup.Post("", todoHandler.InsertTodos)
	}
}
