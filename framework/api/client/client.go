package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/kukinsula/boxy/entity"
)

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	DELETE Method = "DELETE"
)

const (
	X_REQUEST_ID = "X-Request-ID"
)

type Headers map[string][]string

type Set map[string]interface{}

type Request struct {
	UUID    string
	URL     string
	Method  Method
	Path    string
	Headers Headers
	Query   Set
	codec   Codec
	Body    interface{}
	body    io.ReadCloser
}

type Response struct {
	Request  *Request
	Status   int
	Headers  Headers
	Error    error
	codec    Codec
	body     io.ReadCloser
	Duration time.Duration
}

// TODO: put into entity package (used by redis too)
type Codec interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, result interface{}) error
}

type Requester interface {
	Request(req *Request) (*Response, error)
}

type RequestLogger func(req *Request)
type ResponseLogger func(resp *Response)

type client struct {
	requester      Requester
	codec          Codec
	URL            string
	requestLogger  RequestLogger
	responseLogger ResponseLogger
}

func newClient(
	url string,
	requester Requester,
	codec Codec,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *client {

	return &client{
		URL:            url,
		requester:      requester,
		codec:          codec,
		requestLogger:  requestLogger,
		responseLogger: responseLogger,
	}
}

func (client *client) request(req *Request) *Response {
	start := time.Now()

	req.codec = client.codec

	if req.UUID == "" {
		req.UUID = entity.NewUUID()
	}

	if req.Headers == nil {
		req.Headers = map[string][]string{}
	}

	if req.URL == "" {
		req.URL = client.URL
	}

	if len(req.Query) != 0 {
		req.Path = fmt.Sprintf("%s?", req.Path)

		for name, value := range req.Query {
			req.Path = fmt.Sprintf("%s&%s=%v", req.Path, name, value)
		}
	}

	req.URL = fmt.Sprintf("%s%s", req.URL, req.Path)

	req.Headers[X_REQUEST_ID] = []string{req.UUID}

	client.requestLogger(req)

	resp, err := client.requester.Request(req)
	if err != nil {
		resp = &Response{Error: err}
	}

	resp.Request = req
	resp.Duration = time.Since(start)

	client.responseLogger(resp)

	resp.codec = client.codec

	return resp
}

func (client *client) GET(req *Request) *Response {
	req.Method = GET

	return client.request(req)
}

func (client *client) POST(req *Request) *Response {
	req.Method = POST

	return client.request(req)
}

func (client *client) PUT(req *Request) *Response {
	req.Method = PUT

	return client.request(req)
}

func (client *client) DELETE(req *Request) *Response {
	req.Method = DELETE

	return client.request(req)
}

func (resp *Response) Decode(data interface{}) (*Response, error) {
	if resp.Error != nil {
		return nil, resp.Error
	}

	body, err := ioutil.ReadAll(resp.body)
	if err != nil {
		return nil, err
	}

	err = resp.codec.Decode(body, data)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
