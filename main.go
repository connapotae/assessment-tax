package main

import (
	"net/http"

	"github.com/connapotae/assessment-tax/config"
	"github.com/connapotae/assessment-tax/postgres"
	"github.com/connapotae/assessment-tax/tax"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.LoadConfig()

	p, err := postgres.New(cfg)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	handler := tax.New(p)
	e.POST("/tax/calculations", handler.TaxCalculationsHandler)

	e.Logger.Fatal(e.Start(cfg.Port()))
}
