package config

import (
	"os"

	"github.com/go-pg/pg/v10"
)

func NewPostgres() (*pg.DB) {
	addr := os.Getenv("PG_ADDR")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	database := os.Getenv("PG_DATABASE")

	db := pg.Connect(&pg.Options{
		Addr:     addr, // example: "localhost:5432"
		User:     user,
		Password: password,
		Database: database,
	})
	db.AddQueryHook(dbLogger{}) 
	// Check connection with simple query
	_, err := db.Exec("SELECT 1")
	if err != nil {
		panic("Unable to connect to PostgreSQL" + err.Error())
	}
	return db
}
