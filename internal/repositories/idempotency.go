package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepository interface {
	FindByKey(ctx context.Context, key string, userID int) (*IdempotencyResult, error)
	Store(ctx context.Context, key string, userID, orderID int) error
	Upsert(ctx context.Context, key string, userID, orderID int, paymentIntentID string) (*IdempotencyResult, error)
}

type IdempotencyResult struct {
	OrderID         int
	PaymentIntentID string
}

type idempotencyRepository struct {
	db *pgxpool.Pool
}

func NewIdempotencyRepository(db *pgxpool.Pool) IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) FindByKey(ctx context.Context, key string, userID int) (*IdempotencyResult, error) {
	var result IdempotencyResult
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(order_id, 0), COALESCE(payment_intent_id, '')
		 FROM idempotency_keys
		 WHERE key = $1 AND user_id = $2`,
		key, userID,
	).Scan(&result.OrderID, &result.PaymentIntentID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *idempotencyRepository) Store(ctx context.Context, key string, userID, orderID int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO idempotency_keys (key, user_id, order_id)
		 VALUES ($1, $2, $3)`,
		key, userID, orderID)
	return err
}

func (r *idempotencyRepository) Upsert(ctx context.Context, key string, userID, orderID int, paymentIntentID string) (*IdempotencyResult, error) {
	var result IdempotencyResult
	err := r.db.QueryRow(ctx,
		`INSERT INTO idempotency_keys (key, user_id, order_id, payment_intent_id)
		 VALUES ($1, $2, NULLIF($3, 0), $4)
		 ON CONFLICT (key, user_id) DO UPDATE SET
		   order_id = COALESCE(EXCLUDED.order_id, idempotency_keys.order_id),
		   payment_intent_id = COALESCE(EXCLUDED.payment_intent_id, idempotency_keys.payment_intent_id)
		 RETURNING COALESCE(idempotency_keys.order_id, 0), COALESCE(idempotency_keys.payment_intent_id, '')`,
		key, userID, orderID, paymentIntentID,
	).Scan(&result.OrderID, &result.PaymentIntentID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
