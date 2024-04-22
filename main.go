package main

import (
	"net/http"

	"github.com/connapotae/assessment-tax/admin"
	"github.com/connapotae/assessment-tax/config"
	"github.com/connapotae/assessment-tax/postgres"
	"github.com/connapotae/assessment-tax/tax"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	taxHandler := tax.New(p)
	e.POST("/tax/calculations", taxHandler.TaxCalculationsHandler)

	adminHandler := admin.New(p)
	a := e.Group("/admin")
	a.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == cfg.Admin().User() && password == cfg.Admin().Pass() {
			return true, nil
		}
		return false, nil
	}))
	a.POST("/deductions/:deductType", adminHandler.SetupDeductionHandler)

	e.Logger.Fatal(e.Start(cfg.Port()))
}
