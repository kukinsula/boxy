package redis

import (
	"context"
	_ "fmt"
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

func (client *Client) Subscribe(
	ctx context.Context,
	channel Channel,
	ping time.Duration) *Subscription {

	subscription := NewSusbcription(ctx, channel, ping)
	conn := client.pool.Get()

	go subscription.Start(conn)

	return subscription
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

func (client *Client) Request(req *Request) *Response {
	start := time.Now()
	failure := make(chan error)
	conn := client.pool.Get()
	ctx, cancel := context.WithCancel(req.Context)
	resp := &Response{Request: req, codec: client.codec, logger: client.logger}
	subscription := NewSusbcription(ctx, Channel(req.UUID), req.Ping)

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

	client.logger(req.UUID, log.DEBUG, "REDIS request finished",
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

type Handler interface {
	Params() interface{}
	Exec(uuid string, ctx context.Context) (interface{}, error)
}

type HandlerBuilder func() Handler

func (client *Client) Handle(channel Channel, builder HandlerBuilder) (err error) {
	conn := client.pool.Get()

	for err == nil {
		values, err := redis.Values(conn.Do("BLPOP", string(channel), 0))
		if err != nil {
			break
		}

		go func() {
			var tmp string
			var raw []byte

			// Scan the BLOP received request
			_, err = redis.Scan(values, &tmp, &raw)
			if err != nil {
				return
			}

			// Extract the reponse ID from raw payload
			requestId, err := entity.FromBytes(raw[:responseIdLen])
			if err != nil {
				return
			}

			handler := builder()
			params := handler.Params()

			// Decode request parameters
			err = client.codec.Decode(raw[responseIdLen:], params)

			client.logger(requestId, log.DEBUG,
				"REDIS BLOP received a request to handle",
				map[string]interface{}{
					"channel": channel,
					"params":  params,
					"error":   err,
				})

			if err != nil {
				return
			}

			// Execute the handler
			var body interface{}
			result, err := handler.Exec(requestId, context.Background())
			if err == nil {
				body = result
			} else {
				body = err
			}

			// Send the response
			err = client.Publish(Channel(string(requestId)), body)

			client.logger(requestId, log.DEBUG,
				"REDIS PUBLISH response",
				map[string]interface{}{
					"channel": channel,
					"result":  result,
					"error":   err,
				})

			if err != nil {
				return
			}
		}()
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
