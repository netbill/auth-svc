package messenger

import (
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

type Messenger struct {
	addr []string
	pool *pgdbx.DB
	log  *logium.Logger
}

func New(
	log *logium.Logger,
	pool *pgdbx.DB,
	addr ...string,
) Messenger {
	return Messenger{
		addr: addr,
		pool: pool,
		log:  log,
	}
}
