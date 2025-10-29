package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"

	"server/pkg/app/query"
	"server/pkg/app/service"
	"server/pkg/infrastructure/redis/repo"
	"server/pkg/infrastructure/transport"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	urlRepo := repo.NewUrlRepository(rdb)
	urlQueryService := query.NewUrlQueryService(urlRepo)
	urlService := service.NewUrlService(urlRepo)

	ctx := context.Background()
	handler := transport.NewHandler(ctx, urlService, urlQueryService)
	r := mux.NewRouter()
	
	r.HandleFunc("/", handler.Index).Methods(http.MethodGet)
	r.HandleFunc("/shorten/create", handler.ShortenCreate).Methods(http.MethodPost)
	r.HandleFunc("/shorten/result", handler.ShortenResult).Methods(http.MethodGet)
	r.HandleFunc("/{url}", handler.RedirectLongUrl).Methods(http.MethodGet)
	r.HandleFunc("/shorten/batchSave", handler.BatchSave).Methods(http.MethodPost)

	log.Println(fmt.Sprintf("Starting server on %s", os.Getenv("LISTENING_SERVER_PORT")))
	http.ListenAndServe(os.Getenv("LISTENING_SERVER_PORT"), r)
}
