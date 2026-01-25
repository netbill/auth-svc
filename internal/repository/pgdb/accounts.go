package pgdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/pgxtx"
)

const accountsTable = "accounts"

const accountsColumns = "id, username, role, created_at, updated_at"
const accountsColumnsA = "a.id, a.username, a.role, a.created_at, a.updated_at"

type Account struct {
	ID        pgtype.UUID        `db:"id"`
	Username  pgtype.Text        `db:"username"`
	Role      pgtype.Text        `db:"role"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`
}

func (a *Account) scan(row sq.RowScanner) error {
	err := row.Scan(
		&a.ID,
		&a.Username,
		&a.Role,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning account: %w", err)
	}
	return nil
}

type AccountsQ struct {
	db       pgxtx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountsQ(db pgxtx.DBTX) AccountsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return AccountsQ{
		db:       db,
		selector: builder.Select(accountsTable + ".*").From(accountsTable),
		inserter: builder.Insert(accountsTable),
		updater:  builder.Update(accountsTable),
		deleter:  builder.Delete(accountsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountsTable),
	}
}

type InsertAccountParams struct {
	ID       uuid.UUID
	Username string
	Role     string
}

func (q AccountsQ) Insert(ctx context.Context, input InsertAccountParams) (Account, error) {
	id := pgtype.UUID{Bytes: [16]byte(input.ID), Valid: true}

	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":       id,
		"username": pgtype.Text{String: input.Username, Valid: true},
		"role":     pgtype.Text{String: input.Role, Valid: true},
	}).Suffix("RETURNING " + accountsTable + ".*").ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building insert query for %s: %w", accountsTable, err)
	}

	var out Account
	if err = out.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		return Account{}, err
	}
	return out, nil
}

func (q AccountsQ) Get(ctx context.Context) (Account, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building get query for %s: %w", accountsTable, err)
	}

	var a Account
	err = a.scan(q.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, nil
		}
		return Account{}, err
	}

	return a, nil
}

func (q AccountsQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", accountsTable, err)
	}

	tag, err := q.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", accountsTable, err)
	}

	return tag.RowsAffected(), nil
}

func (q AccountsQ) UpdateOne(ctx context.Context) (Account, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.
		Suffix("RETURNING " + accountsColumns).
		ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building update query for %s: %w", accountsTable, err)
	}

	var updated Account
	if err = updated.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		return Account{}, err
	}

	return updated, nil
}

func (q AccountsQ) UpdateRole(role string) AccountsQ {
	q.updater = q.updater.Set("role", pgtype.Text{String: role, Valid: true})
	return q
}

func (q AccountsQ) UpdateUsername(username string) AccountsQ {
	q.updater = q.updater.Set("username", pgtype.Text{String: username, Valid: true})
	return q
}

func (q AccountsQ) Select(ctx context.Context) ([]Account, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Account, 0)
	for rows.Next() {
		var a Account
		err = a.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning account: %w", err)
		}
		out = append(out, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q AccountsQ) Exists(ctx context.Context) (bool, error) {
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

func (q AccountsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q AccountsQ) FilterID(id uuid.UUID) AccountsQ {
	pid := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"id": pid})
	q.counter = q.counter.Where(sq.Eq{"id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"id": pid})
	q.updater = q.updater.Where(sq.Eq{"id": pid})
	return q
}

func (q AccountsQ) FilterRole(role string) AccountsQ {
	val := pgtype.Text{String: role, Valid: true}

	q.selector = q.selector.Where(sq.Eq{"role": val})
	q.counter = q.counter.Where(sq.Eq{"role": val})
	q.deleter = q.deleter.Where(sq.Eq{"role": val})
	q.updater = q.updater.Where(sq.Eq{"role": val})
	return q
}

func (q AccountsQ) FilterUsername(username string) AccountsQ {
	val := pgtype.Text{String: username, Valid: true}

	q.selector = q.selector.Where(sq.Eq{"username": val})
	q.counter = q.counter.Where(sq.Eq{"username": val})
	q.deleter = q.deleter.Where(sq.Eq{"username": val})
	q.updater = q.updater.Where(sq.Eq{"username": val})
	return q
}

func (q AccountsQ) FilterEmail(email string) AccountsQ {
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

func (q AccountsQ) Count(ctx context.Context) (uint, error) {
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

func (q AccountsQ) Page(limit, offset uint) AccountsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
