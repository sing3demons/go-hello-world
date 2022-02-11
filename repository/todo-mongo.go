package repository

import (
	"context"
	"time"

	"github.com/sing3demons/hello-world/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type todoMongoRepository struct {
	db *mongo.Database
}

type TodoMongoRepository interface {
	Create(todo model.Todo) (*model.Todo, error)
	FindAll() ([]model.Todo, error)
}

func NewTodoMongoRepository(db *mongo.Database) TodoMongoRepository {
	return &todoMongoRepository{db: db}
}

func (tx *todoMongoRepository) collection() *mongo.Collection {
	return tx.db.Collection("todos")
}

func (tx *todoMongoRepository) FindAll() ([]model.Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := tx.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var todo []model.Todo
	if err := cursor.All(ctx, &todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (tx *todoMongoRepository) Create(todo model.Todo) (*model.Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := tx.collection().InsertOne(ctx, todo)
	if err != nil || result == nil {
		return nil, err
	}
	return &todo, nil
}
