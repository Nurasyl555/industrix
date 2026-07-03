package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all catalog-related database operations
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new catalog repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// === Categories ===

func (r *Repository) ListCategories(ctx context.Context) ([]*Category, error) {
	rows, err := r.pg.Query(ctx, "SELECT id, name, slug, COALESCE(parent_id::text, '') FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var c Category
		var parentID string
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &parentID); err != nil {
			continue
		}
		if parentID != "" {
			c.ParentID = &parentID
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

func (r *Repository) CategoryExists(ctx context.Context, categoryID string) (bool, error) {
	var count int
	err := r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM categories WHERE id = $1", categoryID).Scan(&count)
	return count > 0, err
}

// === Equipment ===

func (r *Repository) CreateEquipment(ctx context.Context, eq *Equipment) error {
	err := r.pg.QueryRow(ctx,
		`INSERT INTO equipment (owner_id, category_id, title, description, condition, region)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`,
		eq.OwnerID, eq.CategoryID, eq.Title, eq.Description, eq.Condition, eq.Region,
	).Scan(&eq.ID, &eq.CreatedAt, &eq.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create equipment: %w", err)
	}
	return nil
}

func (r *Repository) GetEquipmentByID(ctx context.Context, id string) (*Equipment, error) {
	var eq Equipment
	err := r.pg.QueryRow(ctx,
		`SELECT id, owner_id, category_id, title, COALESCE(description, ''), condition, COALESCE(region, ''), created_at, updated_at
		 FROM equipment WHERE id = $1`, id,
	).Scan(&eq.ID, &eq.OwnerID, &eq.CategoryID, &eq.Title, &eq.Description, &eq.Condition, &eq.Region, &eq.CreatedAt, &eq.UpdatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Equipment not found")
	}
	return &eq, nil
}

func (r *Repository) ListEquipment(ctx context.Context, f ListEquipmentFilter) ([]*Equipment, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("category_id = $%d", argN))
		args = append(args, f.CategoryID)
		argN++
	}
	if f.Region != "" {
		where = append(where, fmt.Sprintf("region = $%d", argN))
		args = append(args, f.Region)
		argN++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", argN))
		args = append(args, "%"+f.Search+"%")
		argN++
	}
	whereClause := strings.Join(where, " AND ")

	offset := (f.Page - 1) * f.Limit
	query := fmt.Sprintf(
		`SELECT id, owner_id, category_id, title, COALESCE(description, ''), condition, COALESCE(region, ''), created_at, updated_at
		 FROM equipment WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argN, argN+1,
	)
	args = append(args, f.Limit, offset)

	rows, err := r.pg.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Equipment
	for rows.Next() {
		var eq Equipment
		if err := rows.Scan(&eq.ID, &eq.OwnerID, &eq.CategoryID, &eq.Title, &eq.Description, &eq.Condition, &eq.Region, &eq.CreatedAt, &eq.UpdatedAt); err != nil {
			continue
		}
		items = append(items, &eq)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM equipment WHERE %s", whereClause)
	var total int64
	_ = r.pg.QueryRow(ctx, countQuery, args[:argN-1]...).Scan(&total)

	return items, total, nil
}

func (r *Repository) UpdateEquipment(ctx context.Context, eq *Equipment) error {
	_, err := r.pg.Exec(ctx,
		`UPDATE equipment SET title = $1, description = $2, condition = $3, region = $4, updated_at = NOW()
		 WHERE id = $5`,
		eq.Title, eq.Description, eq.Condition, eq.Region, eq.ID,
	)
	return err
}

func (r *Repository) DeleteEquipment(ctx context.Context, id string) error {
	_, err := r.pg.Exec(ctx, "DELETE FROM equipment WHERE id = $1", id)
	return err
}
