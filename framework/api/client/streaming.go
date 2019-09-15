package client

import (
	"bufio"
	"fmt"
	"strings"

	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

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

func (streaming *Streaming) Stream(uuid, token string) (chan *monitoringEntity.Metrics, error) {
	resp := streaming.GET(&Request{
		UUID: uuid,
		Path: "/streaming",
		Headers: map[string][]string{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
			"Accept":        []string{"text/event-stream"},
		},
	})

	if resp.Status != 200 {
		return nil, fmt.Errorf("Me should return Status code 200, not %d", resp.Status)
	}

	channel := make(chan *monitoringEntity.Metrics)

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

				channel <- metrics
			}
		}

		resp.body.Close()
	}()

	return channel, nil
}
