package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(databaseURL string) error {
	var err error
	DB, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err = DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL database")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func Migrate(ctx context.Context) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			role VARCHAR(20) DEFAULT 'customer',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			price INTEGER NOT NULL,
			images TEXT[] DEFAULT '{}',
			status VARCHAR(20) DEFAULT 'published',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS product_variants (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			size VARCHAR(50),
			color VARCHAR(50),
			stock INTEGER DEFAULT 0,
			sku VARCHAR(100) UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			session_id VARCHAR(255),
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
			quantity INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, product_id, variant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			status VARCHAR(30) DEFAULT 'pending',
			total INTEGER NOT NULL,
			shipping_first_name VARCHAR(100),
			shipping_last_name VARCHAR(100),
			shipping_address_1 VARCHAR(255),
			shipping_address_2 VARCHAR(255),
			shipping_city VARCHAR(100),
			shipping_state VARCHAR(50),
			shipping_zip VARCHAR(20),
			shipping_country VARCHAR(10),
			stripe_payment_intent VARCHAR(255),
			payment_status VARCHAR(30) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
			product_id INTEGER REFERENCES products(id),
			variant_id INTEGER REFERENCES product_variants(id),
			product_name VARCHAR(255),
			variant_label VARCHAR(100),
			quantity INTEGER NOT NULL,
			price INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL,
			comment TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(product_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS product_categories (
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
			PRIMARY KEY (product_id, category_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)`,
		`CREATE INDEX IF NOT EXISTS idx_products_price ON products(price)`,
		`CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_cart_items_user ON cart_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_product ON reviews(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_product_categories_category ON product_categories(category_id)`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
			payment_intent_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(key, user_id)
		)`,
		`DO $$ BEGIN ALTER TABLE idempotency_keys ADD COLUMN IF NOT EXISTS payment_intent_id VARCHAR(255); EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_keys_lookup ON idempotency_keys(key, user_id)`,
		`CREATE TABLE IF NOT EXISTS dead_letters (
			id SERIAL PRIMARY KEY,
			job_type VARCHAR(100) NOT NULL,
			queue_name VARCHAR(100) NOT NULL,
			payload JSONB NOT NULL,
			error_message TEXT,
			retry_count INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'failed',
			created_at TIMESTAMP DEFAULT NOW(),
			retried_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_dead_letters_status ON dead_letters(status)`,
		`CREATE INDEX IF NOT EXISTS idx_dead_letters_created ON dead_letters(created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS checkout_sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			idempotency_key VARCHAR(255) UNIQUE NOT NULL,
			payment_intent_id VARCHAR(255),
			status VARCHAR(20) DEFAULT 'initiated',
			cart_snapshot JSONB NOT NULL,
			total INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_checkout_sessions_user ON checkout_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_checkout_sessions_status ON checkout_sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_checkout_sessions_created ON checkout_sessions(created_at)`,
		// Migration: drop old columns/indexes if they exist (fresh schema migration)
		`DO $$ BEGIN ALTER TABLE products DROP COLUMN IF EXISTS category; EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products DROP COLUMN IF EXISTS active; EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DROP INDEX IF EXISTS idx_products_category`,
		// Migrate existing data: set status='published' for products that had active=true
		`DO $$ BEGIN UPDATE products SET status = 'published' WHERE status IS NULL; EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		// Story 3-5: Add new columns to products
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN compare_at_price INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN weight INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN length INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN width INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN height INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN meta_title VARCHAR(160); EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN meta_description VARCHAR(500); EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE products ADD COLUMN og_image TEXT; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		// Story 4: Add dimensions to product_variants
		`DO $$ BEGIN ALTER TABLE product_variants ADD COLUMN weight INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE product_variants ADD COLUMN length INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE product_variants ADD COLUMN width INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE product_variants ADD COLUMN height INTEGER; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		// Story 3: product_tags table
		`CREATE TABLE IF NOT EXISTS product_tags (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			tag VARCHAR(50) NOT NULL,
			UNIQUE(product_id, tag)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_tags_tag ON product_tags(tag)`,
		`CREATE INDEX IF NOT EXISTS idx_product_tags_product ON product_tags(product_id)`,
		// Story 5: product_image_order table
		`CREATE TABLE IF NOT EXISTS product_image_order (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			image_path TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			is_primary BOOLEAN DEFAULT false,
			UNIQUE(product_id, image_path)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_image_order_product ON product_image_order(product_id)`,
		// Product Options: 4 new tables
		`CREATE TABLE IF NOT EXISTS product_option_groups (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(product_id, name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_option_groups_product ON product_option_groups(product_id)`,
		`CREATE TABLE IF NOT EXISTS product_option_values (
			id SERIAL PRIMARY KEY,
			group_id INTEGER REFERENCES product_option_groups(id) ON DELETE CASCADE,
			value VARCHAR(100) NOT NULL,
			price_modifier INTEGER DEFAULT 0,
			UNIQUE(group_id, value)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_option_values_group ON product_option_values(group_id)`,
		`CREATE TABLE IF NOT EXISTS variant_option_values (
			id SERIAL PRIMARY KEY,
			variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
			option_value_id INTEGER REFERENCES product_option_values(id) ON DELETE CASCADE,
			UNIQUE(variant_id, option_value_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_variant_option_values_variant ON variant_option_values(variant_id)`,
		`CREATE TABLE IF NOT EXISTS product_customization_fields (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			field_type VARCHAR(50) NOT NULL DEFAULT 'text',
			required BOOLEAN DEFAULT false,
			default_value TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(product_id, name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_product_customization_fields_product ON product_customization_fields(product_id)`,
		// Add customization_data JSONB columns to cart_items and order_items
		`DO $$ BEGIN ALTER TABLE cart_items ADD COLUMN customization_data JSONB; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE order_items ADD COLUMN customization_data JSONB; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		// Data migration: create Size option groups from existing variants
		`DO $$ BEGIN
			INSERT INTO product_option_groups (product_id, name, sort_order)
			SELECT DISTINCT pv.product_id, 'Size', 0
			FROM product_variants pv
			WHERE pv.size IS NOT NULL AND pv.size != ''
			ON CONFLICT (product_id, name) DO NOTHING;
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN
			INSERT INTO product_option_values (group_id, value, price_modifier)
			SELECT DISTINCT pog.id, pv.size, 0
			FROM product_option_groups pog
			JOIN product_variants pv ON pv.product_id = pog.product_id
			WHERE pog.name = 'Size' AND pv.size IS NOT NULL AND pv.size != ''
			AND NOT EXISTS (SELECT 1 FROM product_option_values pov WHERE pov.group_id = pog.id AND pov.value = pv.size);
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN
			INSERT INTO variant_option_values (variant_id, option_value_id)
			SELECT pv.id, pov.id
			FROM product_variants pv
			JOIN product_option_groups pog ON pog.product_id = pv.product_id AND pog.name = 'Size'
			JOIN product_option_values pov ON pov.group_id = pog.id AND pov.value = pv.size
			WHERE pv.size IS NOT NULL AND pv.size != ''
			ON CONFLICT DO NOTHING;
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		// Data migration: create Color option groups from existing variants
		`DO $$ BEGIN
			INSERT INTO product_option_groups (product_id, name, sort_order)
			SELECT DISTINCT pv.product_id, 'Color', 1
			FROM product_variants pv
			WHERE pv.color IS NOT NULL AND pv.color != ''
			ON CONFLICT (product_id, name) DO NOTHING;
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN
			INSERT INTO product_option_values (group_id, value, price_modifier)
			SELECT DISTINCT pog.id, pv.color, 0
			FROM product_option_groups pog
			JOIN product_variants pv ON pv.product_id = pog.product_id
			WHERE pog.name = 'Color' AND pv.color IS NOT NULL AND pv.color != ''
			AND NOT EXISTS (SELECT 1 FROM product_option_values pov WHERE pov.group_id = pog.id AND pov.value = pv.color);
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN
			INSERT INTO variant_option_values (variant_id, option_value_id)
			SELECT pv.id, pov.id
			FROM product_variants pv
			JOIN product_option_groups pog ON pog.product_id = pv.product_id AND pog.name = 'Color'
			JOIN product_option_values pov ON pov.group_id = pog.id AND pov.value = pv.color
			WHERE pv.color IS NOT NULL AND pv.color != ''
			ON CONFLICT DO NOTHING;
		EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		// Drop size/color columns from product_variants
		`DO $$ BEGIN ALTER TABLE product_variants DROP COLUMN IF EXISTS size; EXCEPTION WHEN undefined_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE product_variants DROP COLUMN IF EXISTS color; EXCEPTION WHEN undefined_column THEN NULL; END $$`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w\nQuery: %s", err, m)
		}
	}

	log.Println("Database migrations completed")
	return nil
}

func Seed(ctx context.Context) error {
	// Seed default categories first
	categories := []string{"pants", "shirts", "shorts", "shoes", "accessories"}
	for _, catName := range categories {
		var exists bool
		err := DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)", catName).Scan(&exists)
		if err != nil {
			log.Printf("Warning: could not check category %s: %v", catName, err)
			continue
		}
		if exists {
			continue
		}
		if _, err := DB.Exec(ctx, "INSERT INTO categories (name) VALUES ($1)", catName); err != nil {
			log.Printf("Warning: could not seed category %s: %v", catName, err)
		} else {
			log.Printf("Seeded category: %s", catName)
		}
	}

	// Seed products and link to categories
	type seedProduct struct {
		Name        string
		Slug        string
		Description string
		Price       int
		CategoryIDs []int
	}
	products := []seedProduct{
		{"Forever Pants", "forever-pants", "Comfortable everyday pants made from sustainable materials.", 26000, []int{1}},
		{"Forever Shirt", "forever-shirt", "A classic shirt for all occasions.", 15500, []int{2}},
		{"Forever Shorts", "forever-shorts", "Lightweight shorts for warm weather.", 30000, []int{3}},
	}

	for _, p := range products {
		var exists bool
		err := DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)", p.Slug).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check product exists: %w", err)
		}
		if exists {
			continue
		}

		var productID int
		err = DB.QueryRow(ctx,
			`INSERT INTO products (name, slug, description, price, status)
			 VALUES ($1, $2, $3, $4, 'published')
			 RETURNING id`,
			p.Name, p.Slug, p.Description, p.Price,
		).Scan(&productID)
		if err != nil {
			return fmt.Errorf("seed product %s: %w", p.Name, err)
		}

		for _, catID := range p.CategoryIDs {
			if _, err := DB.Exec(ctx,
				`INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				productID, catID,
			); err != nil {
				log.Printf("Warning: could not link product %s to category %d: %v", p.Name, catID, err)
			}
		}
		log.Printf("Seeded product: %s", p.Name)
	}

	log.Println("Seed completed")

	// Seed admin user
	var adminExists bool
	if err := DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = 'admin@eshop.com')").Scan(&adminExists); err == nil && !adminExists {
		if _, err := DB.Exec(ctx,
			`INSERT INTO users (email, password_hash, first_name, last_name, role)
			 VALUES ('admin@eshop.com', '$2a$10$0b482qObKVu5o095Ymk1peTOH8HbogPoeHJdbfEOPdWdgWrKSbBIO', 'Admin', 'User', 'admin')`,
		); err != nil {
			log.Printf("Warning: could not seed admin user: %v", err)
		} else {
			log.Println("Seeded admin user: admin@eshop.com / password")
		}
	}

	return nil
}
