package service

import (
	"context"
	"order/cmd/order/repository"
	"order/models"

	"gorm.io/gorm"
)

type OrderService struct {
	OrderRepository *repository.OrderRepository
}

func NewOrderService(orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{
		OrderRepository: orderRepo,
	}
}

func (s *OrderService) CheckIdempotency(ctx context.Context, token string) (bool, error) {
	isExist, err := s.OrderRepository.CheckIdempotency(ctx, token)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (s *OrderService) SaveIdempontency(ctx context.Context, token string) error {
	err := s.OrderRepository.SaveIdempontency(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) GetOrderInfoByOrderID(ctx context.Context, orderID int64) (models.Order, error) {
	order, err := s.OrderRepository.GetOrderInfoByOrderID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (s *OrderService) GetOrderDetailByID(ctx context.Context, orderDetailID int64) (models.OrderDetail, error) {
	orderDetail, err := s.OrderRepository.GetOrderDetailByID(ctx, orderDetailID)
	if err != nil {
		return models.OrderDetail{}, err
	}

	return orderDetail, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int) error {
	err := s.OrderRepository.UpdateOrderStatus(ctx, orderID, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) SaveOrderAndOrderDetail(ctx context.Context, order *models.Order, orderDetail *models.OrderDetail) (int64, error) {
	var orderID int64

	err := s.OrderRepository.WithTransaction(ctx, func(tx *gorm.DB) error {
		err := s.OrderRepository.InsertOrderDetailTx(ctx, tx, orderDetail)
		if err != nil {
			return err
		}

		order.OrderDetailID = orderDetail.ID
		err = s.OrderRepository.InsertOrderTx(ctx, tx, order)
		if err != nil {
			return err
		}

		orderID = order.ID
		return nil
	})

	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (s *OrderService) GetOrderHistoryByUserID(ctx context.Context, param *models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	orderHistory, err := s.OrderRepository.GetOrderHistoryByUserID(ctx, param)
	if err != nil {
		return nil, err
	}

	return orderHistory, nil
}

func (s *OrderService) GetProductInfo(ctx context.Context, productID int64) (models.Product, error) {
	productInfo, err := s.OrderRepository.GetProductInfo(ctx, productID)
	if err != nil {
		return models.Product{}, err
	}

	return productInfo, nil
}
