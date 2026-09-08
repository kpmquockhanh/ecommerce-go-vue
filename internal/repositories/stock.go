package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockRepository interface {
	LockForUpdate(ctx context.Context, tx pgx.Tx, variantID int) (int, error)
	Decrement(ctx context.Context, tx pgx.Tx, variantID, quantity int) error
	Increment(ctx context.Context, variantID, quantity int) error
	GetAvailable(ctx context.Context, variantID, productID int) (int, error)
}

type stockRepository struct {
	db *pgxpool.Pool
}

func NewStockRepository(db *pgxpool.Pool) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) LockForUpdate(ctx context.Context, tx pgx.Tx, variantID int) (int, error) {
	var stock int
	err := tx.QueryRow(ctx,
		"SELECT stock FROM product_variants WHERE id = $1 FOR UPDATE",
		variantID).Scan(&stock)
	return stock, err
}

func (r *stockRepository) Decrement(ctx context.Context, tx pgx.Tx, variantID, quantity int) error {
	result, err := tx.Exec(ctx,
		"UPDATE product_variants SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
		quantity, variantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInsufficientStock
	}
	return nil
}

func (r *stockRepository) Increment(ctx context.Context, variantID, quantity int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE product_variants SET stock = stock + $1 WHERE id = $2",
		quantity, variantID)
	return err
}

func (r *stockRepository) GetAvailable(ctx context.Context, variantID, productID int) (int, error) {
	var stock int
	err := r.db.QueryRow(ctx,
		"SELECT stock FROM product_variants WHERE id = $1 AND product_id = $2",
		variantID, productID).Scan(&stock)
	return stock, err
}
