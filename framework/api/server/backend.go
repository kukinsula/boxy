package server

import (
	"context"

	loginEntity "github.com/kukinsula/boxy/entity/login"
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
	Signin(
		uuid string,
		context context.Context,
		params loginUsecase.SigninParams) (*loginEntity.User, error)

	Me(
		uuid string,
		context context.Context,
		token string) (*loginEntity.User, error)

	Logout(
		uuid string,
		context context.Context,
		token string) (*loginEntity.User, error)
}
