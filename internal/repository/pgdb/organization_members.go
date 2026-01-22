package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/netbill/pgx"
)

const OrganizationMemberTable = "organization_members"

const OrganizationMemberColumns = "id, account_id, organization_id, source_created_at, replica_created_at"
const OrganizationMemberColumnsM = "m.id, m.account_id, m.organization_id, m.source_created_at, m.replica_created_at"

type OrganizationMember struct {
	ID             uuid.UUID `json:"id"`
	AccountID      uuid.UUID `json:"account_id"`
	OrganizationID uuid.UUID `json:"organization_id"`

	SourceCreatedAt  time.Time `json:"source_created_at"`
	ReplicaCreatedAt time.Time `json:"replica_created_at"`
}

func (m *OrganizationMember) scan(row sq.RowScanner) error {
	err := row.Scan(
		&m.ID,
		&m.AccountID,
		&m.OrganizationID,
		&m.SourceCreatedAt,
		&m.ReplicaCreatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning organization member: %w", err)
	}
	return nil
}

type OrganizationMembersQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrganizationMembersQ(db pgx.DBTX) OrganizationMembersQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrganizationMembersQ{
		db:       db,
		selector: builder.Select(OrganizationMemberColumnsM).From(OrganizationMemberTable + " m"),
		inserter: builder.Insert(OrganizationMemberTable),
		deleter:  builder.Delete(OrganizationMemberTable + " m"),
		counter:  builder.Select("COUNT(*)").From(OrganizationMemberTable + " m"),
	}
}

type OrganizationMemberInsertInput struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	OrganizationID uuid.UUID

	SourceCreatedAt time.Time
}

func (q OrganizationMembersQ) Insert(ctx context.Context, data OrganizationMemberInsertInput) (OrganizationMember, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":                 data.ID,
		"account_id":         data.AccountID,
		"organization_id":    data.OrganizationID,
		"source_created_at":  data.SourceCreatedAt.UTC(),
		"replica_created_at": time.Now().UTC(),
	}).Suffix("RETURNING " + OrganizationMemberColumns).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building insert query for %s: %w", OrganizationMemberTable, err)
	}

	var inserted OrganizationMember
	if err = inserted.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationMember{}, err
	}
	return inserted, nil
}

func (q OrganizationMembersQ) Get(ctx context.Context) (OrganizationMember, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	var m OrganizationMember
	if err = m.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationMember{}, nil
		default:
			return OrganizationMember{}, err
		}
	}
	return m, nil
}

func (q OrganizationMembersQ) Select(ctx context.Context) ([]OrganizationMember, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationMemberTable, err)
	}
	defer rows.Close()

	var out []OrganizationMember
	for rows.Next() {
		var m OrganizationMember
		if err = m.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrganizationMembersQ) Exists(ctx context.Context) (bool, error) {
	query, args, err := q.selector.
		Columns("1").
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("building exists query for %s: %w", OrganizationMemberTable, err)
	}

	var one int
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&one)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		default:
			return false, fmt.Errorf("executing exists query for %s: %w", OrganizationMemberTable, err)
		}
	}

	return true, nil
}

func (q OrganizationMembersQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationMemberTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationMemberTable, err)
	}

	return nil
}

func (q OrganizationMembersQ) FilterByID(id uuid.UUID) OrganizationMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.id": id})
	q.counter = q.counter.Where(sq.Eq{"m.id": id})
	q.deleter = q.deleter.Where(sq.Eq{"m.id": id})
	return q
}

func (q OrganizationMembersQ) FilterByAccountID(accountID uuid.UUID) OrganizationMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"m.account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"m.account_id": accountID})
	return q
}

func (q OrganizationMembersQ) FilterByOrganizationID(organizationID uuid.UUID) OrganizationMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.organization_id": organizationID})
	q.counter = q.counter.Where(sq.Eq{"m.organization_id": organizationID})
	q.deleter = q.deleter.Where(sq.Eq{"m.organization_id": organizationID})
	return q
}

func (q OrganizationMembersQ) Page(limit, offset uint) OrganizationMembersQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

func (q OrganizationMembersQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationMemberTable, err)
	}

	var count uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationMemberTable, err)
	}

	return count, nil
}
