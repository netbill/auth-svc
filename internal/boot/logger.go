package boot

import (
	"github.com/netbill/logium"
	"github.com/sirupsen/logrus"
)

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func NewLogger(config LogConfig) *logium.Entry {
	log := logium.New()

	lvl, err := logrus.ParseLevel(config.Level)
	if err != nil {
		lvl = logrus.InfoLevel
		log.WithField("bad_level", config.Level).Warn("unknown log level, fallback to info")
	}

	log.SetLevel(lvl)

	switch {
	case config.Format == "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return log.WithField("service", ServiceName)
}
