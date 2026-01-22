package outbound

import (
	"database/sql"

	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
)

type Producer struct {
	log    logium.Logger
	outbox outbox.Box
}

func New(log logium.Logger, db *sql.DB) *Producer {
	return &Producer{
		log:    log,
		outbox: outbox.New(db),
	}
}
