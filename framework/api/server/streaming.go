package server

import (
	loginUsecase "github.com/kukinsula/boxy/usecase/login"

	"github.com/gin-gonic/gin"
)

type Streamer struct {
	UUID    string
	Context *gin.Context
}

type StreamingSet struct {
	streamers map[string]*Streamer
	add       chan *Streamer
	remove    chan string
	close     chan struct{}
}

func NewStreamingSet() *StreamingSet {
	set := &StreamingSet{
		streamers: make(map[string]*Streamer),
		add:       make(chan *Streamer),
		remove:    make(chan string),
		close:     make(chan struct{}),
	}

	go set.monitoring()

	return set
}

func (set *StreamingSet) monitoring() {
	for goOn := true; goOn; {
		select {
		case streamer := <-set.add:
			set.streamer[streamer.UUID] = streamer

		case uuid <- set.remove:
			delete(set.streamer, uuid)

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

func Metrics(monitoring MonitoringBackender) gin.HandlerFunc {
	channel := monitoring.Subscribe(ctx)
	set := NewStreamingSet()

	return func(ctx *gin.Context) {
		uuid := getRequestUUID(ctx)
		streamer := &Streamer{UUID: uuid, Context: ctx}

		set.Add(streamer)

		ctx.JSON(200, result)
	}
}
