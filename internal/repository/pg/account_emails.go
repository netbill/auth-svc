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

const (
	accountEmailsTable = "account_emails"

	accountEmailsPrefix = "ae.account_id, ae.email, ae.verified, ae.version, ae.created_at, ae.updated_at"
	accountEmailsReturn = "account_id, email, verified, version, created_at, updated_at"
)

func scanAccountEmail(row sq.RowScanner) (r repository.AccountEmailRow, err error) {
	err = row.Scan(
		&r.AccountID,
		&r.Email,
		&r.Verified,
		&r.Version,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.AccountEmailRow{}, nil
	case err != nil:
		return repository.AccountEmailRow{}, fmt.Errorf("scanning account_email: %w", err)
	}
	return r, nil
}

type accountEmails struct {
	db       *pgdbx.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountEmailsQ(db *pgdbx.DB) repository.AccountEmailsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return accountEmails{
		db:       db,
		selector: builder.Select(accountEmailsPrefix).From(accountEmailsTable + " ae"),
		inserter: builder.Insert(accountEmailsTable),
		updater:  builder.Update(accountEmailsTable),
		deleter:  builder.Delete(accountEmailsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountEmailsTable + " ae"),
	}
}

func (q accountEmails) New() repository.AccountEmailsQ {
	return NewAccountEmailsQ(q.db)
}

func (q accountEmails) Insert(ctx context.Context, input repository.AccountEmailRow) (repository.AccountEmailRow, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"account_id": input.AccountID,
		"email":      input.Email,
		"verified":   input.Verified,
	}).Suffix("RETURNING " + accountEmailsReturn).ToSql()
	if err != nil {
		return repository.AccountEmailRow{}, fmt.Errorf("building insert query for %s: %w", accountEmailsTable, err)
	}

	return scanAccountEmail(q.db.QueryRow(ctx, query, args...))
}

func (q accountEmails) UpdateOne(ctx context.Context) (repository.AccountEmailRow, error) {
	q.updater = q.updater.
		Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}).
		Set("version", sq.Expr("version + 1"))

	query, args, err := q.updater.Suffix("RETURNING " + accountEmailsReturn).ToSql()
	if err != nil {
		return repository.AccountEmailRow{}, fmt.Errorf("building update query for %s: %w", accountEmailsTable, err)
	}

	return scanAccountEmail(q.db.QueryRow(ctx, query, args...))
}

func (q accountEmails) UpdateEmail(email string) repository.AccountEmailsQ {
	q.updater = q.updater.Set("email", pgtype.Text{String: email, Valid: true})
	return q
}

func (q accountEmails) UpdateVerified(verified bool) repository.AccountEmailsQ {
	q.updater = q.updater.Set("verified", pgtype.Bool{Bool: verified, Valid: true})
	return q
}

func (q accountEmails) Get(ctx context.Context) (repository.AccountEmailRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.AccountEmailRow{}, fmt.Errorf("building get query for %s: %w", accountEmailsTable, err)
	}
	return scanAccountEmail(q.db.QueryRow(ctx, query, args...))
}

func (q accountEmails) Select(ctx context.Context) ([]repository.AccountEmailRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountEmailsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]repository.AccountEmailRow, 0)
	for rows.Next() {
		r, err := scanAccountEmail(rows)
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

func (q accountEmails) Exists(ctx context.Context) (bool, error) {
	subSQL, subArgs, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return false, err
	}
	sql := "SELECT EXISTS (" + subSQL + ")"

	var exists bool
	if err = q.db.QueryRow(ctx, sql, subArgs...).Scan(&exists); err != nil {
		return false, fmt.Errorf("sql=%s args=%v: %w", sql, subArgs, err)
	}
	return exists, nil
}

func (q accountEmails) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountEmailsTable, err)
	}
	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q accountEmails) FilterAccountID(accountID uuid.UUID) repository.AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"ae.account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"ae.account_id": accountID})

	q.deleter = q.deleter.Where(sq.Eq{"account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"account_id": accountID})
	return q
}

func (q accountEmails) FilterEmail(email string) repository.AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"ae.email": email})
	q.counter = q.counter.Where(sq.Eq{"ae.email": email})

	q.deleter = q.deleter.Where(sq.Eq{"email": email})
	q.updater = q.updater.Where(sq.Eq{"email": email})
	return q
}

func (q accountEmails) FilterVerified(verified bool) repository.AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"ae.verified": verified})
	q.counter = q.counter.Where(sq.Eq{"ae.verified": verified})

	q.deleter = q.deleter.Where(sq.Eq{"verified": verified})
	q.updater = q.updater.Where(sq.Eq{"verified": verified})
	return q
}

func (q accountEmails) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountEmailsTable, err)
	}
	var count uint
	if err = q.db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (q accountEmails) Page(limit, offset uint) repository.AccountEmailsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
