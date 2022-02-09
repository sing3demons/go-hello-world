package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/hello-world/model"
	"github.com/sing3demons/hello-world/repository"
)

type todoHandler struct {
	tx repository.TodoRepository
}

type TodoHandler interface {
	FindTodos(c *fiber.Ctx) error
	InsertTodos(c *fiber.Ctx) error
}

func NewTodoHandler(tx repository.TodoRepository) TodoHandler {
	return &todoHandler{tx: tx}
}

func (h *todoHandler) FindTodos(c *fiber.Ctx) error {
	todo, err := h.tx.FindAll()
	if err != nil {
		return c.JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(todo)
}

func (h *todoHandler) InsertTodos(c *fiber.Ctx) error {
	var todo model.Todo

	if err := c.BodyParser(&todo); err != nil {
		return c.JSON(err)
	}

	_, err := h.tx.Create(todo)
	if err != nil {
		return c.JSON(err)
	}
	return c.SendStatus(fiber.StatusCreated)
}
