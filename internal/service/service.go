package service

import (
	"context"
	"errors"
	"l0/internal/cache"
	"l0/internal/db"
	"l0/internal/models"
)

type Service struct {
	db    db.DBClient
	cache *cache.Cache
}

func New(db db.DBClient, cache *cache.Cache) Service {
	return Service{
		db:    db,
		cache: cache,
	}
}

//Метод восстановления кэша из БД при запуске сервиса
func (s Service) Init(ctx context.Context) error {
	orders, err := s.db.GetOrders(ctx)
	if err != nil {
		return err
	}
	for _, u := range orders {
		s.cache.Set(u)
	}
	return nil
}

//Метод получения данных заказа по его ID из кэша
func (s Service) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	u, ok := s.cache.Get(id)
	if !ok {
		return models.Order{}, errors.New("order not found")
	}
	return u, nil
}

//Метод создания заказа в базе данных и в кэше
func (s Service) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	order, err := s.db.CreateOrder(ctx, order)
	if err != nil {
		return models.Order{}, err
	}
	s.cache.Set(order)
	return order, nil
}
