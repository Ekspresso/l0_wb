package db

import (
	"context"
	"database/sql"
	"l0/internal/models"
)

type DBClient struct {
	db *sql.DB
}

func New(db *sql.DB) DBClient {
	return DBClient{db: db}
}

//Метод создания нового заказа в базе данных
func (db *DBClient) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	const q1 = `insert into public.orders(order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) values($1,
			 $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id`

	var id int
	err := db.db.QueryRowContext(ctx, q1, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard).Scan(&id)
	if err != nil {
		return models.Order{}, err
	}
	order.Id = id

	const q2 = `insert into public.delivery(id, name, phone, zip, city, address, region, email) values($1,
		 $2, $3, $4, $5, $6, $7, $8)`

	_, err = db.db.ExecContext(ctx, q2, id, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return models.Order{}, err
	}

	const q3 = `insert into public.payment(id, transaction, request_id, currency, provider, amount, payment_dt, bank,
		 delivery_cost, goods_total, custom_fee) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.db.ExecContext(ctx, q3, id, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return models.Order{}, err
	}

	const q4 = `insert into public.items(id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id,
		 brand, status) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for i := 0; i < len(order.Items); i++ {
		_, err = db.db.ExecContext(ctx, q4, id, order.Items[i].ChrtID, order.Items[i].TrackNumber, order.Items[i].Price,
			order.Items[i].Rid, order.Items[i].Name, order.Items[i].Sale, order.Items[i].Size, order.Items[i].TotalPrice,
			order.Items[i].NmID, order.Items[i].Brand, order.Items[i].Status)
		if err != nil {
			return models.Order{}, err
		}
	}
	return order, nil
}

//Метод получения всех данных заказа по id
func (db *DBClient) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	var order models.Order

	const q1 = `select id, order_uid, track_number, entry, locale, internal_signature,
	customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard from public.orders where id = $1`

	err := db.db.QueryRowContext(ctx, q1, id).Scan(
		&order.Id, &order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return models.Order{}, err
	}

	const q2 = `select id, name, phone, zip, city, address, region, email from public.delivery where id = $1`

	err = db.db.QueryRowContext(ctx, q2, id).Scan(
		&order.Delivery.Id, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return models.Order{}, err
	}

	const q3 = `select id, transaction, request_id, currency, provider, amount, payment_dt, bank,
	delivery_cost, goods_total, custom_fee from public.payment where id = $1`

	err = db.db.QueryRowContext(ctx, q3, id).Scan(
		&order.Payment.Id, &order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		return models.Order{}, err
	}

	const q4 = `select id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id,
	brand, status from public.items where id = $1`

	rows, err := db.db.QueryContext(ctx, q4, id)
	if err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err = rows.Scan(&item.Id, &item.ChrtID, &item.TrackNumber, &item.Price,
			&item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status); err != nil {
			return models.Order{}, err
		}
		order.Items = append(order.Items, item)
	}
	return order, nil
}

//Метод получения списка всех заказов в бд. Не оптимален, т.к. к каждой таблице обращений столько, сколько в ней записей.
//Можно реализовать за 1 обращение к каждой таблице, но тогда кода будет больше, чем в предыдущей функции.
//Для этого сначала делается проход через rows по главной таблице с выборкой всех данных, с созданием слайса структур order, и заполнением этих структур данными
//из главной таблицы. Затем делается ещё по 1 обращению к остальным таблицам по очереди с заполнением недостающих данных в order-структурах в слайсе.
func (db *DBClient) GetOrders(ctx context.Context) ([]models.Order, error) {
	const q1 = `select id from public.orders`

	var id int
	rows, err := db.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		order, err = db.GetOrderByID(ctx, id)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
