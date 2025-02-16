package order

import (
	"context"
	"time"

	"github.com/jaykapade/cart-microservice/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewOrderServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]*Order, error) {
	r, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{AccountId: accountID})
	if err != nil {
		return nil, err
	}

	orders := []*Order{}
	for _, orderProto := range r.Orders {
		newOrder := &Order{
			ID:         orderProto.Id,
			AccountID:  orderProto.AccountId,
			TotalPrice: orderProto.TotalPrice,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)

		products := []*OrderedProduct{}
		for _, productProto := range orderProto.Products {
			products = append(products, &OrderedProduct{
				ID:       productProto.Id,
				Quantity: productProto.Quantity,
			})
		}
		newOrder.Products = products
		orders = append(orders, newOrder)
	}

	return orders, nil

}

func (c *Client) PostOrder(ctx context.Context, accountID string, products []*OrderedProduct) (*Order, error) {
	protoProducts := []*pb.PostOrderRequest_OrderedProduct{}
	for _, p := range products {
		protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderedProduct{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}
	r, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountID,
		Products:  protoProducts,
	})
	if err != nil {
		return nil, err
	}

	newOrder := r.Order
	newOrderCreatedAt := time.Time{}
	newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt)
	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrderCreatedAt,
		AccountID:  newOrder.AccountId,
		TotalPrice: newOrder.TotalPrice,
		Products:   products,
	}, nil
}
