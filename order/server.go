package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/jaykapade/cart-microservice/account"
	"github.com/jaykapade/cart-microservice/catalog"
	"github.com/jaykapade/cart-microservice/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListenGRPC(s Service, accountURL string, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}
	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		catalogClient.Close()
		accountClient.Close()
		return err
	}

	serv := grpc.NewServer()

	pb.RegisterOrderServiceServer(serv, &grpcServer{
		UnimplementedOrderServiceServer: pb.UnimplementedOrderServiceServer{},
		service:                         s,
		accountClient:                   accountClient,
		catalogClient:                   catalogClient,
	})

	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	accountOrders, err := s.service.GetOrdersForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting orders for account", err)
		return nil, errors.New("Error getting orders for account")
	}

	productIDMap := map[string]bool{}
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIDMap[p.ID] = true
		}
	}

	productIDs := []string{}
	for id := range productIDMap {
		productIDs = append(productIDs, id)
	}

	products, err := s.catalogClient.GetProducts(ctx, 0, 0, productIDs, "")
	if err != nil {
		log.Println("Error getting products", err)
		return nil, errors.New("Error getting products")
	}

	orders := []*pb.Order{}

	for _, o := range accountOrders {
		op := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}

		op.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		for _, product := range o.Products {
			for _, p := range products {
				if product.ID == p.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}

			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}

		orders = append(orders, op)

	}

	return &pb.GetOrdersForAccountResponse{
		Orders: orders,
	}, nil
}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account", err)
		return nil, errors.New("Error getting account")
	}
	productIds := []string{}
	orderedPdts, err := s.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		log.Println("Error getting products", err)
		return nil, errors.New("Error getting products")
	}

	products := []*OrderedProduct{}
	for _, pdt := range orderedPdts {
		product := OrderedProduct{
			ID:          pdt.ID,
			Name:        pdt.Name,
			Description: pdt.Description,
			Price:       pdt.Price,
			Quantity:    0,
		}

		for _, order := range r.Products {
			if order.ProductId == pdt.ID {
				product.Quantity = order.Quantity
				break
			}
		}

		if product.Quantity > 0 {
			products = append(products, &product)
		}
	}

	o, err := s.service.PostOrder(ctx, r.AccountId, products)

	if err != nil {
		log.Println("Error posting order", err)
		return nil, errors.New("Error posting order")
	}

	orderProto := &pb.Order{
		Id:         o.ID,
		AccountId:  o.AccountID,
		TotalPrice: o.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}

	orderProto.CreatedAt, _ = o.CreatedAt.MarshalBinary()

	for _, p := range o.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}

	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}
