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
	Admin() IAdmin
}

type config struct {
	port  string
	url   string
	admin *admin
}

type IAdmin interface {
	User() string
	Pass() string
}

type admin struct {
	adminUsername string
	adminPassword string
}

func (c *config) Port() string  { return fmt.Sprintf(":%s", c.port) }
func (c *config) Db() string    { return c.url }
func (c *config) Admin() IAdmin { return c.admin }
func (a *admin) User() string   { return a.adminUsername }
func (a *admin) Pass() string   { return a.adminPassword }

func LoadConfig() IConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &config{
		port: os.Getenv("PORT"),
		url:  os.Getenv("DATABASE_URL"),
		admin: &admin{
			adminUsername: os.Getenv("ADMIN_USERNAME"),
			adminPassword: os.Getenv("ADMIN_PASSWORD"),
		},
	}
}
