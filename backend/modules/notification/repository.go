package notification

import (
	"context"

	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles notification persistence.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) Create(ctx context.Context, userID, ntype, message, link string) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO notifications (user_id, type, message, link) VALUES ($1, $2, $3, $4)`,
		userID, ntype, message, link,
	)
	return err
}

func (r *Repository) List(ctx context.Context, userID string, limit int) ([]*Notification, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, user_id, type, message, COALESCE(link, ''), read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &n.Link, &n.Read, &n.CreatedAt); err != nil {
			continue
		}
		out = append(out, &n)
	}
	return out, nil
}

func (r *Repository) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pg.QueryRow(ctx,
		"SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = FALSE", userID).Scan(&count)
	return count, err
}

func (r *Repository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pg.Exec(ctx,
		"UPDATE notifications SET read = TRUE WHERE id = $1 AND user_id = $2", id, userID)
	return err
}

func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.pg.Exec(ctx,
		"UPDATE notifications SET read = TRUE WHERE user_id = $1 AND read = FALSE", userID)
	return err
}
