package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func main() {
	port := os.Getenv("MY_APP_PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	// DATABASE CONNECTION (SIMULATE)
	products := []map[int]string{{1: "phone"}, {2: "tv"}, {3: "laptop"}}

	// HANDLERS
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})	
	e.GET("/products", func(c echo.Context) error {
		return c.JSON(http.StatusOK, products)
	})
	e.GET("/products/:id", func(c echo.Context) error {
		var product map[int]string
		for _, p := range products {
			for k := range p {
				pID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					return err
				}
				if pID == k {
					product = p
				}
			}
		}
		if product == nil {
			return c.JSON(http.StatusNotFound, "Product Not Found")
		}

		return c.JSON(http.StatusOK, product)
	})
	e.POST("/products", func(c echo.Context) error {
		type body struct {
			Name string `json:"product_name"`
		}
		var reqBody body
		err := c.Bind(&reqBody)
		if err != nil {
			return err
		}

		product := map[int]string{
			len(products) + 1: reqBody.Name,
		}
		products = append(products, product)
		return c.JSON(http.StatusOK, product)
	})

	// START SERVER
	e.Logger.Print(fmt.Sprintf("Listening on port %s", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf("localhost:%s", port)))
}