package client

import (
	"fmt"

	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

type Login struct {
	*client
}

func NewLogin(
	URL string,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *Login {

	return &Login{
		client: newClient(
			URL,
			newRequester(),
			&JSONCodec{},
			requestLogger,
			responseLogger),
	}
}

func (login *Login) Signin(
	uuid, email, password string) (*loginUsecase.SigninResult, error) {

	result := &loginUsecase.SigninResult{}
	resp, err := login.POST(&Request{
		UUID: uuid,
		Path: "/login/signin",
		Headers: map[string][]string{
			"Encoding-Type": []string{"application/json"},
		},
		Body: map[string]interface{}{
			"email":    email,
			"password": password,
		},
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("Signin should return Status code 200, not %d", resp.Status)
	}

	return result, nil
}

func (login *Login) Me(uuid, token string) (*loginUsecase.SigninResult, error) {
	result := &loginUsecase.SigninResult{}
	resp, err := login.GET(&Request{
		UUID: uuid,
		Path: "/login/me",
		Headers: map[string][]string{
			"X-Access-Token": []string{token},
		},
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("Me should return Status code 200, not %d", resp.Status)
	}

	return result, nil
}

func (login *Login) Logout(uuid, token string) error {
	resp := login.DELETE(&Request{
		UUID: uuid,
		Path: "/login/logout",
		Headers: map[string][]string{
			"X-Access-Token": []string{token},
		},
	})

	if resp.Error != nil {
		return resp.Error
	}

	if resp.Status != 204 {
		return fmt.Errorf("Logout should return Status code 200, not %d", resp.Status)
	}

	return nil
}
