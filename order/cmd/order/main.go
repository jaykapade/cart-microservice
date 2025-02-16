package main

import (
	"log"
	"time"

	"github.com/jaykapade/cart-microservice/order"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	var r order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		var err error
		r, err = order.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Connection to postgres successful...", cfg.DatabaseURL)
		return err
	})
	defer r.Close()
	log.Println("Listening on port 8080...")
	s := order.NewOrderService(r)
	log.Fatal(order.ListenGRPC(s, cfg.AccountURL, cfg.CatalogURL, 8080))
}
