package repository

import (
	"context"
	"encoding/json"
	"errors"
	"order/infrastructure/constant"
	"order/models"
	"time"

	"gorm.io/gorm"
)

func (r *OrderRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.Database.Begin().WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

/*
	withTransaction(
		insert order detail
		insert order
	)

	if error => rollback
*/

func (r *OrderRepository) InsertOrderTx(ctx context.Context, tx *gorm.DB, order *models.Order) error {
	err := tx.WithContext(ctx).Table("orders").Create(order).Error

	return err
}

func (r *OrderRepository) InsertOrderDetailTx(ctx context.Context, tx *gorm.DB, orderDetail *models.OrderDetail) error {
	err := tx.WithContext(ctx).Table("order_detail").Create(orderDetail).Error

	return err
}

func (r *OrderRepository) CheckIdempotency(ctx context.Context, token string) (bool, error) {
	var log models.OrderRequestLog

	err := r.Database.WithContext(ctx).Table("order_request_log").First(&log, "idempontency_token = ?", token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *OrderRepository) SaveIdempontency(ctx context.Context, token string) error {
	log := models.OrderRequestLog{
		IdempontencyToken: token,
		CreateTime:        time.Now(),
	}

	err := r.Database.WithContext(ctx).Table("order_request_log").Create(&log).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) GetOrderHistoryByUserID(ctx context.Context, param *models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	var results []models.OrderHistoryResponse
	var queryResults []models.OrderHistoryResult

	query := r.Database.WithContext(ctx).Table("orders AS o").
		Select("o.id, o.amount, o.total_qty, o.status, o.payment_method, o.shipping_address, od.products, od.order_history").
		Joins("JOIN order_detail od ON od.id = o.order_detail_id").
		Where("o.user_id = ?", param.UserID)

	if param.Status > 0 {
		query = query.Where("o.Status = ?", param.Status)
	}

	err := query.Order("o.id DESC").Scan(&queryResults).Error
	if err != nil {
		return nil, err
	}

	for _, result := range queryResults {
		var products []models.CheckoutItem
		var orderHistory []models.StatusHistory

		err = json.Unmarshal([]byte(result.Products), &products)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(result.OrderHistory), &orderHistory)
		if err != nil {
			return nil, err
		}

		results = append(results, models.OrderHistoryResponse{
			OrderID:         result.ID,
			TotalAmount:     result.Amount,
			TotalQty:        result.TotalQty,
			Status:          constant.OrderStatusTranslated[result.Status],
			PaymentMethod:   result.PaymentMethod,
			ShippingAddress: result.ShippingAddress,
			Products:        products,
			History:         orderHistory,
		})
	}

	return results, nil
}
