package handlers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bedminer1/SampleEchoServer/dbiface"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

var (
	v = validator.New()
)

//Product describes an electronic product e.g. phone
type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"product_name" bson:"product_name" validate:"required,max=10"`
	Price       int                `json:"price" bson:"price" validate:"required,max=2000"`
	Currency    string             `json:"currency" bson:"currency" validate:"required,len=3"`
	Discount    int                `json:"discount" bson:"discount"`
	Vendor      string             `json:"vendor" bson:"vendor" validate:"required"`
	Accessories []string           `json:"accessories,omitempty" bson:"accessories,omitempty"`
	IsEssential bool               `json:"is_essential" bson:"is_essential"`
}

// ProductHandler pass in col(reference to mongodb collection) as attribute
type ProductHandler struct {
	Col dbiface.CollectionAPI
}

// ProductValidator class with validate method
type ProductValidator struct {
	validator *validator.Validate
}


// Validate method that validates a product
func (p *ProductValidator) Validate(i interface{}) error {
	return p.validator.Struct(i)
}
 
func findProducts(ctx context.Context, q url.Values, collection dbiface.CollectionAPI) ([]Product, error) {
	var products []Product
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}

	cursor, err := collection.Find(ctx, bson.M(filter))
	if err != nil {
		log.Errorf("Unable to find products: %v", err)
		return products, err
	}
	if err := cursor.All(ctx, &products); err != nil {
		log.Errorf("Unable to read cursor: %v", err)
		return products, err
	}
	return products, nil
}

// GetProducts is a HandlerFunc that responds with a list of products
func (h ProductHandler) GetProducts(c echo.Context) error {
	products, err := findProducts(context.Background(), c.QueryParams(), h.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, products)
}

// insertProducts generates IDs and inserts products into mongo col
func insertProducts(ctx context.Context, products []Product, collection dbiface.CollectionAPI) ([]interface{}, error) {
	var insertedIds []interface{}
	for _, product := range products {
		product.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, product)
		if err != nil {
			log.Errorf("Unable to insert %v", err)
			return nil, err
		}
		insertedIds = append(insertedIds, insertID.InsertedID)
	}
	return insertedIds, nil
}

// CreateProducts create products on mongodb and responds with IDs of products
func (h *ProductHandler) CreateProducts(c echo.Context) error {
	var products []Product
	c.Echo().Validator = &ProductValidator{validator: v}

	// bind echoContext to products
	if err := c.Bind(&products); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return err
	}

	// validate products
	for _, product := range products {
		if err := c.Validate(product); err != nil {
			log.Errorf("Unable to validate product %+v, %v", product, err)
			return err
		}
	}

	IDs, err := insertProducts(context.Background(), products, h.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, IDs)
}