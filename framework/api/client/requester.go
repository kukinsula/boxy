package client

import (
	"bytes"
	"net/http"
	"net/url"
)

type requester struct {
	client *http.Client
}

type buffer struct {
	*bytes.Buffer
}

func (buffer *buffer) Close() error {
	return nil
}

func newRequester() *requester {
	return &requester{client: &http.Client{}}
}

func (requester *requester) Request(req *Request) (*Response, error) {
	url, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}

	body, err := req.codec.Encode(req.Body)
	if err != nil {
		return nil, err
	}

	request := &http.Request{
		Method: string(req.Method),
		URL:    url,
		Header: http.Header(req.Headers),
		Body:   &buffer{Buffer: bytes.NewBuffer(body)},
	}

	response, err := requester.client.Do(request)
	if err != nil {
		return nil, err
	}

	return &Response{
		Status:  response.StatusCode,
		Headers: Headers(response.Header),
		body:    response.Body,
	}, nil
}
