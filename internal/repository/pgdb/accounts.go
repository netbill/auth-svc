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

const accountsTable = "accounts"

const accountsColumns = "id, role, username, created_at, updated_at"
const accountsColumnsA = "a.id, a.role, a.username, a.created_at, a.updated_at"

type Account struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountsQ(db pgx.DBTX) AccountsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return AccountsQ{
		db:       db,
		selector: builder.Select("accounts.*").From(accountsTable),
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
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":       input.ID,
		"username": input.Username,
		"role":     input.Role,
	}).Suffix("RETURNING accounts.*").ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building insert query for %s: %w", accountsTable, err)
	}

	var out Account
	row := q.db.QueryRowContext(ctx, query, args...)
	if err = out.scan(row); err != nil {
		return Account{}, err
	}
	return out, nil
}

func (q AccountsQ) Get(ctx context.Context) (Account, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building get query for %s: %w", accountsTable, err)
	}

	row := q.db.QueryRowContext(ctx, query, args...)

	var a Account
	err = a.scan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, nil
		}
		return Account{}, err
	}

	return a, nil
}

func (q AccountsQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", accountsTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", accountsTable, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", accountsTable, err)
	}

	return affected, nil
}

func (q AccountsQ) UpdateOne(ctx context.Context) (Account, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.
		Suffix("RETURNING " + accountsColumns).
		ToSql()
	if err != nil {
		return Account{}, fmt.Errorf("building update query for %s: %w", accountsTable, err)
	}

	var updated Account
	if err = updated.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return Account{}, err
	}

	return updated, nil
}

func (q AccountsQ) UpdateRole(role string) AccountsQ {
	q.updater = q.updater.Set("role", role)
	return q
}

func (q AccountsQ) UpdateUsername(username string) AccountsQ {
	q.updater = q.updater.Set("username", username)
	return q
}

func (q AccountsQ) Select(ctx context.Context) ([]Account, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
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

	return out, nil
}

func (q AccountsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	return err
}

func (q AccountsQ) FilterID(id uuid.UUID) AccountsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	return q
}

func (q AccountsQ) FilterRole(role string) AccountsQ {
	q.selector = q.selector.Where(sq.Eq{"role": role})
	q.counter = q.counter.Where(sq.Eq{"role": role})
	q.deleter = q.deleter.Where(sq.Eq{"role": role})
	q.updater = q.updater.Where(sq.Eq{"role": role})
	return q
}

func (q AccountsQ) FilterUsername(username string) AccountsQ {
	q.selector = q.selector.Where(sq.Eq{"username": username})
	q.counter = q.counter.Where(sq.Eq{"username": username})
	q.deleter = q.deleter.Where(sq.Eq{"username": username})
	q.updater = q.updater.Where(sq.Eq{"username": username})
	return q
}

func (q AccountsQ) FilterEmail(email string) AccountsQ {
	q.selector = q.selector.
		Join("account_emails ae ON ae.account_id = accounts.id").
		Where(sq.Eq{"ae.email": email})

	q.counter = q.counter.
		Join("account_emails ae ON ae.account_id = accounts.id").
		Where(sq.Eq{"ae.email": email})

	sub := sq.Select("account_id").
		From("account_emails").
		Where(sq.Eq{"email": email})

	q.updater = q.updater.Where(sq.Expr("id IN (?)", sub))
	q.deleter = q.deleter.Where(sq.Expr("id IN (?)", sub))

	return q
}

func (q AccountsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountsTable, err)
	}

	var count uint
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q AccountsQ) Page(limit, offset uint) AccountsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
