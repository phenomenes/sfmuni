package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/redis.v5"
)

const TTL = time.Duration(30) * time.Second

func newRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_MASTER_SERVICE_HOST" + ":6397"),
		Password: "",
		DB:       0,
	})

	return client
}

func getFromCache(client *redis.Client, req string) ([]byte, error) {
	body, err := client.Get(req).Bytes()
	switch err {
	case nil:
		return body, nil
	case redis.Nil:
		fallthrough
	default:
		return nil, err
	}
}

func writeToCache(client *redis.Client, req string, body []byte) {
	err := client.Set(req, body, TTL).Err()
	if err != nil {
		log.Printf("%s", err)
	}
}
