package main

import (
	"context"
	"time"
)

type accountResolver struct {
	server *Server
}

func (r *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	orderList, err := r.server.orderClient.GetOrdersForAccount(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	var orders []*Order

	for _, order := range orderList {
		var products []*OrderedProduct
		for _, product := range order.Products {
			products = append(products, &OrderedProduct{
				ID:       product.ID,
				Name:     product.Name,
				Price:    product.Price,
				Quantity: int(product.Quantity),
			})
		}
		orders = append(orders, &Order{
			ID:         order.ID,
			CreatedAt:  order.CreatedAt,
			Products:   products,
			TotalPrice: order.TotalPrice,
		})
	}

	return orders, nil

}
