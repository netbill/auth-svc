package bus

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client *redis.Client
}

func NewSubscriber(client *redis.Client) *Subscriber {
	return &Subscriber{
		client: client,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, key string) (<-chan []byte, func()) {
	sub := s.client.Subscribe(ctx, key)

	ch := make(chan []byte, 1)

	go func() {
		defer close(ch)
		for msg := range sub.Channel() {
			ch <- []byte(msg.Payload)
		}
	}()

	cleanup := func() {
		sub.Close()
	}

	return ch, cleanup
}
