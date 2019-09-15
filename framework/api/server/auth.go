package server

import (
	"github.com/kukinsula/boxy/entity/log"

	"github.com/gin-gonic/gin"
)

func Authenticate(login LoginBackender, logger log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := getAccessToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": err})
			return
		}

		uuid := getRequestUUID(ctx)
		result, err := login.Me(uuid, ctx, token)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "ME_UNAVAILABLE"})
			return
		}

		ctx.Set(REQUESTER_INFO, result)
	}
}
