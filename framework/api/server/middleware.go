package server

import (
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"

	"github.com/gin-gonic/gin"
)

const (
	X_REQUEST_ID   = "X-Request-ID"
	X_ACCESS_TOKEN = "X-Access-Token"
)

// TODO: Context personnalisé composé de gin.Context
//   * contenant UUID, Start Time, *User, ...
//
// type Middleware func(*Context)

func Welcome(logger log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		uuid := getRequestUUID(ctx)

		ctx.Set("id", uuid)
		ctx.Writer.Header().Set(X_REQUEST_ID, uuid)

		logger(uuid, log.DEBUG, "API <-",
			map[string]interface{}{
				"method": ctx.Request.Method,
				"path":   ctx.Request.URL.Path,
			})

		ctx.Next()

		logger(uuid, log.DEBUG, "API ->",
			map[string]interface{}{
				"method":   ctx.Request.Method,
				"path":     ctx.Request.URL.Path,
				"status":   ctx.Writer.Status(),
				"duration": time.Since(start),
			})
	}
}

func getRequestUUID(ctx *gin.Context) string {
	rawUUID := ctx.Request.Header.Get(X_REQUEST_ID)
	if rawUUID != "" {
		return rawUUID
	}

	rawUUIDInterface, exists := ctx.Get(X_REQUEST_ID)
	if !exists {
		uuid, ok := rawUUIDInterface.(string)
		if ok {
			return uuid
		}
	}

	return entity.NewUUID()
}
