package redis

import (
	"context"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/codec"
	"github.com/kukinsula/boxy/entity/log"

	"github.com/gomodule/redigo/redis"
)

const responseIdLen = 16 // bytes

type Config struct {
	Address         string        `yaml:"address"`
	MaxIdle         int           `yaml:"max-idle"`
	MaxActive       int           `yaml:"max-active"`
	IdleTimeout     time.Duration `yaml:"idle-timeout"`
	MaxConnLifetime time.Duration `yaml:"max-conn-lifetime"`
	Codec           codec.Codec
	Logger          log.Logger
}

type Client struct {
	pool   *redis.Pool
	codec  codec.Codec
	logger log.Logger
}

type Request struct {
	UUID    string
	Context context.Context
	Channel Channel
	Ping    time.Duration
	Params  interface{}
}

type Response struct {
	Request *Request
	data    []byte
	Error   error
	codec   codec.Codec
	logger  log.Logger
}

type Handler struct {
	Channel Channel
	Params  Reseter // interface{}
	Handle  func(uuid string, ctx context.Context) (interface{}, error)
}

type Reseter interface {
	Reset() error
}

func NewClient(config Config) (*Client, error) {
	client := &Client{
		pool: &redis.Pool{
			MaxActive:   config.MaxActive,
			MaxIdle:     config.MaxIdle,
			IdleTimeout: config.IdleTimeout,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", config.Address)
			},
		},
		codec:  config.Codec,
		logger: config.Logger,
	}

	return client, nil
}

func (client *Client) Close() error {
	return client.pool.Close()
}

func (client *Client) Ping() (string, error) {
	conn := client.pool.Get()
	defer conn.Close()

	return redis.String(conn.Do("PING"))
}

func (client *Client) Publish(channel Channel, v interface{}) error {
	data, err := client.codec.Encode(v)
	if err != nil {
		return err
	}

	conn := client.pool.Get()
	defer conn.Close()

	_, err = conn.Do("PUBLISH", string(channel), data)

	return err
}

func (client *Client) Request(req *Request) *Response {
	start := time.Now()
	failure := make(chan error)
	conn := client.pool.Get()
	ctx, cancel := context.WithCancel(req.Context)
	resp := &Response{Request: req, codec: client.codec, logger: client.logger}
	subscription := NewSusbcription(req.UUID, ctx, Channel(req.UUID), req.Ping)

	go func() {
		failure <- subscription.Start(conn)
		close(failure)
	}()

	var err error

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-failure:

		case <-subscription.Subscribed:
			err = client.sendRequest(req)

		case data := <-subscription.Message:
			resp.data = data

			// Cancels the subscription after the first message is received
			cancel()

			goOn = false
		}
	}

	client.logger(subscription.UUID, log.DEBUG,
		"REDIS request finished",
		map[string]interface{}{
			"channel":  subscription.channel,
			"duration": time.Since(start),
		})

	if err != nil {
		resp.Error = err
	}

	return resp
}

func (client *Client) sendRequest(req *Request) error {
	conn := client.pool.Get()
	defer conn.Close()

	uuid, err := entity.FromStringToBytes(req.UUID)
	if err != nil {
		return err
	}

	data, err := client.codec.Encode(req.Params)
	if err != nil {
		return err
	}

	_, err = conn.Do("RPUSH", string(req.Channel), append(uuid, data...))

	client.logger(req.UUID, log.DEBUG, "REDIS RPUSH",
		map[string]interface{}{
			"channel": req.Channel,
			"params":  req.Params,
			"error":   err,
		})

	return err
}

func (client *Client) Handle(handler *Handler) (err error) {
	conn := client.pool.Get()

	for err == nil {
		values, err := redis.Values(conn.Do("BLPOP", string(handler.Channel), 0))
		if err != nil {
			break
		}

		var tmp string
		var raw []byte

		// Scan the BLOP received request
		_, err = redis.Scan(values, &tmp, &raw)
		if err != nil {
			break
		}

		// Extract the reponse ID from raw payload
		requestId, err := entity.FromBytes(raw[:responseIdLen])
		if err != nil {
			break
		}

		err = handler.Params.Reset()
		if err != nil {
			break
		}

		// Decode request parameters
		err = client.codec.Decode(raw[responseIdLen:], handler.Params)

		client.logger(requestId, log.DEBUG,
			"REDIS BLOP received a request to handle",
			map[string]interface{}{
				"channel": handler.Channel,
				"params":  handler.Params,
				"error":   err,
			})

		if err != nil {
			break
		}

		// Execute the handler
		result, err := handler.Handle(requestId, context.Background())
		if err != nil {
			break
		}

		// Send the response
		err = client.Publish(Channel(string(requestId)), result)

		client.logger(requestId, log.DEBUG, "REDIS PUBLISH response",
			map[string]interface{}{
				"channel": handler.Channel,
				"result":  result,
				"error":   err,
			})

		if err != nil {
			break
		}
	}

	conn.Close()

	return err
}

func (resp *Response) Decode(result interface{}) error {
	if resp.Error != nil {
		return resp.Error
	}

	err := resp.codec.Decode(resp.data, result)

	resp.logger(resp.Request.UUID, log.DEBUG,
		"REDIS <- request received a response",
		map[string]interface{}{
			"channel": resp.Request.Channel,
			"result":  result,
			"error":   err,
		})

	return err
}
