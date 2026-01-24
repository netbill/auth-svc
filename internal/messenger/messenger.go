package messenger

import (
	"database/sql"

	"github.com/netbill/logium"
)

type Messenger struct {
	addr []string
	db   *sql.DB
	log  *logium.Logger
}

func New(
	log *logium.Logger,
	db *sql.DB,
	addr ...string,
) Messenger {
	return Messenger{
		addr: addr,
		db:   db,
		log:  log,
	}
}
