package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/framework/api/client"
)

func main() {
	logger := entity.CleanMetaLogger(entity.StdoutLogger)
	service := client.NewService("http://127.0.0.1:9000",
		func(req *client.Request) {
			logger(entity.Log{
				UUID:    req.UUID,
				Level:   "debug",
				Message: fmt.Sprintf("HTTP %s %s", req.Method, req.URL),
				Meta: map[string]interface{}{
					"headers": req.Headers,
					"body":    req.Body,
				},
			})
		},

		func(resp *client.Response) {
			logger(entity.Log{
				UUID:    resp.Request.UUID,
				Level:   "debug",
				Message: fmt.Sprintf("HTTP %s => %s", resp.Request.Method, resp.Request.URL),
				Meta: map[string]interface{}{
					"duration": resp.Duration,
					"headers":  resp.Headers,
					"status":   resp.Status,
					"error":    resp.Error,
				},
			})
		})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		randomer := rand.New(rand.NewSource(99))

		for index := 0; index < 1000; index++ {
			worker(service, logger, randomer)
		}
	}()

	<-signals

	fmt.Println("Finished!")
}

func worker(service *client.Service, logger entity.Logger, randomer *rand.Rand) {
	var err error

	for err == nil {
		resultSignin, err := service.Login.Signin(entity.NewUUID(), "titi@mail.io", "Azerty1234.")
		if err != nil {
			break
		}

		logger(entity.Log{
			UUID:    entity.NewUUID(),
			Level:   "debug",
			Message: "Signin result",
			Meta:    map[string]interface{}{"result": *resultSignin, "error": err},
		})

		time.Sleep(time.Duration(randomer.Float64()) * time.Second)

		resultMe, err := service.Login.Me(entity.NewUUID(), resultSignin.AccessToken)
		if err != nil {
			break
		}

		logger(entity.Log{
			UUID:    entity.NewUUID(),
			Level:   "debug",
			Message: "Me result",
			Meta:    map[string]interface{}{"result": *resultMe, "error": err},
		})

		time.Sleep(time.Duration(randomer.Float64()) * time.Second)

		err = service.Login.Logout(entity.NewUUID(), resultMe.AccessToken)
		if err != nil {
			break
		}

		logger(entity.Log{
			UUID:    entity.NewUUID(),
			Level:   "debug",
			Message: "Logout result",
			Meta:    map[string]interface{}{"error": err},
		})

		time.Sleep(time.Duration(randomer.Float64()) * time.Second)
	}

	logger(entity.Log{
		UUID:    entity.NewUUID(),
		Level:   "error",
		Message: "worker failed",
		Meta:    map[string]interface{}{"error": err},
	})
}
