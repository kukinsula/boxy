package server

import (
	loginUsecase "github.com/kukinsula/boxy/usecase/login"

	"github.com/gin-gonic/gin"
)

func Signup(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params loginUsecase.CreateUserParams

		err := ctx.BindJSON(&params)
		if err != nil {
			ctx.JSON(400, gin.H{
				"error":   "INVALID_JSON",
				"message": err.Error(),
			})
			return
		}

		uuid := getRequestUUID(ctx)
		user, err := login.Signup(uuid, ctx, &params)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "SIGNUP_UNVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(201, user)
	}
}

func CheckActivate(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params loginUsecase.EmailAndTokenParams

		err := ctx.ShouldBindQuery(&params)
		if err != nil {
			ctx.JSON(400, gin.H{
				"error":   "INVALID_QUERY",
				"message": err.Error(),
			})
			return
		}

		uuid := getRequestUUID(ctx)
		err = login.CheckActivate(uuid, ctx, &params)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "CHECK_ACTIVATE_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(204, nil)
	}
}

func Activate(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params loginUsecase.EmailAndTokenParams

		err := ctx.BindJSON(&params)
		if err != nil {
			ctx.JSON(400, gin.H{
				"error":   "INVALID_JSON",
				"message": err.Error(),
			})
			return
		}

		uuid := getRequestUUID(ctx)
		err = login.Activate(uuid, ctx, &params)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "ACTIVATE_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(204, nil)
	}
}

func Signin(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var params loginUsecase.SigninParams

		err := ctx.BindJSON(&params)
		if err != nil {
			ctx.JSON(400, gin.H{
				"error":   "INVALID_JSON",
				"message": err.Error(),
			})
			return
		}

		uuid := getRequestUUID(ctx)
		result, err := login.Signin(uuid, ctx, &params)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "SIGNIN_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, result)
	}
}

func Me(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uuid := getRequestUUID(ctx)
		token, err := getAccessToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "AccessToken missing"})
			return
		}

		result, err := login.Me(uuid, ctx, token)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "ME_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(200, result)
	}
}

func Logout(login LoginBackender) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uuid := getRequestUUID(ctx)
		token, err := getAccessToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "AccessToken missing"})
			return
		}

		err = login.Logout(uuid, ctx, token)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":   "LOGOUT_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(204, nil)
	}
}
