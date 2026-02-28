package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/pgdbx"
)

const organizationsTable = "organizations"
const organizationsColumns = "id, source_created_at, replica_created_at"

func scanOrganization(row sq.RowScanner) (r repository.OrganizationRow, err error) {
	err = row.Scan(
		&r.ID,
		&r.SourceCreatedAt,
		&r.ReplicaCreatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.OrganizationRow{}, nil
	case err != nil:
		return repository.OrganizationRow{}, fmt.Errorf("scanning organization: %w", err)
	}

	return r, nil
}

type organizations struct {
	db       *pgdbx.DB
	inserter sq.InsertBuilder
	selector sq.SelectBuilder
	deleter  sq.DeleteBuilder
}

func NewOrganizationsQ(db *pgdbx.DB) repository.OrganizationsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return organizations{
		db:       db,
		inserter: builder.Insert(organizationsTable),
		selector: builder.Select(organizationsColumns).From(organizationsTable),
		deleter:  builder.Delete(organizationsTable),
	}
}

func (q organizations) New() repository.OrganizationsQ {
	return NewOrganizationsQ(q.db)
}

func (q organizations) Insert(ctx context.Context, data repository.OrganizationRow) error {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":                pgtype.UUID{Bytes: data.ID, Valid: true},
		"source_created_at": pgtype.Timestamptz{Time: data.SourceCreatedAt.UTC(), Valid: true},
	}).ToSql()
	if err != nil {
		return fmt.Errorf("building insert query for %s: %w", organizationsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("inserting organization: %w", err)
	}

	return nil
}

func (q organizations) Get(ctx context.Context) (repository.OrganizationRow, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return repository.OrganizationRow{}, fmt.Errorf("building get query for %s: %w", organizationsTable, err)
	}

	return scanOrganization(q.db.QueryRow(ctx, query, args...))
}

func (q organizations) Select(ctx context.Context) ([]repository.OrganizationRow, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", organizationsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", organizationsTable, err)
	}
	defer rows.Close()

	var out []repository.OrganizationRow
	for rows.Next() {
		r, err := scanOrganization(rows)
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

func (q organizations) Exists(ctx context.Context) (bool, error) {
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

func (q organizations) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", organizationsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", organizationsTable, err)
	}

	return nil
}

func (q organizations) FilterByID(id uuid.UUID) repository.OrganizationsQ {
	pid := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"id": pid})

	return q
}
