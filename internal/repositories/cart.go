package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]CartItemRow, error)
	FindByUserIDForUpdate(ctx context.Context, tx pgx.Tx, userID int) ([]CartItemRow, error)
	FindBySessionID(ctx context.Context, sessionID string) ([]CartItemRow, error)
	Upsert(ctx context.Context, userID *int, sessionID *string, productID int, variantID *int, quantity int, customizationData []byte) error
	UpdateQuantity(ctx context.Context, itemID int, userID *int, sessionID *string, quantity int) error
	Delete(ctx context.Context, itemID int, userID *int, sessionID *string) error
	GetGuestItems(ctx context.Context, sessionID string) ([]GuestCartItem, error)
	MergeGuestCart(ctx context.Context, userID int, sessionID string) (int, error)
	ClearByUserID(ctx context.Context, userID int) error
	InsertItems(ctx context.Context, userID int, items []CartItemRow) error
}

type CartItemRow struct {
	CartItemID       int
	Quantity         int
	ProductID        int
	ProductName      string
	ProductSlug      string
	ProductPrice     int
	ProductImages    []string
	VariantID        *int
	VariantLabel     *string
	VariantStock     *int
	CustomizationData []byte
}

type GuestCartItem struct {
	ProductID int
	VariantID *int
	Quantity  int
}

type cartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) CartRepository {
	return &cartRepository{db: db}
}

const cartSelectQuery = `SELECT ci.id, ci.quantity, 
	p.id, p.name, p.slug, p.price, p.images,
	pv.id, pv.stock, ci.customization_data,
	(SELECT string_agg(pov.value, ' / ' ORDER BY pog.sort_order)
	 FROM variant_option_values vov
	 JOIN product_option_values pov ON vov.option_value_id = pov.id
	 JOIN product_option_groups pog ON pov.group_id = pog.id
	 WHERE vov.variant_id = pv.id) as variant_label
 FROM cart_items ci
 JOIN products p ON ci.product_id = p.id
 LEFT JOIN product_variants pv ON ci.variant_id = pv.id`

func (r *cartRepository) FindByUserID(ctx context.Context, userID int) ([]CartItemRow, error) {
	rows, err := r.db.Query(ctx, cartSelectQuery+` WHERE ci.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCartRows(rows)
}

func (r *cartRepository) FindByUserIDForUpdate(ctx context.Context, tx pgx.Tx, userID int) ([]CartItemRow, error) {
	rows, err := tx.Query(ctx, cartSelectQuery+` WHERE ci.user_id = $1 FOR UPDATE`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCartRows(rows)
}

func (r *cartRepository) FindBySessionID(ctx context.Context, sessionID string) ([]CartItemRow, error) {
	rows, err := r.db.Query(ctx, cartSelectQuery+` WHERE ci.session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCartRows(rows)
}

func scanCartRows(rows pgx.Rows) ([]CartItemRow, error) {
	var items []CartItemRow
	for rows.Next() {
		var row CartItemRow
		if err := rows.Scan(&row.CartItemID, &row.Quantity,
			&row.ProductID, &row.ProductName, &row.ProductSlug, &row.ProductPrice, &row.ProductImages,
			&row.VariantID, &row.VariantStock, &row.CustomizationData, &row.VariantLabel); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, nil
}

func (r *cartRepository) Upsert(ctx context.Context, userID *int, sessionID *string, productID int, variantID *int, quantity int, customizationData []byte) error {
	if userID != nil {
		_, err := r.db.Exec(ctx,
			`INSERT INTO cart_items (user_id, product_id, variant_id, quantity, customization_data)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (user_id, product_id, variant_id)
			 DO UPDATE SET quantity = cart_items.quantity + $4, customization_data = $5`,
			*userID, productID, variantID, quantity, customizationData)
		return err
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO cart_items (session_id, product_id, variant_id, quantity, customization_data)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (session_id, product_id, variant_id)
		 DO UPDATE SET quantity = cart_items.quantity + $4, customization_data = $5`,
		*sessionID, productID, variantID, quantity, customizationData)
	return err
}

func (r *cartRepository) UpdateQuantity(ctx context.Context, itemID int, userID *int, sessionID *string, quantity int) error {
	if userID != nil {
		_, err := r.db.Exec(ctx,
			"UPDATE cart_items SET quantity = $1 WHERE id = $2 AND user_id = $3",
			quantity, itemID, *userID)
		return err
	}
	_, err := r.db.Exec(ctx,
		"UPDATE cart_items SET quantity = $1 WHERE id = $2 AND session_id = $3",
		quantity, itemID, *sessionID)
	return err
}

func (r *cartRepository) Delete(ctx context.Context, itemID int, userID *int, sessionID *string) error {
	if userID != nil {
		_, err := r.db.Exec(ctx,
			"DELETE FROM cart_items WHERE id = $1 AND user_id = $2", itemID, *userID)
		return err
	}
	_, err := r.db.Exec(ctx,
		"DELETE FROM cart_items WHERE id = $1 AND session_id = $2", itemID, *sessionID)
	return err
}

func (r *cartRepository) GetGuestItems(ctx context.Context, sessionID string) ([]GuestCartItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT product_id, variant_id, quantity 
		 FROM cart_items WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GuestCartItem
	for rows.Next() {
		var item GuestCartItem
		if err := rows.Scan(&item.ProductID, &item.VariantID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *cartRepository) MergeGuestCart(ctx context.Context, userID int, sessionID string) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	items, err := r.GetGuestItems(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	merged := 0
	for _, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO cart_items (user_id, product_id, variant_id, quantity)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, product_id, variant_id)
			 DO UPDATE SET quantity = cart_items.quantity + $4`,
			userID, item.ProductID, item.VariantID, item.Quantity)
		if err != nil {
			continue
		}
		merged++
	}

	_, _ = tx.Exec(ctx,
		"DELETE FROM cart_items WHERE session_id = $1", sessionID)

	return merged, tx.Commit(ctx)
}

func (r *cartRepository) ClearByUserID(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM cart_items WHERE user_id = $1", userID)
	return err
}

func (r *cartRepository) InsertItems(ctx context.Context, userID int, items []CartItemRow) error {
	for _, item := range items {
		_, err := r.db.Exec(ctx,
			`INSERT INTO cart_items (user_id, product_id, variant_id, quantity)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, product_id, variant_id)
			 DO UPDATE SET quantity = cart_items.quantity + $4`,
			userID, item.ProductID, item.VariantID, item.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}
