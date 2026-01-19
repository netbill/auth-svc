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

const accountEmailsTable = "account_emails"

const accountEmailsColumns = "account_id, email, verified, created_at, updated_at"

type AccountEmail struct {
	AccountID uuid.UUID `db:"account_id"`
	Email     string    `db:"email"`
	Verified  bool      `db:"verified"`
	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (e *AccountEmail) scan(row sq.RowScanner) error {
	err := row.Scan(
		&e.AccountID,
		&e.Email,
		&e.Verified,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning account email: %w", err)
	}
	return nil
}

type AccountEmailsQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountEmailsQ(db pgx.DBTX) AccountEmailsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return AccountEmailsQ{
		db:       db,
		selector: builder.Select("account_emails.*").From(accountEmailsTable),
		inserter: builder.Insert(accountEmailsTable),
		updater:  builder.Update(accountEmailsTable),
		deleter:  builder.Delete(accountEmailsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountEmailsTable),
	}
}

func (q AccountEmailsQ) Insert(ctx context.Context, input AccountEmail) error {
	values := map[string]interface{}{
		"account_id": input.AccountID,
		"email":      input.Email,
		"verified":   input.Verified,
		"updated_at": input.UpdatedAt,
		"created_at": input.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building insert query for %s: %w", accountEmailsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)

	return err
}

func (q AccountEmailsQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", accountEmailsTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", accountEmailsTable, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", accountEmailsTable, err)
	}

	return affected, nil
}

func (q AccountEmailsQ) UpdateOne(ctx context.Context) (AccountEmail, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.
		Suffix("RETURNING " + accountEmailsColumns).
		ToSql()
	if err != nil {
		return AccountEmail{}, fmt.Errorf("building update query for %s: %w", accountEmailsTable, err)
	}

	var updated AccountEmail
	if err = updated.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return AccountEmail{}, err
	}

	return updated, nil
}

func (q AccountEmailsQ) UpdateEmail(email string) AccountEmailsQ {
	q.updater = q.updater.Set("email", email)
	return q
}

func (q AccountEmailsQ) UpdateVerified(verified bool) AccountEmailsQ {
	q.updater = q.updater.Set("verified", verified)
	return q
}

func (q AccountEmailsQ) Get(ctx context.Context) (AccountEmail, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return AccountEmail{}, fmt.Errorf("building get query for %s: %w", accountEmailsTable, err)
	}

	row := q.db.QueryRowContext(ctx, query, args...)

	var e AccountEmail
	err = e.scan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountEmail{}, nil
		}
		return AccountEmail{}, err
	}

	return e, nil
}

func (q AccountEmailsQ) Select(ctx context.Context) ([]AccountEmail, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountEmailsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AccountEmail
	for rows.Next() {
		var e AccountEmail
		err = e.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning account_email: %w", err)
		}
		out = append(out, e)
	}

	return out, nil
}

func (q AccountEmailsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountEmailsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	return err
}

func (q AccountEmailsQ) FilterAccountID(accountID uuid.UUID) AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"account_id": accountID})
	return q
}

func (q AccountEmailsQ) FilterEmail(email string) AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"email": email})
	q.counter = q.counter.Where(sq.Eq{"email": email})
	q.deleter = q.deleter.Where(sq.Eq{"email": email})
	q.updater = q.updater.Where(sq.Eq{"email": email})
	return q
}

func (q AccountEmailsQ) FilterVerified(verified bool) AccountEmailsQ {
	q.selector = q.selector.Where(sq.Eq{"verified": verified})
	q.counter = q.counter.Where(sq.Eq{"verified": verified})
	q.deleter = q.deleter.Where(sq.Eq{"verified": verified})
	q.updater = q.updater.Where(sq.Eq{"verified": verified})
	return q
}

func (q AccountEmailsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountEmailsTable, err)
	}

	var count uint
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q AccountEmailsQ) Page(limit, offset uint) AccountEmailsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
