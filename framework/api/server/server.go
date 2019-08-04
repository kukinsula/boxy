package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"

	"github.com/gin-gonic/gin"
)

type Logger func(
	uuid interface{},
	level, message string,
	meta map[string]interface{})

type Config struct {
	Address string `yaml:"address"`
	Backend *Backend
	Logger  log.Logger
}

type API struct {
	config  Config
	backend *Backend
	logger  log.Logger
	server  *http.Server
	engine  *gin.Engine
	done    chan error
}

func NewAPI(config Config) *API {
	engine := gin.New()
	api := &API{
		config:  config,
		backend: config.Backend,
		logger:  config.Logger,
		server: &http.Server{
			Addr:    config.Address,
			Handler: engine,
		},
		engine: engine,
		done:   make(chan error),
	}

	return api
}

func (api *API) Run() {
	api.engine.Use(Welcome(api.logger))

	// Login
	api.engine.POST("/login/signin", Signin(api.backend.Login))
	api.engine.GET("/login/me", Me(api.backend.Login))
	api.engine.DELETE("/login/logout", Logout(api.backend.Login))

	api.engine.Run(api.config.Address)
}

func (api *API) Shutdown(ctx context.Context) (err error) {
	failure := make(chan error)
	defer close(failure)

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	go func() { failure <- api.server.Shutdown(ctx) }()

	select {
	case <-ctx.Done():
		err = fmt.Errorf("Server shutdown timeout")

	case err = <-failure:

	case err = <-api.done:
	}

	return err
}

func getRequestUuid(ctx *gin.Context) string {
	rawUuid, ok := ctx.Get("uuid")
	if !ok {
		return entity.NewUUID()
	}

	uuid, ok := rawUuid.(string)
	if !ok {
		return entity.NewUUID()
	}

	return uuid
}
