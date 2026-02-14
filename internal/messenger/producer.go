package messenger

import (
	"github.com/netbill/eventbox"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/segmentio/kafka-go"
)

func (m *Manager) NewProducer() eventbox.Producer {
	w := m.buildWriter()
	return eventpg.NewProducer(w, m.db)
}

func (m *Manager) buildWriter() *kafka.Writer {
	cfg := m.config.Writer

	w := &kafka.Writer{
		Addr:         kafka.TCP(m.config.Brokers...),
		RequiredAcks: parseRequiredAcks(cfg.RequiredAcks),
		Compression:  parseCompression(cfg.Compression),
		Balancer:     parseBalancer(cfg.Balancer),
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
	}

	if cfg.DialTimeout > 0 || cfg.IdleTimeout > 0 {
		w.Transport = &kafka.Transport{
			DialTimeout: cfg.DialTimeout,
			IdleTimeout: cfg.IdleTimeout,
		}
	}

	return w
}

func parseRequiredAcks(v string) kafka.RequiredAcks {
	switch v {
	case "none":
		return kafka.RequireNone
	case "one":
		return kafka.RequireOne
	default:
		return kafka.RequireAll
	}
}

func parseCompression(v string) kafka.Compression {
	switch v {
	case "gzip":
		return kafka.Gzip
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	case "none":
		return 0
	default:
		return kafka.Snappy
	}
}

func parseBalancer(v string) kafka.Balancer {
	switch v {
	case "round_robin":
		return &kafka.RoundRobin{}
	case "hash":
		return &kafka.Hash{}
	default:
		return &kafka.LeastBytes{}
	}
}
