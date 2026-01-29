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

const sessionsTable = "sessions"

func scanSession(row sq.RowScanner) (r repository.SessionRow, err error) {
	err = row.Scan(
		&r.ID,
		&r.AccountID,
		&r.HashToken,
		&r.LastUsed,
		&r.CreatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.SessionRow{}, nil
	case err != nil:
		return repository.SessionRow{}, fmt.Errorf("scanning session: %w", err)
	}

	return r, nil
}

type sessions struct {
	db       *pgdbx.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewSessionsQ(db *pgdbx.DB) repository.SessionsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return sessions{
		db:       db,
		selector: builder.Select(sessionsTable + ".*").From(sessionsTable),
		inserter: builder.Insert(sessionsTable),
		updater:  builder.Update(sessionsTable),
		deleter:  builder.Delete(sessionsTable),
		counter:  builder.Select("COUNT(*) AS count").From(sessionsTable),
	}
}

func (q sessions) New() repository.SessionsQ {
	return NewSessionsQ(q.db)
}

func (q sessions) Insert(ctx context.Context, input repository.SessionRow) (repository.SessionRow, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":         pgtype.UUID{Bytes: input.ID, Valid: true},
		"account_id": pgtype.UUID{Bytes: input.AccountID, Valid: true},
		"hash_token": pgtype.Text{String: input.HashToken, Valid: true},
	}).Suffix("RETURNING id, account_id, hash_token, last_used, created_at").ToSql()
	if err != nil {
		return repository.SessionRow{}, fmt.Errorf("building insert query for %s: %w", sessionsTable, err)
	}

	return scanSession(q.db.QueryRow(ctx, query, args...))
}

func (q sessions) Update(ctx context.Context) ([]repository.SessionRow, error) {
	q.updater = q.updater.
		Set("last_used", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}).
		Suffix("RETURNING " + sessionsTable + ".*")

	query, args, err := q.updater.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building update query for %s: %w", sessionsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []repository.SessionRow
	for rows.Next() {
		r, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning updated session: %w", err)
		}
		out = append(out, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q sessions) UpdateToken(token string) repository.SessionsQ {
	q.updater = q.updater.Set("hash_token", pgtype.Text{String: token, Valid: true})
	return q
}

func (q sessions) UpdateLastUsed(lastUsed time.Time) repository.SessionsQ {
	q.updater = q.updater.Set("last_used", pgtype.Timestamptz{Time: lastUsed.UTC(), Valid: true})
	return q
}

func (q sessions) Get(ctx context.Context) (repository.SessionRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.SessionRow{}, fmt.Errorf("building get query for sessions: %w", err)
	}

	return scanSession(q.db.QueryRow(ctx, query, args...))
}

func (q sessions) GetHashToken(ctx context.Context) (string, error) {
	query, args, err := q.selector.
		Columns("hash_token").
		Limit(1).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("building get hash token query for sessions: %w", err)
	}

	var token string
	err = q.db.QueryRow(ctx, query, args...).Scan(&token)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("scanning hash token for sessions: %w", err)
	}

	return token, nil
}

func (q sessions) Select(ctx context.Context) ([]repository.SessionRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for sessions: %w", err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []repository.SessionRow
	for rows.Next() {
		r, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}
		out = append(out, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q sessions) Exists(ctx context.Context) (bool, error) {
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

func (q sessions) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for sessions: %w", err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q sessions) FilterID(ID uuid.UUID) repository.SessionsQ {
	pid := pgtype.UUID{Bytes: [16]byte(ID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"id": pid})
	q.updater = q.updater.Where(sq.Eq{"id": pid})
	q.counter = q.counter.Where(sq.Eq{"id": pid})

	return q
}

func (q sessions) FilterAccountID(accountID uuid.UUID) repository.SessionsQ {
	pid := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"account_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": pid})
	q.updater = q.updater.Where(sq.Eq{"account_id": pid})
	q.counter = q.counter.Where(sq.Eq{"account_id": pid})

	return q
}

func (q sessions) OrderCreatedAt(ascending bool) repository.SessionsQ {
	if ascending {
		q.selector = q.selector.OrderBy("created_at ASC")
	} else {
		q.selector = q.selector.OrderBy("created_at DESC")
	}
	return q
}

func (q sessions) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for sessions: %w", err)
	}

	var count int64
	err = q.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for sessions: %d", count)
	}

	return uint(count), nil
}

func (q sessions) Page(limit, offset uint) repository.SessionsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
