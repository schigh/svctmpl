package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Common sentinel errors.
var (
	ErrNotFound = errors.New("resource not found")
)

// Resource is the domain entity persisted in the resources table.
type Resource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResourceRepository defines the persistence operations for resources.
type ResourceRepository interface {
	Create(ctx context.Context, name, description string) (Resource, error)
	GetByID(ctx context.Context, id string) (Resource, error)
	List(ctx context.Context) ([]Resource, error)
	Update(ctx context.Context, id, name, description string) (Resource, error)
	Delete(ctx context.Context, id string) error
}

// --- SQL queries (inline constants, sqlc-style) ---

const queryCreate = `
INSERT INTO resources (name, description)
VALUES ($1, $2)
RETURNING id, name, description, created_at, updated_at`

const queryGetByID = `
SELECT id, name, description, created_at, updated_at
FROM resources
WHERE id = $1`

const queryList = `
SELECT id, name, description, created_at, updated_at
FROM resources
ORDER BY created_at DESC`

const queryUpdate = `
UPDATE resources
SET name = $1, description = $2, updated_at = NOW()
WHERE id = $3
RETURNING id, name, description, created_at, updated_at`

const queryDelete = `
DELETE FROM resources WHERE id = $1`

// PgxResourceRepository implements ResourceRepository using pgx/v5.
type PgxResourceRepository struct {
	pool *pgxpool.Pool
}

// NewPgxResourceRepository creates a repository backed by the given connection pool.
func NewPgxResourceRepository(pool *pgxpool.Pool) *PgxResourceRepository {
	return &PgxResourceRepository{pool: pool}
}

// Create inserts a new resource and returns the created row.
func (r *PgxResourceRepository) Create(ctx context.Context, name, description string) (Resource, error) {
	var res Resource
	err := r.pool.QueryRow(ctx, queryCreate, name, description).
		Scan(&res.ID, &res.Name, &res.Description, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return Resource{}, fmt.Errorf("creating resource: %w", err)
	}
	return res, nil
}

// GetByID retrieves a single resource by primary key.
func (r *PgxResourceRepository) GetByID(ctx context.Context, id string) (Resource, error) {
	var res Resource
	err := r.pool.QueryRow(ctx, queryGetByID, id).
		Scan(&res.ID, &res.Name, &res.Description, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Resource{}, ErrNotFound
		}
		return Resource{}, fmt.Errorf("getting resource %s: %w", id, err)
	}
	return res, nil
}

// List returns all resources ordered by creation time descending.
func (r *PgxResourceRepository) List(ctx context.Context) ([]Resource, error) {
	rows, err := r.pool.Query(ctx, queryList)
	if err != nil {
		return nil, fmt.Errorf("listing resources: %w", err)
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(&res.ID, &res.Name, &res.Description, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning resource row: %w", err)
		}
		resources = append(resources, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating resource rows: %w", err)
	}
	return resources, nil
}

// Update modifies an existing resource's name and description.
func (r *PgxResourceRepository) Update(ctx context.Context, id, name, description string) (Resource, error) {
	var res Resource
	err := r.pool.QueryRow(ctx, queryUpdate, name, description, id).
		Scan(&res.ID, &res.Name, &res.Description, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Resource{}, ErrNotFound
		}
		return Resource{}, fmt.Errorf("updating resource %s: %w", id, err)
	}
	return res, nil
}

// Delete removes a resource by primary key. Returns ErrNotFound if no row
// was affected.
func (r *PgxResourceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryDelete, id)
	if err != nil {
		return fmt.Errorf("deleting resource %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
