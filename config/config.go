package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type IConfig interface {
	Port() string
	Db() string
}

type config struct {
	port string
	url  string
}

func (c *config) Port() string { return fmt.Sprintf(":%s", c.port) }
func (c *config) Db() string   { return c.url }

func LoadConfig() IConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &config{
		port: os.Getenv("PORT"),
		url:  os.Getenv("DATABASE_URL"),
	}
}
