package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity/codec"
	"github.com/kukinsula/boxy/entity/log"
	"github.com/kukinsula/boxy/framework/api/server"
	redis "github.com/kukinsula/boxy/framework/redis"
	redisClient "github.com/kukinsula/boxy/framework/redis/client"
)

func main() {
	logger := log.CleanMetaLogger(log.StdoutLogger)
	client, err := redis.NewClient(redis.Config{
		Address:     "127.0.0.1:6379",
		MaxActive:   10,
		MaxIdle:     5,
		IdleTimeout: 200 * time.Second,
		Codec:       &codec.JSONCodec{},
		Logger:      logger,
	})

	if err != nil {
		fmt.Println("redis.NewClient failed: %s", err)
		return
	}

	login := redisClient.NewLogin(client)
	streaming := redisClient.NewStreaming(client)
	backend := server.NewBackend(login, streaming)
	api := server.NewAPI(server.Config{
		Address: "127.0.0.1:9000",
		Backend: backend,
		Logger:  logger,
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go api.Run()

	<-signals

	fmt.Println("Finished!")
}
