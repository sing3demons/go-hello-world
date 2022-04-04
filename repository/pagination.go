package repository

import (
	"math"

	"github.com/sing3demons/hello-world/model"
	"gorm.io/gorm"
)

func (repo *todoRepository) paginate(value interface{}, pagination *model.Pagination) func(db *gorm.DB) *gorm.DB {
	ch := make(chan int64)
	go repo.countRecords(ch, value)

	pagination.TotalRows = <-ch
	totalPages := int(math.Ceil(float64(pagination.TotalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.Limit).Order(pagination.GetSort())
	}
}
func (repo *todoRepository) countRecords(ch chan int64, value interface{}) {
	var totalRows int64
	repo.db.Model(value).Count(&totalRows)
	ch <- totalRows
}
