package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// ErrNotFound is returned when a subscription does not exist.
var ErrNotFound = errors.New("subscription not found")

// ErrDuplicate is returned when a subscription already exists.
var ErrDuplicate = errors.New("subscription already exists")

// Repository provides data access for subscriptions.
type Repository struct {
	db *pgxpool.Pool
}

// New creates a new Repository backed by the given connection pool.
func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a new subscription and returns it.
func (r *Repository) Create(ctx context.Context, webhook string, watchType models.WatchType, watchTarget string) (*models.Subscription, error) {
	const q = `
		INSERT INTO subscriptions (discord_webhook, watch_type, watch_target)
		VALUES ($1, $2, $3)
		RETURNING id, discord_webhook, watch_type, watch_target, active, created_at`

	var s models.Subscription
	err := r.db.QueryRow(ctx, q, webhook, watchType, watchTarget).
		Scan(&s.ID, &s.DiscordWebhook, &s.WatchType, &s.WatchTarget, &s.Active, &s.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &s, nil
}

// GetByID retrieves a single subscription by its UUID.
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	const q = `
		SELECT id, discord_webhook, watch_type, watch_target, active, created_at
		FROM subscriptions WHERE id = $1`

	var s models.Subscription
	err := r.db.QueryRow(ctx, q, id).
		Scan(&s.ID, &s.DiscordWebhook, &s.WatchType, &s.WatchTarget, &s.Active, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete soft-deletes a subscription by setting active = false.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const q = `UPDATE subscriptions SET active = FALSE WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListActive returns all subscriptions where active = true.
func (r *Repository) ListActive(ctx context.Context) ([]models.Subscription, error) {
	const q = `
		SELECT id, discord_webhook, watch_type, watch_target, active, created_at
		FROM subscriptions WHERE active = TRUE`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var s models.Subscription
		if err := rows.Scan(&s.ID, &s.DiscordWebhook, &s.WatchType, &s.WatchTarget, &s.Active, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// isUniqueViolation checks whether the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	// pgx wraps pgconn.PgError; check SQLSTATE 23505.
	type pgErr interface{ SQLState() string }
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
