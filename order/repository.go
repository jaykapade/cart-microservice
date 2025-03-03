package order

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	GetOrdersForAccount(ctx context.Context, accountID string) ([]*Order, error)
	PutOrder(ctx context.Context, o *Order) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}
func (r *PostgresRepository) GetOrdersForAccount(ctx context.Context, accountID string) ([]*Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT 
		o.id,
		o.created_at,
		o.account_id,
		o.total_price::money::numeric::float8,
		op.product_id,
		op.quantity
		FROM orders o JOIN order_products op ON o.id = op.order_id 
		WHERE o.account_id = $1
		ORDER BY o.created_at DESC`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []*Order{}
	order := &Order{}
	lastOrder := &Order{}
	orderedProduct := &OrderedProduct{}
	products := []*OrderedProduct{}

	for rows.Next() {
		err := rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.AccountID,
			&order.TotalPrice,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
		)
		if err != nil {
			return nil, err
		}
		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			newOrder := &Order{
				ID:         lastOrder.ID,
				CreatedAt:  lastOrder.CreatedAt,
				AccountID:  lastOrder.AccountID,
				TotalPrice: lastOrder.TotalPrice,
				Products:   lastOrder.Products,
			}

			orders = append(orders, newOrder)
			products = []*OrderedProduct{}
		}
		products = append(products, &OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
		})
		*lastOrder = *order
	}

	return orders, nil

}

func (r *PostgresRepository) PutOrder(ctx context.Context, o *Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO orders (id, created_at, account_id, total_price) VALUES ($1, $2, $3, $4)`,
		o.ID,
		o.CreatedAt,
		o.AccountID,
		o.TotalPrice,
	)

	if err != nil {
		return err
	}

	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	for _, p := range o.Products {
		_, err := stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return err
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}

	stmt.Close()
	return nil
}
