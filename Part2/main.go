package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bedminer1/SampleEchoServer/config"
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
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("Config cannot be read: %v", err)
	}
	connectURI := fmt.Sprintf("mongodb://%s%s", cfg.DBHost, cfg.DBPort)
	mongo.Connect(context.Background(), options.Client().ApplyURI(connectURI))
}

func main() {
	e := echo.New()

	// HANDLERS
	e.POST("/products", CreateProduct)

	e.Logger.Fatal(e.Start(":8000"))
}