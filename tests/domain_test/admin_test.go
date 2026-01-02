package domain_test

import (
	"context"
	"testing"

	"github.com/netbill/restkit/roles"
)

func TestAdminBlockUser(t *testing.T) {
	s, err := newSetup(t)
	if err != nil {
		t.Fatalf("newSetup: %v", err)
	}

	cleanDb(t)

	ctx := context.Background()

	admin := CreateUser(s, t, "admin@example", "Admin@1234", roles.Admin)
	_ = CreateSession(s, t, admin.ID)
	_ = CreateSession(s, t, admin.ID)
	_ = CreateSession(s, t, admin.ID)
	user := CreateUser(s, t, "user@example", "User@1234", roles.User)
	_ = CreateSession(s, t, user.ID)
	_ = CreateSession(s, t, user.ID)
	_ = CreateSession(s, t, user.ID)

	sess, err := s.core.Session.ListForUser(ctx, user.ID, 0, 100)
	if err != nil {
		t.Fatalf("ListMySessions: unexpected error: %v", err)
	}
	if len(sess.Data) != 3 {
		t.Fatalf("ListMySessions: expected 3 sessions, got %d", len(sess.Data))
	}

	_, err = s.core.User.BlockUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("BlockUser: unexpected error: %v", err)
	}

	sess, err = s.core.Session.ListForUser(ctx, user.ID, 0, 100)
	if err != nil {
		t.Fatalf("ListMySessions: unexpected error: %v", err)
	}
	if len(sess.Data) != 0 {
		t.Fatalf("ListMySessions: expected 0 sessions, got %d", len(sess.Data))
	}

	sess, err = s.core.Session.ListForUser(ctx, admin.ID, 0, 100)
	if err != nil {
		t.Fatalf("ListMySessions: unexpected error: %v", err)
	}
	if len(sess.Data) != 3 {
		t.Fatalf("ListMySessions: expected 3 sessions, got %d", len(sess.Data))
	}
}
