package messenger

import (
	"github.com/netbill/eventbox"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/pgdbx"
	"github.com/segmentio/kafka-go"
)

func NewProducer(database *pgdbx.DB, addr ...string) eventbox.Producer {
	return eventpg.NewProducer(
		&kafka.Writer{
			Addr:         kafka.TCP(addr...),
			RequiredAcks: kafka.RequireAll,
			Compression:  kafka.Snappy,
			Balancer:     &kafka.LeastBytes{},
		},
		database,
	)
}
