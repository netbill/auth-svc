package messenger

import (
	"time"

	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

type ProducerConfig struct {
	Producer string   `json:"producer"`
	Brokers  []string `json:"brokers"`

	AccountV1 ProduceKafkaConfig `json:"account_v1"`
}

type ProduceKafkaConfig struct {
	RequiredAcks string        `json:"required_acks"`
	Compression  string        `json:"compression"`
	Balancer     string        `json:"balancer"`
	BatchSize    int           `json:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout"`
}

func NewProducer(log *log.Logger, cfg ProducerConfig) *eventbox.Producer {
	producer := eventbox.NewProducer(log, cfg.Brokers...)

	err := producer.AddWriter(evtypes.AccountsTopicV1, eventbox.WriterTopicConfig{
		RequiredAcks: cfg.AccountV1.RequiredAcks,
		Compression:  cfg.AccountV1.Compression,
		Balancer:     cfg.AccountV1.Balancer,
		BatchSize:    cfg.AccountV1.BatchSize,
		BatchTimeout: cfg.AccountV1.BatchTimeout,
	})
	if err != nil {
		panic(err)
	}

	return producer
}
