package server

import (
	loginUsecase "github.com/kukinsula/boxy/usecase/login"

	"github.com/gin-gonic/gin"
)

func Signin(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params loginUsecase.SigninParams

		err := ctx.BindJSON(&params)
		if err != nil {
			ctx.JSON(400, gin.H{
				"error": "BAD_JSON_BODY",
				"raw":   err,
			})
			return
		}

		uuid := getRequestUUID(ctx)
		result, err := login.Signin(uuid, ctx, params)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "SIGNIN_UNAVAILABLE"})
			return
		}

		ctx.JSON(200, result)
	}
}

func Me(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get(X_ACCESS_TOKEN)
		if token == "" {
			ctx.JSON(401, gin.H{
				"error": "MISSING ACCESS TOKEN",
			})
			return
		}

		uuid := getRequestUUID(ctx)
		result, err := login.Me(uuid, ctx, token)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "ME_UNAVAILABLE",
			})
			return
		}

		ctx.JSON(200, result)
	}
}

func Logout(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get(X_ACCESS_TOKEN)
		if token == "" {
			ctx.JSON(401, gin.H{
				"error": "MISSING ACCESS TOKEN",
			})
			return
		}

		uuid := getRequestUUID(ctx)
		err := login.Logout(uuid, ctx, token)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "LOGOUT_UNAVAILABLE",
			})
			return
		}

		ctx.JSON(204, nil)
	}
}
