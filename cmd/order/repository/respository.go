package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"order/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	Database    *gorm.DB
	ProductHost string
}

func NewOrderRepository(db *gorm.DB, productHost string) *OrderRepository {
	return &OrderRepository{
		Database:    db,
		ProductHost: productHost,
	}
}

func (r *OrderRepository) GetProductInfo(ctx context.Context, productID int64) (models.Product, error) {
	var product models.Product
	var response models.GetProductInfo

	url := fmt.Sprintf("%s/v1/product/%d", r.ProductHost, productID)
	fmt.Println(url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.Product{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.Product{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return models.Product{}, fmt.Errorf("Invalid response - get product info")
	}

	err = json.NewDecoder(res.Body).Decode(&response)

	product = response.Product
	return product, nil
}
