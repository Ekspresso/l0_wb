package main

import (
	"context"
	"database/sql"
	"l0/internal/api"
	"l0/internal/cache"
	dbcli "l0/internal/db"
	"l0/internal/service"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/stan.go"
)

const (
	dbConnStr     = "postgres://konstantin:postgres@localhost:5432/konstantin_l0?sslmode=disable"
	cacheInitSize = 100
)

func main() {
	db, err := sql.Open("pgx", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbCli := dbcli.New(db)
	cacheCli := cache.New(cacheInitSize)
	svc := service.New(dbCli, cacheCli)
	if err = svc.Init(context.Background()); err != nil {
		log.Fatal(err)
	}
	h := api.New(svc)

	conn, err := stan.Connect("test-cluster", "lo", stan.NatsURL(stan.DefaultNatsURL))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Subscribe("user", h.CreateOrder)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "form.html")
	})

	http.HandleFunc("/id-order", h.IdOrder)

	http.HandleFunc("/get-order", h.GetOrderHandler)

	if err = http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal(err)
	}
}
