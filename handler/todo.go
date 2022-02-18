package handler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
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

type createTodo struct {
	Name  string `form:"name" json:"name"`
	Image string `form:"image" json:"image"`
}

func (h *todoHandler) FindTodos(c *fiber.Ctx) error {
	todo, err := h.tx.FindAll()
	if err != nil {
		return c.JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(todo)
}

func (h *todoHandler) InsertTodos(c *fiber.Ctx) error {
	var form createTodo

	if err := c.BodyParser(&form); err != nil {
		return c.JSON(err)
	}

	image, err := uploadImage(c, "todo")
	if err != nil {

		return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
	}

	var todo model.Todo
	copier.Copy(&todo, &form)
	todo.Image = image

	_, err = h.tx.Create(todo)
	if err != nil {
		return c.JSON(err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func uploadImage(c *fiber.Ctx, name string) (string, error) {
	file, err := c.FormFile("image")
	if err != nil || file == nil {
		log.Println(err)
		return "", err
	}
	m := time.Now().UnixMilli()
	n := time.Now().Unix() + m
	s := strconv.FormatInt(n, 12)
	filename := "uploads/" + name + "/" + "images" + "/" + strings.Replace(s, "-", "", -1)
	os.MkdirAll("uploads/"+name+"/"+"images", 0755)
	// extract image extension from original file filename
	fileExt := strings.Split(file.Filename, ".")[1]
	// generate image from filename and extension
	image := fmt.Sprintf("%s.%s", filename, fileExt)

	if err := c.SaveFile(file, image); err != nil {
		return "", err
	}

	return image, nil
}
