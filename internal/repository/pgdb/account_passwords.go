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

const accountPasswordsTable = "account_passwords"

type AccountPassword struct {
	AccountID uuid.UUID `db:"account_id"`
	Hash      string    `db:"hash"`
	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (a *AccountPassword) scan(row sq.RowScanner) error {
	err := row.Scan(
		&a.AccountID,
		&a.Hash,
		&a.UpdatedAt,
		&a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning account password: %w", err)
	}
	return nil
}

type AccountPasswordsQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountPasswordsQ(db pgx.DBTX) AccountPasswordsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return AccountPasswordsQ{
		db:       db,
		selector: builder.Select("account_passwords.*").From(accountPasswordsTable),
		inserter: builder.Insert(accountPasswordsTable),
		updater:  builder.Update(accountPasswordsTable),
		deleter:  builder.Delete(accountPasswordsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountPasswordsTable),
	}
}

func (q AccountPasswordsQ) Insert(ctx context.Context, input AccountPassword) error {
	values := map[string]interface{}{
		"account_id": input.AccountID,
		"hash":       input.Hash,
		"updated_at": input.UpdatedAt,
		"created_at": input.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building insert query for %s: %w", accountPasswordsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	return err
}

func (q AccountPasswordsQ) Update(
	ctx context.Context,
) ([]AccountPassword, error) {
	q.updater = q.updater.
		Set("updated_at", time.Now().UTC()).
		Suffix("RETURNING account_passwords.*")

	query, args, err := q.updater.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building update query for %s: %w", accountPasswordsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AccountPassword
	for rows.Next() {
		var p AccountPassword
		err = p.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning updated account password: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

func (q AccountPasswordsQ) UpdateHash(hash string) AccountPasswordsQ {
	q.updater = q.updater.Set("hash", hash)
	return q
}

func (q AccountPasswordsQ) Get(ctx context.Context) (AccountPassword, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return AccountPassword{}, fmt.Errorf("building get query for %s: %w", accountPasswordsTable, err)
	}

	row := q.db.QueryRowContext(ctx, query, args...)

	var p AccountPassword
	err = p.scan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountPassword{}, nil
		}
		return AccountPassword{}, err
	}

	return p, nil
}

func (q AccountPasswordsQ) Select(ctx context.Context) ([]AccountPassword, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountPasswordsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AccountPassword
	for rows.Next() {
		var p AccountPassword
		err = p.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning account_password: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

func (q AccountPasswordsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountPasswordsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	return err
}

func (q AccountPasswordsQ) FilterAccountID(accountID uuid.UUID) AccountPasswordsQ {
	q.selector = q.selector.Where(sq.Eq{"account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"account_id": accountID})
	return q
}

func (q AccountPasswordsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountPasswordsTable, err)
	}

	var count uint
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q AccountPasswordsQ) Page(limit, offset uint) AccountPasswordsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
