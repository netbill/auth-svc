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

const accountPasswordsTable = "account_passwords"

const accountPasswordsColumns = "account_id, hash, created_at, updated_at"

func scanAccountPassword(row sq.RowScanner) (r repository.AccountPasswordRow, err error) {
	err = row.Scan(
		&r.AccountID,
		&r.Hash,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.AccountPasswordRow{}, nil
	case err != nil:
		return repository.AccountPasswordRow{}, fmt.Errorf("scanning account_password: %w", err)
	}

	return r, nil
}

type accountPasswords struct {
	db       *pgdbx.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountPasswordsQ(db *pgdbx.DB) repository.AccountPasswordsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return accountPasswords{
		db:       db,
		selector: builder.Select(accountPasswordsTable + ".*").From(accountPasswordsTable),
		inserter: builder.Insert(accountPasswordsTable),
		updater:  builder.Update(accountPasswordsTable),
		deleter:  builder.Delete(accountPasswordsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountPasswordsTable),
	}
}

func (q accountPasswords) New() repository.AccountPasswordsQ {
	return NewAccountPasswordsQ(q.db)
}

func (q accountPasswords) Insert(ctx context.Context, input repository.AccountPasswordRow) (repository.AccountPasswordRow, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"account_id": input.AccountID,
		"hash":       input.Hash,
	}).Suffix("RETURNING " + accountPasswordsColumns).ToSql()
	if err != nil {
		return repository.AccountPasswordRow{}, fmt.Errorf("building insert query for %s: %w", accountPasswordsTable, err)
	}

	return scanAccountPassword(q.db.QueryRow(ctx, query, args...))
}

func (q accountPasswords) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", accountPasswordsTable, err)
	}

	tag, err := q.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", accountPasswordsTable, err)
	}

	return tag.RowsAffected(), nil
}

func (q accountPasswords) UpdateOne(ctx context.Context) (repository.AccountPasswordRow, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.
		Suffix("RETURNING " + accountPasswordsColumns).
		ToSql()
	if err != nil {
		return repository.AccountPasswordRow{}, fmt.Errorf("building update query for %s: %w", accountPasswordsTable, err)
	}

	return scanAccountPassword(q.db.QueryRow(ctx, query, args...))
}

func (q accountPasswords) UpdateHash(hash string) repository.AccountPasswordsQ {
	q.updater = q.updater.Set("hash", pgtype.Text{String: hash, Valid: true})
	return q
}

func (q accountPasswords) Get(ctx context.Context) (repository.AccountPasswordRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.AccountPasswordRow{}, fmt.Errorf("building get query for %s: %w", accountPasswordsTable, err)
	}

	return scanAccountPassword(q.db.QueryRow(ctx, query, args...))
}

func (q accountPasswords) Select(ctx context.Context) ([]repository.AccountPasswordRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountPasswordsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []repository.AccountPasswordRow
	for rows.Next() {
		r, err := scanAccountPassword(rows)
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

func (q accountPasswords) Exists(ctx context.Context) (bool, error) {
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

func (q accountPasswords) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountPasswordsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q accountPasswords) FilterAccountID(accountID uuid.UUID) repository.AccountPasswordsQ {
	id := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"account_id": id})
	q.counter = q.counter.Where(sq.Eq{"account_id": id})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": id})
	q.updater = q.updater.Where(sq.Eq{"account_id": id})
	return q
}

func (q accountPasswords) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountPasswordsTable, err)
	}

	var count int64
	err = q.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for %s: %d", accountPasswordsTable, count)
	}

	return uint(count), nil
}

func (q accountPasswords) Page(limit, offset uint) repository.AccountPasswordsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
