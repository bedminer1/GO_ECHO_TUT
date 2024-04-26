package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bedminer1/SampleEchoServer/config"
	"github.com/bedminer1/SampleEchoServer/handlers"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	c *mongo.Client
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

func main() {
	e := echo.New()
	h := handlers.ProductHandler{Col: col}

	// HANDLERS
	e.POST("/products", h.CreateProducts)


	// START SERVER
	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}