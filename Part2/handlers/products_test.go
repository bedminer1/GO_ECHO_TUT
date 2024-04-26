package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bedminer1/SampleEchoServer/config"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// c *mongo.Client
	db *mongo.Database
	col *mongo.Collection
	cfg config.Properties
	h ProductHandler
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

func TestProduct(t *testing.T) {
	t.Run("test create product", func (t *testing.T) {
		body := `
		[{
			"product_name": "iphone",
			"price": 250,
			"currency": "SGD",
			"vendor": "Apple",
			"accessories": ["charger"]
		  }]
		`
		req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()

		e := echo.New()
		c := e.NewContext(req , res)
		h.Col = col
		err := h.CreateProducts(c)
		// if err == nil, test passed
		assert.Nil(t, err)
	})
}