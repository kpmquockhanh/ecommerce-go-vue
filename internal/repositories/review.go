package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api-go/internal/models"
)

type ReviewRepository interface {
	Create(ctx context.Context, productID, userID, rating int, comment string) (*models.Review, error)
	ListByProduct(ctx context.Context, productID, limit, offset int) ([]models.Review, error)
	GetStats(ctx context.Context, productID int) (float64, int, error)
	ExistsProduct(ctx context.Context, productID int) (bool, error)
}

type reviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, productID, userID, rating int, comment string) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRow(ctx,
		`INSERT INTO reviews (product_id, user_id, rating, comment)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (product_id, user_id) DO NOTHING
		 RETURNING id, product_id, user_id, rating, comment, created_at`,
		productID, userID, rating, comment,
	).Scan(&review.ID, &review.ProductID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) ListByProduct(ctx context.Context, productID, limit, offset int) ([]models.Review, error) {
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.product_id, r.user_id, u.first_name || ' ' || u.last_name as user_name,
			r.rating, r.comment, r.created_at
		 FROM reviews r
		 JOIN users u ON r.user_id = u.id
		 WHERE r.product_id = $1
		 ORDER BY r.created_at DESC
		 LIMIT $2 OFFSET $3`, productID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		if err := rows.Scan(&review.ID, &review.ProductID, &review.UserID, &review.UserName,
			&review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *reviewRepository) GetStats(ctx context.Context, productID int) (float64, int, error) {
	var avgRating float64
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE product_id = $1", productID,
	).Scan(&avgRating, &count)
	return avgRating, count, err
}

func (r *reviewRepository) ExistsProduct(ctx context.Context, productID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND status = 'published')", productID).Scan(&exists)
	return exists, err
}
