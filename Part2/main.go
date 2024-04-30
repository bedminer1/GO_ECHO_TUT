package main

import (
	"context"
	"fmt"


	"github.com/bedminer1/SampleEchoServer/config"
	"github.com/bedminer1/SampleEchoServer/handlers"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"github.com/labstack/gommon/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// CorrelationID is request id unique to request
	CorrelationID = "X-Correlation-ID"
)

var (
	db *mongo.Database
	col *mongo.Collection
	cfg config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v", err)
	}

	connectURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectURI))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	db = c.Database(cfg.DBName)
	col = db.Collection(cfg.ProductCollection)
}

func addCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Request().Header.Get(CorrelationID)
		if id == "" {
			// generate ID
			id = random.String(12)
		}
		// set to existing ID
		c.Request().Header.Set(CorrelationID, id)
		c.Response().Header().Set(CorrelationID, id)
		
		return next(c)
	}
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	h := handlers.ProductHandler{Col: col}

	// MIDDLEWARE
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationID) // passes correlation ID, passing to microservices

	// HANDLERS
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M")) // 1mb payload limit
	e.GET("/products", h.GetProducts)
	e.PUT("/products/:_id", h.UpdateProduct, middleware.BodyLimit("1M"))

	e.GET("/products/:id", h.GetProduct)
	e.DELETE("/products/:id", h.DeleteProduct)
	


	// START SERVER
	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}