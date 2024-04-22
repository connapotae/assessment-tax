package postgres

import (
	"database/sql"
	"log"

	"github.com/connapotae/assessment-tax/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New(cfg config.IConfig) (*Postgres, error) {
	databaseSource := cfg.Db()

	db, err := sql.Open("postgres", databaseSource)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return &Postgres{Db: db}, nil
}
