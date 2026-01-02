package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/netbill/auth-svc/internal/domain/errx"
	"github.com/netbill/restkit/roles"
)

func TestUserRegistration(t *testing.T) {
	s, err := newSetup(t)
	if err != nil {
		t.Fatalf("newSetup: %v", err)
	}

	cleanDb(t)

	ctx := context.Background()

	firstEmail := "tests@example"
	password := "Test@1234"

	_, err = s.core.Auth.Register(ctx,
		firstEmail,
		password,
		roles.User,
	)
	if err != nil {
		t.Fatalf("Registration: %v", err)
	}

	_, err = s.core.Auth.Login(ctx, firstEmail, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	userFirst, err := s.core.User.GetByEmail(ctx, firstEmail)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}

	res, err := s.core.Session.ListForUser(ctx, userFirst.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListMySessions: %v", err)
	}
	if res.Total != 1 || len(res.Data) != 1 {
		t.Fatalf("ListMySessions: expected 1 session, got %d", res.Total)
	}

	err = s.core.Session.DeleteAllForUser(ctx, userFirst.ID)
	if err != nil {
		t.Fatalf("DeleteMySessions: %v", err)
	}

	res, err = s.core.Session.ListForUser(ctx, userFirst.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListMySessions after delete: %v", err)
	}
	if res.Total != 0 || len(res.Data) != 0 {
		t.Fatalf("ListMySessions after delete: expected 0 sessions, got %d", res.Total)
	}

	_, err = s.core.Auth.Login(ctx, firstEmail, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	_, err = s.core.Auth.Login(ctx, firstEmail, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	res, err = s.core.Session.ListForUser(ctx, userFirst.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListMySessions: %v", err)
	}
	if res.Total != 2 || len(res.Data) != 2 {
		t.Fatalf("ListMySessions: expected 2 session, got %d", res.Total)
	}
}

func TestUpdateUserPassword(t *testing.T) {
	s, err := newSetup(t)
	if err != nil {
		t.Fatalf("newSetup: %v", err)
	}

	cleanDb(t)

	ctx := context.Background()

	firstEmail := "tests@example"
	password := "Test@1234"

	_, err = s.core.Auth.Register(ctx,
		firstEmail,
		password,
		roles.User,
	)
	if err != nil {
		t.Fatalf("Registration: %v", err)
	}

	_, err = s.core.Auth.Login(ctx, firstEmail, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	userFirst, err := s.core.User.GetByEmail(ctx, firstEmail)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}

	res, err := s.core.Session.ListForUser(ctx, userFirst.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListMySessions: %v", err)
	}
	if res.Total != 1 || len(res.Data) != 1 {
		t.Fatalf("ListMySessions: expected 1 session, got %d", res.Total)
	}

	newPassword := "Test2@1234"

	err = s.core.Auth.UpdatePassword(ctx, userFirst.ID, password, newPassword)
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	_, err = s.core.Auth.Login(ctx, firstEmail, password)
	if !errors.Is(err, errx.ErrorInvalidLogin) {
		t.Fatalf("Login with old password: expected error %v, got %v", errx.ErrorInvalidLogin, err)
	}

	_, err = s.core.Auth.Login(ctx, firstEmail, newPassword)
	if err != nil {
		t.Fatalf("Login with new password: %v", err)
	}
}
