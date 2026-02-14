package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/pgdbx"
)

const accountsTable = "accounts"

const accountsColumns = "id, username, role, version, created_at, updated_at"

func scanAccount(row sq.RowScanner) (r repository.AccountRow, err error) {
	err = row.Scan(
		&r.ID,
		&r.Username,
		&r.Role,
		&r.Version,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.AccountRow{}, nil
	case err != nil:
		return repository.AccountRow{}, fmt.Errorf("scanning account: %w", err)
	}
	return r, nil
}

type accounts struct {
	db       *pgdbx.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountsQ(db *pgdbx.DB) repository.AccountsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return accounts{
		db:       db,
		selector: builder.Select(accountsTable).From(accountsTable),
		inserter: builder.Insert(accountsTable),
		updater:  builder.Update(accountsTable),
		deleter:  builder.Delete(accountsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountsTable),
	}
}

func (q accounts) New() repository.AccountsQ {
	return NewAccountsQ(q.db)
}

func (q accounts) Insert(ctx context.Context, input repository.AccountRow) (repository.AccountRow, error) {
	id := pgtype.UUID{Bytes: [16]byte(input.ID), Valid: true}

	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":       id,
		"username": pgtype.Text{String: input.Username, Valid: true},
		"role":     pgtype.Text{String: input.Role, Valid: true},
	}).Suffix("RETURNING " + accountsTable + ".*").ToSql()
	if err != nil {
		return repository.AccountRow{}, fmt.Errorf("building insert query for %s: %w", accountsTable, err)
	}

	return scanAccount(q.db.QueryRow(ctx, query, args...))
}

func (q accounts) Get(ctx context.Context) (repository.AccountRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.AccountRow{}, fmt.Errorf("building get query for %s: %w", accountsTable, err)
	}

	return scanAccount(q.db.QueryRow(ctx, query, args...))
}

func (q accounts) UpdateOne(ctx context.Context) (repository.AccountRow, error) {
	q.updater = q.updater.
		Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}).
		Set("version", sq.Expr("version + 1"))

	query, args, err := q.updater.
		Suffix("RETURNING " + accountsColumns).
		ToSql()
	if err != nil {
		return repository.AccountRow{}, fmt.Errorf("building update query for %s: %w", accountsTable, err)
	}

	return scanAccount(q.db.QueryRow(ctx, query, args...))
}

func (q accounts) UpdateRole(role string) repository.AccountsQ {
	q.updater = q.updater.Set("role", pgtype.Text{String: role, Valid: true})
	return q
}

func (q accounts) UpdateUsername(username string) repository.AccountsQ {
	q.updater = q.updater.Set("username", pgtype.Text{String: username, Valid: true})
	return q
}

func (q accounts) Select(ctx context.Context) ([]repository.AccountRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]repository.AccountRow, 0)
	for rows.Next() {
		r, err := scanAccount(rows)
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

func (q accounts) Exists(ctx context.Context) (bool, error) {
	subSQL, subArgs, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return false, err
	}

	sql := "SELECT EXISTS (" + subSQL + ")"

	var exists bool
	err = q.db.QueryRow(ctx, sql, subArgs...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("sql=%s args=%v: %w", sql, subArgs, err)
	}

	return exists, nil
}

func (q accounts) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q accounts) FilterID(id uuid.UUID) repository.AccountsQ {
	pid := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"id": pid})
	q.counter = q.counter.Where(sq.Eq{"id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"id": pid})
	q.updater = q.updater.Where(sq.Eq{"id": pid})
	return q
}

func (q accounts) FilterRole(role string) repository.AccountsQ {
	val := pgtype.Text{String: role, Valid: true}

	q.selector = q.selector.Where(sq.Eq{"role": val})
	q.counter = q.counter.Where(sq.Eq{"role": val})
	q.deleter = q.deleter.Where(sq.Eq{"role": val})
	q.updater = q.updater.Where(sq.Eq{"role": val})
	return q
}

func (q accounts) FilterUsername(username string) repository.AccountsQ {
	val := pgtype.Text{String: username, Valid: true}

	q.selector = q.selector.Where(sq.Eq{"username": val})
	q.counter = q.counter.Where(sq.Eq{"username": val})
	q.deleter = q.deleter.Where(sq.Eq{"username": val})
	q.updater = q.updater.Where(sq.Eq{"username": val})
	return q
}

func (q accounts) FilterEmail(email string) repository.AccountsQ {
	em := pgtype.Text{String: email, Valid: true}

	q.selector = q.selector.
		Join("account_emails ae ON ae.account_id = accounts.id").
		Where(sq.Eq{"ae.email": em})

	q.counter = q.counter.
		Join("account_emails ae ON ae.account_id = accounts.id").
		Where(sq.Eq{"ae.email": em})

	sub := sq.Select("account_id").
		From("account_emails").
		Where(sq.Eq{"email": em})

	q.updater = q.updater.Where(sq.Expr("id IN (?)", sub))
	q.deleter = q.deleter.Where(sq.Expr("id IN (?)", sub))

	return q
}

func (q accounts) FilterVersion(version int32) repository.AccountsQ {
	q.selector = q.selector.Where(sq.Eq{"version": version})
	q.counter = q.counter.Where(sq.Eq{"version": version})
	q.deleter = q.deleter.Where(sq.Eq{"version": version})
	q.updater = q.updater.Where(sq.Eq{"version": version})
	return q
}

func (q accounts) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountsTable, err)
	}

	var count int64
	err = q.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for %s: %d", accountsTable, count)
	}

	return uint(count), nil
}

func (q accounts) Page(limit, offset uint) repository.AccountsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
