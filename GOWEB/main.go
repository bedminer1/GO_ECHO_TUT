package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	port := os.Getenv("MY_APP_PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	// DATABASE CONNECTION
	products := []map[int]string{{1: "phone"}, {2: "tv"}, {3: "laptop"}}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})	
	e.GET("/products", func(c echo.Context) error {
		return c.JSON(http.StatusOK, products)
	})

	e.Logger.Print(fmt.Sprintf("Listening on port %s", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf("localhost:%s", port)))
}