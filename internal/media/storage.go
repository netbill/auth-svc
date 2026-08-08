package media

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/netbill/awsx"
)

type Config struct {
	LinkTTL    time.Duration
	UserAvatar awsx.ImageValidator
}

type Uploader struct {
	s3     awsx.Bucket
	config Config
}

func NewStorage(s3 awsx.Bucket, config Config) *Uploader {
	return &Uploader{
		s3:     s3,
		config: config,
	}
}
