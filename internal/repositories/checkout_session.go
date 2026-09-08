package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckoutSessionRepository interface {
	Create(ctx context.Context, params CreateCheckoutSessionParams) (*CheckoutSession, error)
	FindByKey(ctx context.Context, idempotencyKey string, userID int) (*CheckoutSession, error)
	UpdateStatus(ctx context.Context, idempotencyKey string, userID int, status string) error
	UpdateOrderID(ctx context.Context, idempotencyKey string, userID int, orderID int) error
	FindAbandoned(ctx context.Context, olderThan time.Duration, limit int) ([]CheckoutSession, error)
	ExpireOldSessions(ctx context.Context, olderThan time.Duration) (int, error)
}

type CheckoutSession struct {
	ID              int
	UserID          int
	IdempotencyKey  string
	PaymentIntentID string
	Status          string
	CartSnapshot    json.RawMessage
	Total           int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CartSnapshotItem struct {
	ProductID        int               `json:"product_id"`
	VariantID        *int              `json:"variant_id"`
	ProductName      string            `json:"product_name"`
	VariantLabel     string            `json:"variant_label"`
	Quantity         int               `json:"quantity"`
	Price            int               `json:"price"`
	CustomizationData map[string]string `json:"customization_data,omitempty"`
}

type CreateCheckoutSessionParams struct {
	UserID          int
	IdempotencyKey  string
	PaymentIntentID string
	CartItems       []CartSnapshotItem
	Total           int
}

type checkoutSessionRepository struct {
	db *pgxpool.Pool
}

func NewCheckoutSessionRepository(db *pgxpool.Pool) CheckoutSessionRepository {
	return &checkoutSessionRepository{db: db}
}

func (r *checkoutSessionRepository) Create(ctx context.Context, params CreateCheckoutSessionParams) (*CheckoutSession, error) {
	cartJSON, err := json.Marshal(params.CartItems)
	if err != nil {
		return nil, err
	}

	var session CheckoutSession
	err = r.db.QueryRow(ctx,
		`INSERT INTO checkout_sessions (user_id, idempotency_key, payment_intent_id, cart_snapshot, total)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (idempotency_key) DO UPDATE SET
		   payment_intent_id = EXCLUDED.payment_intent_id,
		   cart_snapshot = EXCLUDED.cart_snapshot,
		   total = EXCLUDED.total,
		   status = 'initiated',
		   updated_at = NOW()
		 RETURNING id, user_id, idempotency_key, payment_intent_id, status, cart_snapshot, total, created_at, updated_at`,
		params.UserID, params.IdempotencyKey, params.PaymentIntentID, cartJSON, params.Total,
	).Scan(&session.ID, &session.UserID, &session.IdempotencyKey, &session.PaymentIntentID,
		&session.Status, &session.CartSnapshot, &session.Total, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *checkoutSessionRepository) FindByKey(ctx context.Context, idempotencyKey string, userID int) (*CheckoutSession, error) {
	var session CheckoutSession
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, idempotency_key, COALESCE(payment_intent_id, ''), status, cart_snapshot, total, created_at, updated_at
		 FROM checkout_sessions
		 WHERE idempotency_key = $1 AND user_id = $2`,
		idempotencyKey, userID,
	).Scan(&session.ID, &session.UserID, &session.IdempotencyKey, &session.PaymentIntentID,
		&session.Status, &session.CartSnapshot, &session.Total, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *checkoutSessionRepository) UpdateStatus(ctx context.Context, idempotencyKey string, userID int, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE checkout_sessions SET status = $1, updated_at = NOW()
		 WHERE idempotency_key = $2 AND user_id = $3`,
		status, idempotencyKey, userID)
	return err
}

func (r *checkoutSessionRepository) UpdateOrderID(ctx context.Context, idempotencyKey string, userID int, orderID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE checkout_sessions SET status = 'confirmed', updated_at = NOW()
		 WHERE idempotency_key = $1 AND user_id = $2`,
		idempotencyKey, userID)
	return err
}

func (r *checkoutSessionRepository) FindAbandoned(ctx context.Context, olderThan time.Duration, limit int) ([]CheckoutSession, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, idempotency_key, COALESCE(payment_intent_id, ''), status, cart_snapshot, total, created_at, updated_at
		 FROM checkout_sessions
		 WHERE status = 'initiated' AND created_at < NOW() - make_interval(secs => $1)
		 ORDER BY created_at ASC
		 LIMIT $2`,
		int(olderThan.Seconds()), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []CheckoutSession
	for rows.Next() {
		var s CheckoutSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.IdempotencyKey, &s.PaymentIntentID,
			&s.Status, &s.CartSnapshot, &s.Total, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *checkoutSessionRepository) ExpireOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	result, err := r.db.Exec(ctx,
		`UPDATE checkout_sessions SET status = 'expired', updated_at = NOW()
		 WHERE status = 'initiated' AND created_at < NOW() - make_interval(secs => $1)`,
		int(olderThan.Seconds()))
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}
