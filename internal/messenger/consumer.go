package messenger

import (
	"time"

	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

type ConsumerConfig struct {
	GroupID string   `json:"group_id"`
	Brokers []string `json:"brokers"`

	MinBackoff time.Duration `json:"min_backoff"`
	MaxBackoff time.Duration `json:"max_backoff"`

	OrganizationsV1 ConsumeKafkaConfig `json:"organizations_v1"`
	OrgMembersV1    ConsumeKafkaConfig `json:"organization_members_v1"`
}

type ConsumeKafkaConfig struct {
	Instances     int           `json:"instances"`
	MinBytes      int           `json:"min_bytes"`
	MaxBytes      int           `json:"max_bytes"`
	MaxWait       time.Duration `json:"max_wait"`
	QueueCapacity int           `json:"queue_capacity"`
}

func NewConsumer(
	logger *log.Logger,
	inbox eventbox.Inbox,
	config ConsumerConfig,
) *eventbox.Consumer {
	consumer := eventbox.NewConsumer(logger, inbox, eventbox.ConsumerConfig{
		MinBackoff: config.MinBackoff,
		MaxBackoff: config.MaxBackoff,
	})

	consumer.AddReader(eventbox.ReaderConfig{
		Brokers:       config.Brokers,
		GroupID:       config.GroupID,
		Topic:         evtypes.OrganizationsTopicV1,
		Instances:     config.OrganizationsV1.Instances,
		MaxWait:       config.OrganizationsV1.MaxWait,
		MinBytes:      config.OrganizationsV1.MinBytes,
		MaxBytes:      config.OrganizationsV1.MaxBytes,
		QueueCapacity: config.OrganizationsV1.QueueCapacity,
	})

	consumer.AddReader(eventbox.ReaderConfig{
		Brokers:       config.Brokers,
		GroupID:       config.GroupID,
		Topic:         evtypes.OrgMembersTopicV1,
		Instances:     config.OrgMembersV1.Instances,
		MaxWait:       config.OrgMembersV1.MaxWait,
		MinBytes:      config.OrgMembersV1.MinBytes,
		MaxBytes:      config.OrgMembersV1.MaxBytes,
		QueueCapacity: config.OrgMembersV1.QueueCapacity,
	})

	return consumer
}
