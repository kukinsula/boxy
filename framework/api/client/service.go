package client

type Service struct {
	Login *Login
}

func NewService(
	URL string,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *Service {

	return &Service{
		Login: NewLogin(URL, requestLogger, responseLogger),
	}
}
