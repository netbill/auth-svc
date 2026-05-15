package bus

import (
	"context"
)

type Broker struct {
	pub *Publisher
	sub *Subscriber
}

func NewBroker(pub *Publisher, sub *Subscriber) *Broker {
	return &Broker{pub: pub, sub: sub}
}

func qrTokenKey(token string) string {
	return "qr-token:" + token
}

func (b *Broker) PublishQRToken(ctx context.Context, key string, payload []byte) error {
	return b.pub.Publish(ctx, qrTokenKey(key), payload)
}

func (b *Broker) SubscribeQRToken(ctx context.Context, key string) (<-chan []byte, func()) {
	return b.sub.Subscribe(ctx, qrTokenKey(key))
}
