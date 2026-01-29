package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/pgdbx"
)

const OrganizationMemberTable = "organization_members"

const OrganizationMemberColumns = "id, account_id, organization_id, source_created_at, replica_created_at"
const OrganizationMemberColumnsM = "m.id, m.account_id, m.organization_id, m.source_created_at, m.replica_created_at"

func scanOrganizationMember(row sq.RowScanner) (r repository.OrganizationMemberRow, err error) {
	err = row.Scan(
		&r.ID,
		&r.AccountID,
		&r.OrganizationID,
		&r.SourceCreatedAt,
		&r.ReplicaCreatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.OrganizationMemberRow{}, nil
	case err != nil:
		return repository.OrganizationMemberRow{}, fmt.Errorf("scanning organization member: %w", err)
	}

	return r, nil
}

type organizationMembers struct {
	db       *pgdbx.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrganizationMembersQ(db *pgdbx.DB) repository.OrganizationMembersQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return organizationMembers{
		db:       db,
		selector: builder.Select(OrganizationMemberColumnsM).From(OrganizationMemberTable + " m"),
		inserter: builder.Insert(OrganizationMemberTable),
		deleter:  builder.Delete(OrganizationMemberTable + " m"),
		counter:  builder.Select("COUNT(*)").From(OrganizationMemberTable + " m"),
	}
}

func (q organizationMembers) New() repository.OrganizationMembersQ {
	return NewOrganizationMembersQ(q.db)
}

func (q organizationMembers) Insert(ctx context.Context, data repository.OrganizationMemberRow) (repository.OrganizationMemberRow, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":                pgtype.UUID{Bytes: data.ID, Valid: true},
		"account_id":        pgtype.UUID{Bytes: data.AccountID, Valid: true},
		"organization_id":   pgtype.UUID{Bytes: data.OrganizationID, Valid: true},
		"source_created_at": pgtype.Timestamptz{Time: data.SourceCreatedAt.UTC(), Valid: true},
	}).Suffix("RETURNING " + OrganizationMemberColumns).ToSql()
	if err != nil {
		return repository.OrganizationMemberRow{}, fmt.Errorf("building insert query for %s: %w", OrganizationMemberTable, err)
	}

	return scanOrganizationMember(q.db.QueryRow(ctx, query, args...))
}

func (q organizationMembers) Get(ctx context.Context) (repository.OrganizationMemberRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.OrganizationMemberRow{}, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	return scanOrganizationMember(q.db.QueryRow(ctx, query, args...))
}

func (q organizationMembers) Select(ctx context.Context) ([]repository.OrganizationMemberRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationMemberTable, err)
	}
	defer rows.Close()

	var out []repository.OrganizationMemberRow
	for rows.Next() {
		r, err := scanOrganizationMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q organizationMembers) Exists(ctx context.Context) (bool, error) {
	query, args, err := q.selector.
		Columns("1").
		Limit(1).
		ToSql()
	if err != nil {
		return false, err
	}

	var one int
	err = q.db.QueryRow(ctx, query, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (q organizationMembers) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationMemberTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationMemberTable, err)
	}

	return nil
}

func (q organizationMembers) FilterByID(id uuid.UUID) repository.OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.id": pid})
	return q
}

func (q organizationMembers) FilterByAccountID(accountID uuid.UUID) repository.OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.account_id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.account_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.account_id": pid})
	return q
}

func (q organizationMembers) FilterByOrganizationID(organizationID uuid.UUID) repository.OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(organizationID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.organization_id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.organization_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.organization_id": pid})
	return q
}

func (q organizationMembers) Page(limit, offset uint) repository.OrganizationMembersQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

func (q organizationMembers) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationMemberTable, err)
	}

	var count int64
	if err = q.db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationMemberTable, err)
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for %s: %d", OrganizationMemberTable, count)
	}

	return uint(count), nil
}
