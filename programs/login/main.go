package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/framework/mongo"
	redis "github.com/kukinsula/boxy/framework/redis"
	redisServer "github.com/kukinsula/boxy/framework/redis/server"
	"github.com/kukinsula/boxy/usecase"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

func main() {
	client, err := redis.NewClient(redis.Config{
		Address:     "127.0.0.1:6379",
		MaxActive:   10,
		MaxIdle:     5,
		IdleTimeout: 200 * time.Second,
		Codec:       &entity.JSONCodec{},
		Logger:      entity.StdoutLogger,
	})

	if err != nil {
		fmt.Println("redis.NewClient failed: %s", err)
		return
	}

	ctx := context.Background()
	database, err := mongo.NewDatabase(mongo.NewDatabaseParams{
		Context:  ctx,
		URI:      "mongodb://localhost:27017",
		Database: "boxy",
		Logger:   entity.StdoutLogger,
	})

	if err != nil {
		fmt.Println("NewDatabase failed: %s", err)
		return
	}

	err = database.Init(ctx)

	if err != nil {
		fmt.Println("Database.Init failed: %s", err)
		return
	}

	tokener := usecase.NewTokener("TopSecret")
	passworder := usecase.NewPassworder(10)
	login := loginUsecase.NewLogin(database.User, tokener, passworder)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go redisServer.HandleSignin(client, login)
	go redisServer.HandleMe(client, login)
	go redisServer.HandleLogout(client, login)

	<-signals
	fmt.Println("Finished!")
}
