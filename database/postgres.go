package database

import (
	"os"

	"github.com/sing3demons/hello-world/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	dsn := os.Getenv("DSN")
	// dsn := fmt.Sprintf("host=service.postgres user=postgresadmin password=admin123 dbname=postgresdb% port=%s  sslmode=disable TimeZone=Asia/Bangkok", dsn)
	// dsn := fmt.Sprintf("host=service.postgres user=postgresadmin password=admin123 dbname=postgresdb% sslmode=disable TimeZone=Asia/Bangkok")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	database.Migrator().DropTable(&model.Todo{})
	database.AutoMigrate(&model.Todo{})

	db = database
}

func GetDB() *gorm.DB {
	return db
}
