package client

type Service struct {
	Login     *Login
	Streaming *Streaming
}

func NewService(
	URL string,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *Service {

	return &Service{
		Login:     NewLogin(URL, requestLogger, responseLogger),
		Streaming: NewStreaming(URL, requestLogger, responseLogger),
	}
}
