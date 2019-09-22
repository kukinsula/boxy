package client

import (
	"bufio"
	"fmt"
	"strings"
	// "sync"

	// codecEntity "github.com/kukinsula/boxy/entity/codec"
	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

// type Streamer struct {
// 	codec   codecEntity.Codec
// 	queue   chan []byte
// 	push    chan []byte
// 	next    chan interface{}
// 	close   chan struct{}
// 	errLock *sync.RWMutex
// 	err     error
// }

// func NewStreamer(codec codecEntity.Codec) *Streamer {
// 	streamer := &Streamer{
// 		codec:   codec,
// 		queue:   make(chan []byte),
// 		push:    make(chan []byte),
// 		next:    make(chan interface{}),
// 		close:   make(chan struct{}),
// 		errLock: &sync.RWMutex{},
// 	}

// 	go streamer.monitoring()

// 	return streamer
// }

// func (streamer *Streamer) monitoring() {
// 	var err error

// 	for goOn := true; err == nil && goOn; {
// 		select {
// 		case data := <-streamer.push:
// 			streamer.queue <- data

// 		case v := <-streamer.next:
// 			data := <-streamer.queue
// 			err = streamer.codec.Decode(data, v)

// 		case <-streamer.close:
// 			goOn = false
// 		}
// 	}

// 	if err != nil {
// 		streamer.errLock.Lock()
// 		streamer.err = err
// 		streamer.errLock.Unlock()
// 	}

// 	close(streamer.queue)
// 	close(streamer.push)
// 	close(streamer.next)
// 	close(streamer.close)
// }

// func (streamer *Streamer) readErr() error {
// 	var err error

// 	streamer.errLock.RLock()
// 	err = streamer.err
// 	streamer.errLock.RUnlock()

// 	return err
// }

// func (streamer *Streamer) Push(data []byte) error {
// 	err := streamer.readErr()
// 	if err != nil {
// 		return err
// 	}

// 	streamer.push <- data

// 	return nil
// }

// func (streamer *Streamer) Next(v interface{}) error {
// 	err := streamer.readErr()
// 	if err != nil {
// 		return err
// 	}

// 	streamer.next <- v

// 	return nil
// }

// func (streamer *Streamer) Close() error {
// 	err := streamer.readErr()
// 	if err != nil {
// 		return err
// 	}

// 	streamer.close <- struct{}{}

// 	return nil
// }

type Streaming struct {
	*client
}

func NewStreaming(
	URL string,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *Streaming {

	return &Streaming{
		client: newClient(
			URL,
			newRequester(),
			&JSONCodec{},
			requestLogger,
			responseLogger),
	}
}

func (streaming *Streaming) Stream(uuid, token string,
	channel chan *monitoringEntity.Metrics) error {

	resp := streaming.GET(&Request{
		UUID: uuid,
		Path: "/streaming",
		Headers: map[string][]string{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
			"Accept":        []string{"text/event-stream"},
		},
	})

	if resp.Status != 200 {
		return fmt.Errorf("Me should return Status code 200, not %d", resp.Status)
	}

	go func() {
		scanner := bufio.NewScanner(resp.body)
		for scanner.Scan() {
			text := scanner.Text()

			if text == "" {
				continue
			}

			if strings.HasPrefix(text, "data: ") {
				metrics := &monitoringEntity.Metrics{}
				err := streaming.codec.Decode([]byte(text[6:]), metrics)
				if err != nil {
					return
				}

				// fmt.Println(metrics)

				channel <- metrics
			}
		}

		close(channel)
		resp.body.Close()
	}()

	return nil
}
