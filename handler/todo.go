package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"github.com/sing3demons/hello-world/cache"
	"github.com/sing3demons/hello-world/model"
	"github.com/sing3demons/hello-world/repository"
)

type todoHandler struct {
	tx    repository.TodoRepository
	cache *cache.Cacher
}

type Pagination struct {
	Limit      int         `json:"limit,omitempty"`
	Page       int         `json:"page,omitempty"`
	Sort       string      `json:"sort,omitempty"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

type TodoHandler interface {
	FindTodos(c *fiber.Ctx) error
	InsertTodos(c *fiber.Ctx) error
}

func NewTodoHandler(tx repository.TodoRepository, cache *cache.Cacher) TodoHandler {
	return &todoHandler{tx: tx, cache: cache}
}

type createTodo struct {
	Name  string `form:"name" json:"name"`
	Image string `form:"image" json:"image"`
}

func (h *todoHandler) FindTodos(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "24"))
	page, _ := strconv.Atoi(c.Query("page", "1"))

	str := fmt.Sprintf("::%d::%d", limit, page)
	query1CacheKey := "todo::all" + str
	query2CacheKey := "todo::page" + str

	cacheItems, err := h.cache.MGet([]string{query1CacheKey, query2CacheKey})
	if err != nil {
		return c.JSON(err)
	}

	todoJS := cacheItems[0]
	pageJS := cacheItems[1]

	var todos []model.Todo
	var paging *model.Pagination

	if todoJS != nil && len(todoJS.(string)) > 0 {
		err := json.Unmarshal([]byte(todoJS.(string)), &todos)
		if err != nil {
			h.cache.Del()
			log.Printf("redis: %v", err)
		}
	}

	itemToCaches := map[string]interface{}{}

	if todoJS == nil {
		todos, paging, err = h.tx.FindAll(limit, page)
		if err != nil {
			return c.JSON(err)
		}
		itemToCaches[query1CacheKey] = todos
	}

	if pageJS != nil && len(pageJS.(string)) > 0 {
		err := json.Unmarshal([]byte(pageJS.(string)), &paging)
		if err != nil {
			h.cache.Del(query2CacheKey)
			log.Println(err.Error())
		}
	}

	if pageJS == nil {
		itemToCaches[query2CacheKey] = paging
	}

	if len(itemToCaches) > 0 {
		fmt.Println("MSet")
		timeToExpire := 10 * time.Second
		err := h.cache.MSet(itemToCaches)
		if err != nil {
			log.Println(err.Error())
		}

		// Set time to expire
		keys := []string{}
		for k := range itemToCaches {
			keys = append(keys, k)
		}
		err = h.cache.Expires(keys, timeToExpire)
		if err != nil {
			log.Println(err.Error())
		}
	}

	var result Pagination
	copier.Copy(&result, &paging)
	result.Rows = todos

	return c.Status(fiber.StatusOK).JSON(result)
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
