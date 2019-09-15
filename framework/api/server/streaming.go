package server

import (
	"context"
	"fmt"
	"net/http"

	loginEntity "github.com/kukinsula/boxy/entity/login"

	"github.com/gin-gonic/gin"
)

type Streamer struct {
	UUID    string
	Context *gin.Context
	flusher http.Flusher
	User    *loginEntity.User
	Receive chan []byte
}

func NewStreamer(
	uuid string,
	ctx *gin.Context,
	flusher http.Flusher,
	user *loginEntity.User) *Streamer {

	return &Streamer{
		UUID:    uuid,
		Context: ctx,
		flusher: flusher,
		User:    user,
		Receive: make(chan []byte),
	}
}

func (streamer *Streamer) Run() error {
	for data := range streamer.Receive {
		_, err := fmt.Fprintf(streamer.Context.Writer,
			"data: %s\n\n", data)

		if err != nil {
			return err
		}

		streamer.flusher.Flush()
	}

	return nil
}

type StreamingSet struct {
	streamers map[string]*Streamer
	add       chan *Streamer
	remove    chan string
	send      chan []byte
	close     chan struct{}
}

func NewStreamingSet() *StreamingSet {
	set := &StreamingSet{
		streamers: make(map[string]*Streamer),
		add:       make(chan *Streamer),
		remove:    make(chan string),
		send:      make(chan []byte),
		close:     make(chan struct{}),
	}

	go set.monitoring()

	return set
}

func (set *StreamingSet) monitoring() {
	for goOn := true; goOn; {
		select {
		case streamer := <-set.add:
			set.streamers[streamer.UUID] = streamer

		case uuid := <-set.remove:
			streamer, ok := set.streamers[uuid]
			if !ok {
				break
			}

			delete(set.streamers, uuid)
			close(streamer.Receive)

		case data := <-set.send:
			for _, streamer := range set.streamers {
				streamer.Receive <- data
			}

		case <-set.close:
			goOn = false
		}
	}
}

func (set *StreamingSet) Add(streamer *Streamer) {
	set.add <- streamer
}

func (set *StreamingSet) Remove(streamer *Streamer) {
	set.remove <- streamer.UUID
}

func (set *StreamingSet) Send(data []byte) {
	set.send <- data
}

func (set *StreamingSet) Close() {
	set.close <- struct{}{}
}

func Streaming(
	ctx context.Context,
	streaming StreamingBackender) gin.HandlerFunc {

	subscription := streaming.Subscribe(ctx)
	set := NewStreamingSet()

	<-subscription.Subscribed

	go func() {
		for data := range subscription.Message {
			set.Send(data)
		}
	}()

	return func(ctx *gin.Context) {
		flusher, ok := ctx.Writer.(http.Flusher)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Streaming unsuported"})
			return
		}

		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")

		uuid := getRequestUUID(ctx)
		streamer := NewStreamer(uuid, ctx, flusher, nil)

		set.Add(streamer)
		streamer.Run()
		set.Remove(streamer)
	}
}
