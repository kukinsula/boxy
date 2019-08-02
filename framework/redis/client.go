package redis

import (
	"context"
	"time"

	"github.com/kukinsula/boxy/entity"

	"github.com/gomodule/redigo/redis"
)

const responseIdLen = 16 // bytes

type Config struct {
	Address         string        `yaml:"address"`
	MaxIdle         int           `yaml:"max-idle"`
	MaxActive       int           `yaml:"max-active"`
	IdleTimeout     time.Duration `yaml:"idle-timeout"`
	MaxConnLifetime time.Duration `yaml:"max-conn-lifetime"`
	Codec           Codec
	Logger          entity.Logger
}

type Client struct {
	pool   *redis.Pool
	codec  Codec
	logger entity.Logger
}

type Request struct {
	UUID    string
	Context context.Context
	Channel Channel
	Ping    time.Duration
	Params  interface{}
	Result  interface{}
}

type Handler struct {
	Channel Channel
	Params  interface{}
	Handle  func(uuid string, ctx context.Context) (interface{}, error)
}

// TODO: put into entity package
type Codec interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
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

func (client *Client) Publish(channel Channel, data []byte) error {
	conn := client.pool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", string(channel), data)

	return err
}

func (client *Client) Request(req Request) (err error) {
	failure := make(chan error)
	ctx, cancel := context.WithCancel(req.Context)
	conn := client.pool.Get()
	subscription := NewSusbcription(req.UUID, ctx, Channel(req.UUID), req.Ping, client.logger)

	go func() {
		failure <- subscription.Start(conn)
		close(failure)
	}()

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-failure:

		case <-subscription.Subscribed:
			subscription.logger(entity.Log{
				UUID:    subscription.UUID,
				Level:   "debug",
				Message: "REDIS subscribed",
				Meta:    map[string]interface{}{"channel": subscription.channel},
			})

			err = client.request(req)

		case data := <-subscription.Message:
			err = client.codec.Decode(data, req.Result)

			client.logger(entity.Log{
				UUID:    req.UUID,
				Level:   "debug",
				Message: "REDIS <- request received a response",
				Meta: map[string]interface{}{
					"channel": req.Channel,
					"result":  req.Result,
					"error":   err,
				},
			})

			cancel() // Cancels the subscription

			goOn = false
		}
	}

	return err
}

func (client *Client) request(req Request) error {
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

	payload := append(uuid, data...)

	_, err = conn.Do("RPUSH", string(req.Channel), payload)

	client.logger(entity.Log{
		UUID:    req.UUID,
		Level:   "debug",
		Message: "REDIS RPUSH",
		Meta: map[string]interface{}{
			"channel": req.Channel,
			"params":  req.Params,
			"error":   err,
		},
	})

	return err
}

func (client *Client) Handle(handler Handler) (err error) {
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

		// Decode request parameters
		err = client.codec.Decode(raw[responseIdLen:], handler.Params)

		client.logger(entity.Log{
			UUID:    requestId,
			Level:   "debug",
			Message: "REDIS BLOP received a request to handle",
			Meta: map[string]interface{}{
				"channel": handler.Channel,
				"params":  handler.Params,
				"error":   err,
			},
		})

		if err != nil {
			break
		}

		// Execute the handler
		result, err := handler.Handle(requestId, context.Background())
		if err != nil {
			break
		}

		// Encode the result of execution
		data, err := client.codec.Encode(result)
		if err != nil {
			break
		}

		// Send the response
		err = client.Publish(Channel(string(requestId)), data)

		client.logger(entity.Log{
			UUID:    requestId,
			Level:   "debug",
			Message: "REDIS PUBLISH response",
			Meta: map[string]interface{}{
				"channel": handler.Channel,
				"result":  result,
				"error":   err,
			},
		})

		if err != nil {
			break
		}
	}

	conn.Close()

	return err
}
