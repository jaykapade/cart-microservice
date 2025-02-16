package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"gopkg.in/olivere/elastic.v5"
)

var (
	ErrNotFound = errors.New("not found")
)

type Repository interface {
	Close()
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]*Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]*Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]*Product, error)
	PutProduct(ctx context.Context, p *Product) error
}

type ElasticRepository struct {
	client *elastic.Client
}

type ProductDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
	)

	if err != nil {
		return nil, err
	}

	return &ElasticRepository{client: client}, nil
}

func (r *ElasticRepository) Close() {
}

func (r *ElasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get().
		Index("catalog").
		Type("product").
		Id(id).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, ErrNotFound
	}

	p := ProductDocument{}
	if err := json.Unmarshal(*res.Source, &p); err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil
}

func (r *ElasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]*Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").
		Query(elastic.NewMatchAllQuery()).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	products := []*Product{}

	for _, hit := range res.Hits.Hits {
		p := ProductDocument{}
		if err := json.Unmarshal(*hit.Source, &p); err == nil {
			products = append(products, &Product{
				ID:          hit.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}

	return products, nil
}

func (r *ElasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]*Product, error) {
	items := []*elastic.MultiGetItem{}
	for _, id := range ids {
		items = append(items,
			elastic.NewMultiGetItem().
				Index("catalog").
				Type("product").
				Id(id),
		)
	}

	res, err := r.client.MultiGet().
		Add(items...).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	products := []*Product{}

	for _, doc := range res.Docs {
		p := ProductDocument{}
		if err := json.Unmarshal(*doc.Source, &p); err == nil {
			products = append(products, &Product{
				ID:          doc.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}

	}

	return products, nil
}

func (r *ElasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]*Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").
		Query(elastic.NewMultiMatchQuery(query, "name", "description")).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	products := []*Product{}
	for _, hit := range res.Hits.Hits {
		p := ProductDocument{}
		if err := json.Unmarshal(*hit.Source, &p); err == nil {
			products = append(products, &Product{
				ID:          hit.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}

	return products, nil

}

func (r *ElasticRepository) PutProduct(ctx context.Context, p *Product) error {
	_, err := r.client.Index().
		Index("catalog").
		Type("product").
		Id(p.ID).
		BodyJson(ProductDocument{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}).
		Do(ctx)

	return err
}
