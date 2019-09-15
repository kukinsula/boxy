package server

import (
	"errors"
	"strings"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"

	"github.com/gin-gonic/gin"
)

const (
	X_REQUEST_ID_HEADER = "X-Request-ID"

	AUTHORIZATION_HEADER = "Authorization"

	ACCESS_TOKEN   = "ACCESS_TOKEN"
	REQUESTER_INFO = "REQUESTER_INFO"
)

var (
	AccessTokenMissingErr   = errors.New("Missing access token")
	AccessTokenMalformedErr = errors.New("Malformed access token")
	UnanthorizedErr         = errors.New("Unanthorized")
)

func Welcome(logger log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		uuid := getRequestUUID(ctx)

		ctx.Writer.Header().Set(X_REQUEST_ID_HEADER, uuid)

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
	rawUUIDInterface, exists := ctx.Get(X_REQUEST_ID_HEADER)
	if exists {
		uuid, ok := rawUUIDInterface.(string)
		if ok {
			return uuid
		}
	}

	uuid := ctx.Request.Header.Get(X_REQUEST_ID_HEADER)
	if uuid == "" {
		uuid = entity.NewUUID()
	}

	ctx.Set(X_REQUEST_ID_HEADER, uuid)

	return uuid
}

func getAccessToken(ctx *gin.Context) (string, error) {
	rawTokenInterface, exists := ctx.Get(ACCESS_TOKEN)
	if exists {
		token, ok := rawTokenInterface.(string)
		if ok {
			return token, nil
		}
	}

	value := ctx.Request.Header.Get(AUTHORIZATION_HEADER)

	if value == "" {
		return "", AccessTokenMissingErr
	}

	if !strings.HasPrefix(value, "Bearer ") {
		return "", AccessTokenMalformedErr
	}

	token := strings.TrimPrefix(value, "Bearer ")

	ctx.Set(ACCESS_TOKEN, token)

	return token, nil
}

func getRequesterInfo(ctx *gin.Context) (*loginUsecase.SigninResult, error) {
	rawResultInterface, exists := ctx.Get(REQUESTER_INFO)

	if !exists {
		return nil, UnanthorizedErr
	}

	result, ok := rawResultInterface.(*loginUsecase.SigninResult)
	if ok {
		return nil, UnanthorizedErr
	}

	return result, nil
}
