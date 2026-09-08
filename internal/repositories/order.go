package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api-go/internal/models"
)

type OrderRepository interface {
	Checkout(ctx context.Context, userID int, params CheckoutParams) (int, error)
	FindByID(ctx context.Context, orderID, userID int) (*models.Order, error)
	FindByPaymentIntent(ctx context.Context, piID string) (*models.Order, error)
	ListByUserID(ctx context.Context, userID, limit, offset int) ([]OrderSummary, int, error)
	ListAll(ctx context.Context, status string, limit, offset int) ([]AdminOrder, int, error)
	UpdateStatus(ctx context.Context, orderID int, status string) error
	MarkPaid(ctx context.Context, orderID int) (int64, error)
	MarkPaymentFailed(ctx context.Context, orderID int) error
	GetItems(ctx context.Context, orderID int) ([]models.OrderItem, error)
	GetItemsForStockRestore(ctx context.Context, orderID int) ([]StockRestoreItem, error)
	GetStatus(ctx context.Context, orderID int) (string, error)
	GetUserIDAndTotal(ctx context.Context, orderID int) (int, int, error)
	GetStats(ctx context.Context) (*DashboardStats, error)
}

type DashboardStats struct {
	TotalOrders   int   `json:"total_orders"`
	TotalRevenue  int   `json:"total_revenue"`
	PendingOrders int   `json:"pending_orders"`
	DeliveredOrders int `json:"delivered_orders"`
}

type StockRestoreItem struct {
	VariantID int
	Quantity  int
}

type CheckoutParams struct {
	Total            int
	ShippingAddress  models.Address
	PaymentIntentID  string
	CartItems        []CheckoutCartItem
	IdempotencyKey   string
}

type CheckoutCartItem struct {
	ProductID        int
	VariantID        *int
	Quantity         int
	Price            int
	ProductName      string
	VariantLabel     string
	CustomizationData []byte
}

type OrderSummary struct {
	ID         int       `json:"id"`
	Status     string    `json:"status"`
	Total      int       `json:"total"`
	CreatedAt  time.Time `json:"created_at"`
	ItemsCount int       `json:"items_count"`
}

type AdminOrder struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	CustomerName string    `json:"customer_name"`
	Status       string    `json:"status"`
	Total        int       `json:"total"`
	CreatedAt    time.Time `json:"created_at"`
}

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Checkout(ctx context.Context, userID int, params CheckoutParams) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	cartLockRows, err := tx.Query(ctx,
		`SELECT ci.id, ci.quantity, p.id, p.name, p.slug, p.price, p.images,
			pv.id, pv.stock, ci.customization_data
		 FROM cart_items ci
		 JOIN products p ON ci.product_id = p.id
		 LEFT JOIN product_variants pv ON ci.variant_id = pv.id
		 WHERE ci.user_id = $1 FOR UPDATE`, userID)
	if err != nil {
		return 0, err
	}
	cartLockRows.Close()

	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total, 
			shipping_first_name, shipping_last_name, shipping_address_1, shipping_address_2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			stripe_payment_intent, payment_status)
		 VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'pending')
		 RETURNING id`,
		userID, params.Total,
		params.ShippingAddress.FirstName, params.ShippingAddress.LastName,
		params.ShippingAddress.Address1, params.ShippingAddress.Address2,
		params.ShippingAddress.City, params.ShippingAddress.State,
		params.ShippingAddress.Zip, params.ShippingAddress.Country,
		params.PaymentIntentID,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	for _, ci := range params.CartItems {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, variant_id, product_name, variant_label, quantity, price, customization_data)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			orderID, ci.ProductID, ci.VariantID, ci.ProductName, ci.VariantLabel, ci.Quantity, ci.Price, ci.CustomizationData)
		if err != nil {
			return 0, err
		}
	}

	for _, ci := range params.CartItems {
		if ci.VariantID == nil {
			continue
		}
		result, err := tx.Exec(ctx,
			"UPDATE product_variants SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
			ci.Quantity, *ci.VariantID)
		if err != nil {
			return 0, err
		}
		if result.RowsAffected() == 0 {
			return 0, ErrInsufficientStock
		}
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO idempotency_keys (key, user_id, order_id)
		 VALUES ($1, $2, $3)`,
		params.IdempotencyKey, userID, orderID)
	if err != nil {
		return 0, err
	}

	return orderID, tx.Commit(ctx)
}

