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
	FindAll() ([]model.Todo, error)
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
	return &todo,nil
}
func (tx *todoRepository) FindAll() ([]model.Todo, error){
	todos := []model.Todo{}
	if err:=tx.db.Find(&todos).Error;err!=nil{
		return nil,err
	}

	return todos,nil
}
