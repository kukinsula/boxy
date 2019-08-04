package server

import (
	"context"

	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

type Backend struct {
	Login LoginBackender
}

func NewBackend(login LoginBackender) *Backend {
	return &Backend{
		Login: login,
	}
}

type LoginBackender interface {
	Signin(uuid string,
		context context.Context,
		params loginUsecase.SigninParams) (*loginUsecase.SigninResult, error)

	Me(uuid string,
		context context.Context,
		token string) (*loginUsecase.SigninResult, error)

	Logout(uuid string,
		context context.Context,
		token string) error
}
