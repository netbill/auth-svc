package reponses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Account(a models.Account) *pb.Account {
	out := &pb.Account{
		Id:        a.ID.String(),
		Role:      a.Role,
		Version:   a.Version,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}
	if a.DeletedAt != nil {
		out.DeletedAt = timestamppb.New(*a.DeletedAt)
	}
	return out
}

func AccountEmail(e models.AccountEmail) *pb.AccountEmail {
	out := &pb.AccountEmail{
		AccountId: e.AccountID.String(),
		Email:     e.Email,
		Verified:  e.Verified,
		Version:   e.Version,
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}
	if e.DeletedAt != nil {
		out.DeletedAt = timestamppb.New(*e.DeletedAt)
	}
	return out
}

func Session(s models.Session) *pb.Session {
	out := &pb.Session{
		Id:        s.ID.String(),
		AccountId: s.AccountID.String(),
		Version:   s.Version,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
		LastUsed:  timestamppb.New(s.LastUsed),
	}
	if s.DeletedAt != nil {
		out.DeletedAt = timestamppb.New(*s.DeletedAt)
	}
	return out
}

func TokensPair(t models.TokensPair) *pb.TokensPair {
	return &pb.TokensPair{
		SessionId:    t.SessionID.String(),
		AccessToken:  t.Access,
		RefreshToken: t.Refresh,
	}
}
