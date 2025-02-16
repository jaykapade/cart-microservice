package main

import (
	"log"
	"time"

	"github.com/jaykapade/cart-microservice/account"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Connection to postgres successful...", cfg.DatabaseURL)
		return err
	})
	defer r.Close()
	log.Println("Listening on port 8080...")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8080))
}
