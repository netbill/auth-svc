package messenger

import (
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

type Messenger struct {
	addr []string
	db   *pgdbx.DB
	log  *logium.Logger
}

func New(
	log *logium.Logger,
	db *pgdbx.DB,
	addr ...string,
) *Messenger {
	return &Messenger{
		addr: addr,
		db:   db,
		log:  log,
	}
}
