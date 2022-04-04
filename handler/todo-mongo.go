package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"github.com/sing3demons/hello-world/cache"
	"github.com/sing3demons/hello-world/model"
	"github.com/sing3demons/hello-world/repository"
)

type todoMongoHandler struct {
	tx    repository.TodoMongoRepository
	cache *cache.Cacher
}

type TodoMongoHandler interface {
	FindTodos(c *fiber.Ctx) error
	InsertTodos(c *fiber.Ctx) error
}

func NewTodoMongoHandler(tx repository.TodoMongoRepository, cache *cache.Cacher) TodoHandler {
	return &todoMongoHandler{tx: tx, cache: cache}
}

func (h *todoMongoHandler) FindTodos(c *fiber.Ctx) error {

	query1CacheKey := "todo::all"
	query2CacheKey := "todo::page"

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
		todos, err = h.tx.FindAll()
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

func (h *todoMongoHandler) InsertTodos(c *fiber.Ctx) error {
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
