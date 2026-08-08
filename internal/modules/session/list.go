package session

import (
	"context"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

const (
	defaultLimit = uint(20)
	maxLimit     = uint(100)
)

type DeletedFilter uint8

const (
	DeletedFilterAll     DeletedFilter = iota // default: active + deleted
	DeletedFilterActive                       // only active
	DeletedFilterDeleted                      // only deleted
)

type LastUsedOrder uint8

const (
	LastUsedDesc LastUsedOrder = iota // default
	LastUsedAsc
)

type ListSessionsOption func(*ListSessionsOptions)

type ListSessionsOptions struct {
	Deleted  DeletedFilter
	LastUsed LastUsedOrder
	Limit    uint
	Offset   uint
}

func ApplyListOptions(optFns []ListSessionsOption) ListSessionsOptions {
	opts := ListSessionsOptions{
		Limit:  defaultLimit,
		Offset: 0,
	}
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

func WithDeleted(f DeletedFilter) ListSessionsOption {
	return func(o *ListSessionsOptions) {
		o.Deleted = f
	}
}

func WithLastUsedOrder(order LastUsedOrder) ListSessionsOption {
	return func(o *ListSessionsOptions) {
		o.LastUsed = order
	}
}

func WithLimit(limit uint) ListSessionsOption {
	return func(o *ListSessionsOptions) {
		if limit == 0 || limit > maxLimit {
			limit = maxLimit
		}
		o.Limit = limit
	}
}

func WithOffset(offset uint) ListSessionsOption {
	return func(o *ListSessionsOptions) {
		o.Offset = offset
	}
}

func (s *Service) GetMySessions(
	ctx context.Context,
	actor models.UserActor,
	opts ...ListSessionsOption,
) (pagi.Page[[]models.Session], error) {
	return s.sessionRepo.GetListForUser(ctx, actor.ID, opts...)
}
