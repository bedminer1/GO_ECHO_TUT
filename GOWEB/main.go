package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})	

	e.Logger.Print("Listening on port 8080")
	e.Logger.Fatal(e.Start("localhost:8080"))
}