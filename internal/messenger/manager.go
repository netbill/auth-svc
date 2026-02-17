package messenger

import (
	"fmt"
	"os"
	"time"

	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

const AuthSvcGroup = "auth-svc"

type TopicReaderConfig struct {
	NumReaders     int           `mapstructure:"num_readers" validate:"required"`
	MinBytes       int           `mapstructure:"min_bytes"`
	MaxBytes       int           `mapstructure:"max_bytes"`
	MaxWait        time.Duration `mapstructure:"max_wait"`
	CommitInterval time.Duration `mapstructure:"commit_interval"`
	StartOffset    string        `mapstructure:"start_offset"`
	QueueCapacity  int           `mapstructure:"queue_capacity"`
}

type Config struct {
	Brokers []string `mapstructure:"brokers" validate:"required"`
	Writer  struct {
		RequiredAcks string        `mapstructure:"required_acks"`
		Compression  string        `mapstructure:"compression"`
		Balancer     string        `mapstructure:"balancer"`
		BatchSize    int           `mapstructure:"batch_size"`
		BatchTimeout time.Duration `mapstructure:"batch_timeout"`
		DialTimeout  time.Duration `mapstructure:"dial_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	}
	Reader struct {
		Topics struct {
			OrgMembersV1    TopicReaderConfig `mapstructure:"organization_members_v1"`
			OrganizationsV1 TopicReaderConfig `mapstructure:"organizations_v1"`
		} `mapstructure:"topics"`
	} `mapstructure:"reader"`
	Inbox struct {
		Routines       int           `mapstructure:"routines"`
		Slots          int           `mapstructure:"slots"`
		BatchSize      int           `mapstructure:"batch_size"`
		Sleep          time.Duration `mapstructure:"sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"inbox"`
	Outbox struct {
		Routines       int           `mapstructure:"routines"`
		Slots          int           `mapstructure:"slots"`
		BatchSize      int           `mapstructure:"batch_size"`
		Sleep          time.Duration `mapstructure:"sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"outbox"`
}

type Manager struct {
	log *logium.Entry
	db  *pgdbx.DB

	config Config
}

func NewManager(log *logium.Entry, db *pgdbx.DB, config Config) *Manager {
	return &Manager{
		log:    log,
		db:     db,
		config: config,
	}
}

func BuildProcessID(service string) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s-%s-%d", service, hostname, os.Getpid())
}
