package messenger

import (
	"context"
	"sync"
	"time"

	"github.com/netbill/auth-svc/internal/messenger/contracts"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
	"github.com/segmentio/kafka-go"
)

type ConsumerArchitect struct {
	log          *logium.Entry
	db           *pgdbx.DB
	brokers      []string
	topicReaders map[string]int
}

func NewConsumerArchitect(
	log *logium.Logger,
	db *pgdbx.DB,
	brokers []string,
	topicReaders map[string]int,
) *ConsumerArchitect {
	return &ConsumerArchitect{
		log:          log.WithField("component", "kafka-consumer"),
		db:           db,
		brokers:      brokers,
		topicReaders: topicReaders,
	}
}

func (a *ConsumerArchitect) Start(ctx context.Context) {
	var wg sync.WaitGroup

	accountReadersNum, ok := a.topicReaders[contracts.OrgMemberTopicV1]
	if !ok || accountReadersNum <= 0 {
		a.log.Fatalf("number of readers for topic %s must be greater than 0", contracts.OrgMemberTopicV1)
	}

	accountConsumer := eventpg.NewConsumer(a.log, a.db, eventpg.ConsumerConfig{
		MinBackoff: 200 * time.Millisecond,
		MaxBackoff: 5 * time.Second,
	})

	a.log.Infoln("starting kafka consumers process")

	wg.Add(accountReadersNum)

	for i := 0; i < accountReadersNum; i++ {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  a.brokers,
			GroupID:  contracts.AuthSvcGroup,
			Topic:    contracts.OrgMemberTopicV1,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		})

		go func(r *kafka.Reader) {
			defer wg.Done()
			accountConsumer.Read(ctx, r) // Read сам закроет reader
		}(reader)
	}

	wg.Wait()
}
