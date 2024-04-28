package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	e.POST("/tax/calculations/upload-csv", taxHandler.TaxCalculationsCSVHandler)

	adminHandler := admin.New(p)
	a := e.Group("/admin")
	a.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == cfg.Admin().User() && password == cfg.Admin().Pass() {
			return true, nil
		}
		return false, nil
	}))
	a.POST("/deductions/:deductType", adminHandler.SetupDeductionHandler)

	go func() {
		if err := e.Start(cfg.Port()); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server.")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	fmt.Println("shutting down the server.")
	if err := e.Shutdown(context.Background()); err != nil {
		e.Logger.Fatal("shutdown err:", err)
	}
	fmt.Println("shutdown complete.")
}
