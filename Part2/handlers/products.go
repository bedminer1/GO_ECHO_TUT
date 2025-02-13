package handlers

import (
	"context"
	"encoding/json"
	"io"
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

	// filter is a map of query param keys to query param values
	filter := make(map[string]interface{})
	for k, v := range q { // setting first value as value (simplified implementation)
		filter[k] = v[0]
	}

	// changing id from type string to type primitive.ObjectID
	if filter["_id"] != nil { 
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return products, err
		}
		filter["_id"] = docID
	}

	// cursor is a a list of cursors to items in the db that match filter
	cursor, err := collection.Find(ctx, bson.M(filter))
	if err != nil {
		log.Errorf("Unable to find products: %v", err)
		return products, err
	}
	// All will write items pointed to by cursor into the products slice
	if err := cursor.All(ctx, &products); err != nil {
		log.Errorf("Unable to read cursor: %v", err)
		return products, err
	}
	return products, nil
}

// GetProducts is a HandlerFunc that responds with a list of products
func (h *ProductHandler) GetProducts(c echo.Context) error {
	products, err := findProducts(context.Background(), c.QueryParams(), h.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, products)
}

// findProduct finds product with given ID and returns its data
func findProduct(ctx context.Context, id string, collection dbiface.CollectionAPI) (Product, *echo.HTTPError) {
	var product Product

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return product, echo.NewHTTPError(http.StatusInternalServerError, "unable to convert to docID")
	}

	res := collection.FindOne(ctx, bson.M{"_id": docID})
	err = res.Decode(&product)
	if err != nil {
		return product, echo.NewHTTPError(http.StatusNotFound, "unable to find product")
	}

	return product, nil
} 

// GetProduct gets a single product
func (h *ProductHandler) GetProduct(c echo.Context) error {
	product, err := findProduct(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, product)
}

// removeProduct finds product with id given and removes from db, return deletedCount(1)
func removeProduct(ctx context.Context, id string, collection dbiface.CollectionAPI) (int64, error) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}

	res, err := collection.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

// DeleteProduct deletes a product with given id
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	delCount, err := removeProduct(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, delCount)
}


func modifyProduct(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.CollectionAPI) (Product, error) {
	var product Product

	// find if products exists : return 404
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Could not convert to ObjectID: %v", err)
		return product, err
	}
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)

	if err := res.Err(); err != nil {
		log.Errorf("Could not find product in db: %v", err)
		return product, err
	}

	// decode the request payload 
	if err := json.NewDecoder(reqBody).Decode(&product); err != nil {
		log.Errorf("Could not decode reqBody: %v", err)
		return product, err
	}

	// validate request
	if err := v.Struct(product); err != nil {
		log.Errorf("Could not validate product: %v", err)
		return product, err
	}

	// update the product in db
	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set":product}); err != nil {
		log.Errorf("Could not update product in db: %v", err)
		return product, nil
	}


	return product, nil
}

// UpdateProduct updates a product in the db
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	product, err := modifyProduct(context.Background(), c.Param("_id"), c.Request().Body, h.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, product)
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