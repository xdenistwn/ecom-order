package models

import "time"

type GetProductInfo struct {
	Product `json:"product"`
}

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"category_id"`
}

type ProductStockUpdateEvent struct {
	OrderID   int64         `json:"order_id"`
	Products  []ProductItem `json:"products"`
	EventTime time.Time     `json:"event_time"`
}

type ProductItem struct {
	ProductID int64 `json:"product_id"`
	Qty       int   `json:"quantity"`
}
