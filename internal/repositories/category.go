package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api-go/internal/models"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]models.Category, error)
	FindByID(ctx context.Context, id int) (*models.Category, error)
	Create(ctx context.Context, name string) (*models.Category, error)
	Update(ctx context.Context, id int, name string) error
	Delete(ctx context.Context, id int) error
	CountProducts(ctx context.Context, categoryID int) (int, error)
	RemoveCategoryFromProducts(ctx context.Context, categoryID int) error
}

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) List(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.name, c.created_at,
		 COALESCE(pc.product_count, 0) as product_count
		 FROM categories c
		 LEFT JOIN (
			SELECT pc.category_id, COUNT(*) as product_count
			FROM product_categories pc
			INNER JOIN products p ON pc.product_id = p.id AND p.status = 'published'
			GROUP BY pc.category_id
		 ) pc ON pc.category_id = c.id
		 ORDER BY c.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.CreatedAt, &cat.ProductCount); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id int) (*models.Category, error) {
	var cat models.Category
	err := r.db.QueryRow(ctx,
		`SELECT id, name, created_at FROM categories WHERE id = $1`, id,
	).Scan(&cat.ID, &cat.Name, &cat.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) Create(ctx context.Context, name string) (*models.Category, error) {
	var cat models.Category
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&cat.ID, &cat.Name, &cat.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) Update(ctx context.Context, id int, name string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE categories SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *categoryRepository) CountProducts(ctx context.Context, categoryID int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_categories pc
		 INNER JOIN products p ON pc.product_id = p.id
		 WHERE pc.category_id = $1`, categoryID,
	).Scan(&count)
	return count, err
}

func (r *categoryRepository) RemoveCategoryFromProducts(ctx context.Context, categoryID int) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM product_categories WHERE category_id = $1`, categoryID)
	return err
}
