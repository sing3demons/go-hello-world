package model

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	Name string `bson:"name" json:"name"`
}
