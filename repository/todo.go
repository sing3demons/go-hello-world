package repository

import (
	"github.com/sing3demons/hello-world/model"
	"gorm.io/gorm"
)

type todoRepository struct {
	db *gorm.DB
}

type TodoRepository interface {
	Create(todo model.Todo) (*model.Todo, error)
	FindAll(limit int, page int) ([]model.Todo, *model.Pagination, error)
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{
		db: db,
	}
}

func (tx *todoRepository) Create(todo model.Todo) (*model.Todo, error) {
	if err := tx.db.Create(&todo).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}
func (tx *todoRepository) FindAll(limit int, page int) ([]model.Todo, *model.Pagination, error) {
	todos := []model.Todo{}

	pagination := model.Pagination{
		Limit: limit,
		Page:  page,
	}

	if err := tx.db.Scopes(tx.paginate(&todos, &pagination)).Find(&todos).Error; err != nil {
		return nil, nil, err
	}

	// if err:=tx.db.Find(&todos).Error;err!=nil{
	// 	return nil,err
	// }

	return todos, &pagination, nil
}