func (r *orderRepository) FindByID(ctx context.Context, orderID, userID int) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, status, total,
			shipping_first_name, shipping_last_name, shipping_address_1, shipping_address_2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			stripe_payment_intent, payment_status, created_at, updated_at
		 FROM orders WHERE id = $1 AND user_id = $2`, orderID, userID,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.Total,
		&order.ShippingAddress.FirstName, &order.ShippingAddress.LastName,
		&order.ShippingAddress.Address1, &order.ShippingAddress.Address2,
		&order.ShippingAddress.City, &order.ShippingAddress.State,
		&order.ShippingAddress.Zip, &order.ShippingAddress.Country,
		&order.StripePaymentIntent, &order.PaymentStatus,
		&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByPaymentIntent(ctx context.Context, piID string) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, status, total, payment_status
		 FROM orders WHERE stripe_payment_intent = $1`, piID,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.Total, &order.PaymentStatus)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID, limit, offset int) ([]OrderSummary, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT o.id, o.status, o.total, o.created_at,
			(SELECT COUNT(*) FROM order_items WHERE order_id = o.id) as items_count
		 FROM orders o
		 WHERE o.user_id = $1
		 ORDER BY o.created_at DESC
		 LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []OrderSummary
	for rows.Next() {
		var o OrderSummary
		if err := rows.Scan(&o.ID, &o.Status, &o.Total, &o.CreatedAt, &o.ItemsCount); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *orderRepository) ListAll(ctx context.Context, status string, limit, offset int) ([]AdminOrder, int, error) {
	where := "1=1"
	var args []interface{}
	argIdx := 1

	if status != "" {
		validStatuses := map[string]bool{
			"pending": true, "paid": true, "shipped": true, "delivered": true, "cancelled": true,
		}
		if validStatuses[status] {
			where += " AND o.status = $" + strconv.Itoa(argIdx)
			args = append(args, status)
			argIdx++
		}
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM orders o WHERE " + where
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectArgs := make([]interface{}, len(args), len(args)+2)
	copy(selectArgs, args)
	selectArgs = append(selectArgs, limit, offset)

	selectQuery := `SELECT o.id, o.user_id, o.status, o.total, o.created_at,
		u.first_name || ' ' || u.last_name as customer_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE ` + where + `
		ORDER BY o.created_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []AdminOrder
	for rows.Next() {
		var o AdminOrder
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.CreatedAt, &o.CustomerName); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID int, status string) error {
	result, err := r.db.Exec(ctx,
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2",
		status, orderID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *orderRepository) MarkPaid(ctx context.Context, orderID int) (int64, error) {
	result, err := r.db.Exec(ctx,
		`UPDATE orders 
		 SET status = 'paid', payment_status = 'paid', updated_at = NOW() 
		 WHERE id = $1 AND payment_status != 'paid'`,
		orderID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *orderRepository) MarkPaymentFailed(ctx context.Context, orderID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders 
		 SET payment_status = 'failed', updated_at = NOW() 
		 WHERE id = $1`,
		orderID)
	return err
}

func (r *orderRepository) GetItems(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, variant_id, product_name, variant_label, quantity, price, (quantity * price) as subtotal
		 FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.VariantID, &item.ProductName, &item.VariantLabel, &item.Quantity, &item.Price, &item.Subtotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *orderRepository) GetItemsForStockRestore(ctx context.Context, orderID int) ([]StockRestoreItem, error) {
	rows, err := r.db.Query(ctx,
		"SELECT variant_id, quantity FROM order_items WHERE order_id = $1 AND variant_id IS NOT NULL", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StockRestoreItem
	for rows.Next() {
		var item StockRestoreItem
		if err := rows.Scan(&item.VariantID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *orderRepository) GetStatus(ctx context.Context, orderID int) (string, error) {
	var status string
	err := r.db.QueryRow(ctx,
		"SELECT status FROM orders WHERE id = $1", orderID).Scan(&status)
	return status, err
}

func (r *orderRepository) GetUserIDAndTotal(ctx context.Context, orderID int) (int, int, error) {
	var userID, total int
	err := r.db.QueryRow(ctx,
		"SELECT user_id, total FROM orders WHERE id = $1", orderID).Scan(&userID, &total)
	return userID, total, err
}

func (r *orderRepository) GetStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total_orders,
			COALESCE(SUM(total), 0) AS total_revenue,
			COUNT(*) FILTER (WHERE status = 'pending') AS pending_orders,
			COUNT(*) FILTER (WHERE status = 'delivered') AS delivered_orders
		FROM orders
	`).Scan(&stats.TotalOrders, &stats.TotalRevenue, &stats.PendingOrders, &stats.DeliveredOrders)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
