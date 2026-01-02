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

const sessionsTable = "sessions"

type Session struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	HashToken string    `db:"hash_token"`
	LastUsed  time.Time `db:"last_used"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *Session) scan(row sq.RowScanner) error {
	err := row.Scan(
		&s.ID,
		&s.AccountID,
		&s.HashToken,
		&s.LastUsed,
		&s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning session: %w", err)
	}
	return nil
}

type SessionsQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewSessionsQ(db *sql.DB) SessionsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return SessionsQ{
		db:       db,
		selector: builder.Select("sessions.*").From(sessionsTable),
		inserter: builder.Insert(sessionsTable),
		updater:  builder.Update(sessionsTable),
		deleter:  builder.Delete(sessionsTable),
		counter:  builder.Select("COUNT(*) AS count").From(sessionsTable),
	}
}

func (q SessionsQ) Insert(ctx context.Context, input Session) error {
	values := map[string]interface{}{
		"id":         input.ID,
		"account_id": input.AccountID,
		"hash_token": input.HashToken,
		"last_used":  input.LastUsed,
		"created_at": input.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building insert query for %s: %w", sessionsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	return err
}

func (q SessionsQ) Update(ctx context.Context) ([]Session, error) {
	q.updater = q.updater.
		Set("last_used", time.Now().UTC()).
		Suffix("RETURNING sessions.*")

	query, args, err := q.updater.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building update query for %s: %w", sessionsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		err = s.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning updated session: %w", err)
		}
		out = append(out, s)
	}

	return out, nil
}

func (q SessionsQ) UpdateToken(token string) SessionsQ {
	q.updater = q.updater.Set("hash_token", token)
	return q
}

func (q SessionsQ) UpdateLastUsed(lastUsed time.Time) SessionsQ {
	q.updater = q.updater.Set("last_used", lastUsed)
	return q
}

func (q SessionsQ) Get(ctx context.Context) (Session, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Session{}, fmt.Errorf("building get query for sessions: %w", err)
	}

	var sess Session
	row := q.db.QueryRowContext(ctx, query, args...)
	err = sess.scan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, nil
		}

		return Session{}, err
	}

	return sess, nil
}

func (q SessionsQ) Select(ctx context.Context) ([]Session, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for sessions: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		err = sess.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (q SessionsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for sessions: %w", err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)

	return err
}

func (q SessionsQ) FilterID(ID uuid.UUID) SessionsQ {
	q.selector = q.selector.Where(sq.Eq{"id": ID})
	q.deleter = q.deleter.Where(sq.Eq{"id": ID})
	q.updater = q.updater.Where(sq.Eq{"id": ID})
	q.counter = q.counter.Where(sq.Eq{"id": ID})

	return q
}

func (q SessionsQ) FilterAccountID(accountID uuid.UUID) SessionsQ {
	q.selector = q.selector.Where(sq.Eq{"account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"account_id": accountID})

	return q
}

func (q SessionsQ) OrderCreatedAt(ascending bool) SessionsQ {
	if ascending {
		q.selector = q.selector.OrderBy("created_at ASC")
	} else {
		q.selector = q.selector.OrderBy("created_at DESC")
	}
	return q
}

func (q SessionsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for sessions: %w", err)
	}

	var count uint
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q SessionsQ) Page(limit, offset uint) SessionsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
