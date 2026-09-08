package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api-go/internal/models"
)

type ProductRepository interface {
	FindBySlug(ctx context.Context, slug string) (*models.Product, error)
	FindByID(ctx context.Context, id int) (*models.Product, error)
	Create(ctx context.Context, p *models.CreateProductRequest, slug string) (*models.Product, error)
	Update(ctx context.Context, id int, p *models.UpdateProductRequest) (*models.Product, error)
	SoftDelete(ctx context.Context, id int) error
	List(ctx context.Context, filters ProductFilters) ([]models.Product, int, error)
	AdminList(ctx context.Context, filters ProductFilters) ([]models.Product, int, error)
	GetImages(ctx context.Context, productID int) ([]string, error)
	RemoveImageFromProduct(ctx context.Context, productID int, imagePath string) error
	AddImageToProduct(ctx context.Context, productID int, imagePath string) error
	GetVariants(ctx context.Context, productID int) ([]models.ProductVariant, error)
	CountVariants(ctx context.Context, productID int) (int, error)
	GetVariantProductID(ctx context.Context, variantID int) (int, error)
	GetReviewStats(ctx context.Context, productID int) (float64, int, error)
	LinkCategories(ctx context.Context, productID int, categoryIDs []int) error
	GetCategories(ctx context.Context, productID int) ([]models.Category, error)
	CreateVariant(ctx context.Context, productID int, v *models.CreateVariantRequest) (*models.ProductVariant, error)
	UpdateVariant(ctx context.Context, variantID int, v *models.UpdateVariantRequest) (*models.ProductVariant, error)
	DeleteVariant(ctx context.Context, variantID int) error
	ValidateOptionValues(ctx context.Context, productID int, optionValueIDs []int) error
	HasVariantCombination(ctx context.Context, productID int, optionValueIDs []int, excludeVariantID int) (bool, error)
	GetTags(ctx context.Context, productID int) ([]string, error)
	AddTags(ctx context.Context, productID int, tags []string) error
	RemoveTag(ctx context.Context, productID int, tag string) error
	GetImageOrder(ctx context.Context, productID int) ([]models.ProductImageOrder, error)
	UpdateImageOrder(ctx context.Context, productID int, images []models.ImageOrderEntry) error
	GetOptionGroupsByProduct(ctx context.Context, productID int) ([]models.ProductOptionGroup, error)
	CreateOptionGroup(ctx context.Context, productID int, req *models.CreateOptionGroupRequest) (*models.ProductOptionGroup, error)
	UpdateOptionGroup(ctx context.Context, groupID int, req *models.UpdateOptionGroupRequest) (*models.ProductOptionGroup, error)
	DeleteOptionGroup(ctx context.Context, groupID int) error
	CreateOptionValue(ctx context.Context, groupID int, req *models.CreateOptionValueRequest) (*models.ProductOptionValue, error)
	UpdateOptionValue(ctx context.Context, valueID int, req *models.UpdateOptionValueRequest) (*models.ProductOptionValue, error)
	DeleteOptionValue(ctx context.Context, valueID int) error
	ReorderOptionValues(ctx context.Context, groupID int, valueIDs []int) error
	GetCustomizationFieldsByProduct(ctx context.Context, productID int) ([]models.ProductCustomizationField, error)
	CreateCustomizationField(ctx context.Context, productID int, req *models.CreateCustomizationFieldRequest) (*models.ProductCustomizationField, error)
	UpdateCustomizationField(ctx context.Context, fieldID int, req *models.UpdateCustomizationFieldRequest) (*models.ProductCustomizationField, error)
	DeleteCustomizationField(ctx context.Context, fieldID int) error
	ReorderCustomizationFields(ctx context.Context, productID int, fieldIDs []int) error
}

type ProductFilters struct {
	CategoryIDs []int
	Tags        []string
	Status      string
	MinPrice    *int
	MaxPrice    *int
	Search      string
	Sort        string
	Limit       int
	Offset      int
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow(ctx,
		`SELECT p.id, p.name, p.slug, p.description, p.price, p.compare_at_price, p.images, p.status,
		 p.weight, p.length, p.width, p.height, p.meta_title, p.meta_description, p.og_image,
		 p.created_at, p.updated_at
		 FROM products p WHERE p.slug = $1 AND p.status = 'published'`, slug,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.CompareAtPrice, &p.Images, &p.Status,
		&p.Weight, &p.Length, &p.Width, &p.Height, &p.MetaTitle, &p.MetaDescription, &p.OgImage,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := r.loadProductAssociations(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) FindByID(ctx context.Context, id int) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow(ctx,
		`SELECT p.id, p.name, p.slug, p.description, p.price, p.compare_at_price, p.images, p.status,
		 p.weight, p.length, p.width, p.height, p.meta_title, p.meta_description, p.og_image,
		 p.created_at, p.updated_at
		 FROM products p WHERE p.id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.CompareAtPrice, &p.Images, &p.Status,
		&p.Weight, &p.Length, &p.Width, &p.Height, &p.MetaTitle, &p.MetaDescription, &p.OgImage,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := r.loadProductAssociations(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) loadProductAssociations(ctx context.Context, p *models.Product) error {
	categories, err := r.GetCategories(ctx, p.ID)
	if err != nil {
		return err
	}
	p.Categories = categories

	tags, err := r.GetTags(ctx, p.ID)
	if err != nil {
		return err
	}
	p.Tags = tags

	imageOrder, err := r.GetImageOrder(ctx, p.ID)
	if err != nil {
		return err
	}
	if len(imageOrder) > 0 {
		for _, io := range imageOrder {
			if io.IsPrimary {
				p.PrimaryImage = io.ImagePath
				break
			}
		}
	}

	return nil
}

func (r *productRepository) Create(ctx context.Context, p *models.CreateProductRequest, slug string) (*models.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var product models.Product
	err = tx.QueryRow(ctx,
		`INSERT INTO products (name, slug, description, price, compare_at_price, images, status,
		 weight, length, width, height, meta_title, meta_description, og_image)
		 VALUES ($1, $2, $3, $4, $5, $6, 'published', $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, name, slug, description, price, compare_at_price, images, status,
		 weight, length, width, height, meta_title, meta_description, og_image, created_at, updated_at`,
		p.Name, slug, p.Description, p.Price, p.CompareAtPrice, p.Images,
		p.Weight, p.Length, p.Width, p.Height, p.MetaTitle, p.MetaDescription, p.OgImage,
	).Scan(&product.ID, &product.Name, &product.Slug, &product.Description, &product.Price, &product.CompareAtPrice,
		&product.Images, &product.Status, &product.Weight, &product.Length, &product.Width, &product.Height,
		&product.MetaTitle, &product.MetaDescription, &product.OgImage,
		&product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if len(p.CategoryIDs) > 0 {
		for _, catID := range p.CategoryIDs {
			if _, err := tx.Exec(ctx,
				"INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2)",
				product.ID, catID); err != nil {
				return nil, fmt.Errorf("link category: %w", err)
			}
		}
	}

	if len(p.Tags) > 0 {
		for _, tag := range p.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO product_tags (product_id, tag) VALUES ($1, $2)
				 ON CONFLICT (product_id, tag) DO NOTHING`,
				product.ID, strings.ToLower(tag)); err != nil {
				return nil, fmt.Errorf("add tag: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	if err := r.loadProductAssociations(ctx, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, id int, p *models.UpdateProductRequest) (*models.Product, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if p.Name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Name)
		argIdx++
	}
	if p.Description != nil {
		setClauses = append(setClauses, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Description)
		argIdx++
	}
	if p.Price != nil {
		setClauses = append(setClauses, "price = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Price)
		argIdx++
	}
	if p.CompareAtPrice != nil {
		setClauses = append(setClauses, "compare_at_price = $"+strconv.Itoa(argIdx))
		args = append(args, *p.CompareAtPrice)
		argIdx++
	}
	if p.Images != nil {
		setClauses = append(setClauses, "images = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Images)
		argIdx++
	}
	if p.Status != nil {
		validStatus := map[string]bool{"draft": true, "published": true, "archived": true}
		if !validStatus[*p.Status] {
			return nil, fmt.Errorf("invalid status: must be draft, published, or archived")
		}
		setClauses = append(setClauses, "status = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Status)
		argIdx++
	}
	if p.Weight != nil {
		setClauses = append(setClauses, "weight = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Weight)
		argIdx++
	}
	if p.Length != nil {
		setClauses = append(setClauses, "length = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Length)
		argIdx++
	}
	if p.Width != nil {
		setClauses = append(setClauses, "width = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Width)
		argIdx++
	}
	if p.Height != nil {
		setClauses = append(setClauses, "height = $"+strconv.Itoa(argIdx))
		args = append(args, *p.Height)
		argIdx++
	}
	if p.MetaTitle != nil {
		setClauses = append(setClauses, "meta_title = $"+strconv.Itoa(argIdx))
		args = append(args, *p.MetaTitle)
		argIdx++
	}
	if p.MetaDescription != nil {
		setClauses = append(setClauses, "meta_description = $"+strconv.Itoa(argIdx))
		args = append(args, *p.MetaDescription)
		argIdx++
	}
	if p.OgImage != nil {
		setClauses = append(setClauses, "og_image = $"+strconv.Itoa(argIdx))
		args = append(args, *p.OgImage)
		argIdx++
	}

	if len(setClauses) == 0 && p.CategoryIDs == nil && p.Tags == nil {
		return r.FindByID(ctx, id)
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = NOW()")
		args = append(args, id)
		query := `UPDATE products SET ` + strings.Join(setClauses, ", ") +
			` WHERE id = $` + strconv.Itoa(argIdx) +
			` RETURNING id, name, slug, description, price, compare_at_price, images, status,
			weight, length, width, height, meta_title, meta_description, og_image, created_at, updated_at`

		var product models.Product
		err := r.db.QueryRow(ctx, query, args...).Scan(
			&product.ID, &product.Name, &product.Slug, &product.Description,
			&product.Price, &product.CompareAtPrice, &product.Images, &product.Status,
			&product.Weight, &product.Length, &product.Width, &product.Height,
			&product.MetaTitle, &product.MetaDescription, &product.OgImage,
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			if err == ErrNotFound || err.Error() == "no rows in result set" {
				return nil, ErrNotFound
			}
			return nil, err
		}
	}

	if p.CategoryIDs != nil {
		if err := r.LinkCategories(ctx, id, *p.CategoryIDs); err != nil {
			return nil, fmt.Errorf("link categories: %w", err)
		}
	}
	if p.Tags != nil {
		if err := r.syncTags(ctx, id, *p.Tags); err != nil {
			return nil, fmt.Errorf("sync tags: %w", err)
		}
	}

	return r.FindByID(ctx, id)
}

func (r *productRepository) SoftDelete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx,
		"UPDATE products SET status = 'archived', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) List(ctx context.Context, filters ProductFilters) ([]models.Product, int, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	where := []string{"p.status = 'published'"}
	args := []interface{}{}
	argIdx := 1

	if len(filters.CategoryIDs) > 0 {
		where = append(where, "pc.category_id = ANY($"+strconv.Itoa(argIdx)+")")
		args = append(args, filters.CategoryIDs)
		argIdx++
	}
	if len(filters.Tags) > 0 {
		where = append(where, "EXISTS (SELECT 1 FROM product_tags pt WHERE pt.product_id = p.id AND pt.tag = ANY($"+strconv.Itoa(argIdx)+"))")
		args = append(args, filters.Tags)
		argIdx++
	}
	if filters.MinPrice != nil {
		where = append(where, "p.price >= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.MinPrice)
		argIdx++
	}
	if filters.MaxPrice != nil {
		where = append(where, "p.price <= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.MaxPrice)
		argIdx++
	}
	if filters.Search != "" {
		where = append(where, "(p.name ILIKE $"+strconv.Itoa(argIdx)+" OR p.description ILIKE $"+strconv.Itoa(argIdx)+")")
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	joinClause := ""
	if len(filters.CategoryIDs) > 0 {
		joinClause = " INNER JOIN product_categories pc ON p.id = pc.product_id"
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(DISTINCT p.id) FROM products p" + joinClause + " WHERE " + whereClause
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sort := "p.created_at DESC"
	switch filters.Sort {
	case "price_asc":
		sort = "p.price ASC"
	case "price_desc":
		sort = "p.price DESC"
	case "name":
		sort = "p.name ASC"
	case "newest":
		sort = "p.created_at DESC"
	case "oldest":
		sort = "p.created_at ASC"
	default:
		sort = "p.created_at DESC"
	}

	selectArgs := make([]interface{}, len(args), len(args)+2)
	copy(selectArgs, args)
	selectArgs = append(selectArgs, filters.Limit, filters.Offset)

	selectQuery := `SELECT DISTINCT p.id, p.name, p.slug, p.description, p.price, p.compare_at_price, p.images, p.status,
		p.weight, p.length, p.width, p.height, p.meta_title, p.meta_description, p.og_image,
		p.created_at, p.updated_at
		FROM products p` + joinClause + ` WHERE ` + whereClause + ` ORDER BY ` + sort + ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.CompareAtPrice, &p.Images, &p.Status,
			&p.Weight, &p.Length, &p.Width, &p.Height, &p.MetaTitle, &p.MetaDescription, &p.OgImage,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if err := r.loadProductAssociations(ctx, &p); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) AdminList(ctx context.Context, filters ProductFilters) ([]models.Product, int, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if filters.Status != "" {
		where = append(where, "p.status = $"+strconv.Itoa(argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if len(filters.CategoryIDs) > 0 {
		where = append(where, "pc.category_id = ANY($"+strconv.Itoa(argIdx)+")")
		args = append(args, filters.CategoryIDs)
		argIdx++
	}
	if len(filters.Tags) > 0 {
		where = append(where, "EXISTS (SELECT 1 FROM product_tags pt WHERE pt.product_id = p.id AND pt.tag = ANY($"+strconv.Itoa(argIdx)+"))")
		args = append(args, filters.Tags)
		argIdx++
	}
	if filters.MinPrice != nil {
		where = append(where, "p.price >= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.MinPrice)
		argIdx++
	}
	if filters.MaxPrice != nil {
		where = append(where, "p.price <= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.MaxPrice)
		argIdx++
	}
	if filters.Search != "" {
		where = append(where, "(p.name ILIKE $"+strconv.Itoa(argIdx)+" OR p.description ILIKE $"+strconv.Itoa(argIdx)+")")
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	joinClause := ""
	if len(filters.CategoryIDs) > 0 {
		joinClause = " INNER JOIN product_categories pc ON p.id = pc.product_id"
	}

	whereClause := "TRUE"
	if len(where) > 0 {
		whereClause = strings.Join(where, " AND ")
	}

	var total int
	countQuery := "SELECT COUNT(DISTINCT p.id) FROM products p" + joinClause + " WHERE " + whereClause
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sort := "p.created_at DESC"
	switch filters.Sort {
	case "price_asc":
		sort = "p.price ASC"
	case "price_desc":
		sort = "p.price DESC"
	case "name":
		sort = "p.name ASC"
	case "newest":
		sort = "p.created_at DESC"
	case "oldest":
		sort = "p.created_at ASC"
	default:
		sort = "p.created_at DESC"
	}

	selectArgs := make([]interface{}, len(args), len(args)+2)
	copy(selectArgs, args)
	selectArgs = append(selectArgs, filters.Limit, filters.Offset)

	selectQuery := `SELECT DISTINCT p.id, p.name, p.slug, p.description, p.price, p.compare_at_price, p.images, p.status,
		p.weight, p.length, p.width, p.height, p.meta_title, p.meta_description, p.og_image,
		p.created_at, p.updated_at
		FROM products p` + joinClause + ` WHERE ` + whereClause + ` ORDER BY ` + sort + ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.CompareAtPrice, &p.Images, &p.Status,
			&p.Weight, &p.Length, &p.Width, &p.Height, &p.MetaTitle, &p.MetaDescription, &p.OgImage,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if err := r.loadProductAssociations(ctx, &p); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) GetImages(ctx context.Context, productID int) ([]string, error) {
	var images []string
	err := r.db.QueryRow(ctx,
		"SELECT images FROM products WHERE id = $1", productID).Scan(&images)
	return images, err
}

func (r *productRepository) RemoveImageFromProduct(ctx context.Context, productID int, imagePath string) error {
	result, err := r.db.Exec(ctx,
		"UPDATE products SET images = array_remove(images, $1), updated_at = NOW() WHERE id = $2",
		imagePath, productID)
	if err != nil {
		return fmt.Errorf("remove image from product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) AddImageToProduct(ctx context.Context, productID int, imagePath string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT $1 = ANY(images) FROM products WHERE id = $2", imagePath, productID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check duplicate image: %w", err)
	}
	if exists {
		return nil
	}

	result, err := tx.Exec(ctx,
		"UPDATE products SET images = array_append(images, $1), updated_at = NOW() WHERE id = $2",
		imagePath, productID)
	if err != nil {
		return fmt.Errorf("add image to product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	var count int
	err = tx.QueryRow(ctx, "SELECT array_length(images, 1) FROM products WHERE id = $1", productID).Scan(&count)
	if err == nil && count == 1 {
		_, _ = tx.Exec(ctx,
			`INSERT INTO product_image_order (product_id, image_path, sort_order, is_primary)
			 VALUES ($1, $2, 0, true)
			 ON CONFLICT (product_id, image_path) DO UPDATE SET is_primary = true`,
			productID, imagePath)
	}

	return tx.Commit(ctx)
}

func (r *productRepository) GetVariants(ctx context.Context, productID int) ([]models.ProductVariant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, stock, sku, weight, length, width, height
		 FROM product_variants WHERE product_id = $1 ORDER BY id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []models.ProductVariant
	for rows.Next() {
		var v models.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Stock, &v.SKU,
			&v.Weight, &v.Length, &v.Width, &v.Height); err != nil {
			return nil, err
		}
		optionValues, err := r.getVariantOptionValues(ctx, v.ID)
		if err != nil {
			return nil, err
		}
		v.OptionValues = optionValues
		v.Label = buildVariantLabel(optionValues)
		variants = append(variants, v)
	}
	return variants, nil
}

func (r *productRepository) CountVariants(ctx context.Context, productID int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_variants WHERE product_id = $1`, productID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *productRepository) GetVariantProductID(ctx context.Context, variantID int) (int, error) {
	var productID int
	err := r.db.QueryRow(ctx,
		`SELECT product_id FROM product_variants WHERE id = $1`, variantID,
	).Scan(&productID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return productID, nil
}

func (r *productRepository) getVariantOptionValues(ctx context.Context, variantID int) ([]models.VariantOptionValue, error) {
	rows, err := r.db.Query(ctx,
		`SELECT vov.option_value_id, pov.value, pog.id, pog.name
		 FROM variant_option_values vov
		 JOIN product_option_values pov ON vov.option_value_id = pov.id
		 JOIN product_option_groups pog ON pov.group_id = pog.id
		 WHERE vov.variant_id = $1
		 ORDER BY pog.sort_order, pov.id`, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.VariantOptionValue
	for rows.Next() {
		var ov models.VariantOptionValue
		if err := rows.Scan(&ov.OptionValueID, &ov.Value, &ov.OptionGroupID, &ov.OptionGroupName); err != nil {
			return nil, err
		}
		result = append(result, ov)
	}
	return result, nil
}

func buildVariantLabel(optionValues []models.VariantOptionValue) string {
	if len(optionValues) == 0 {
		return ""
	}
	parts := make([]string, len(optionValues))
	for i, ov := range optionValues {
		parts[i] = ov.Value
	}
	return strings.Join(parts, " / ")
}

func (r *productRepository) GetReviewStats(ctx context.Context, productID int) (float64, int, error) {
	var avgRating float64
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*)
		 FROM reviews WHERE product_id = $1`, productID,
	).Scan(&avgRating, &count)
	return avgRating, count, err
}

func (r *productRepository) LinkCategories(ctx context.Context, productID int, categoryIDs []int) error {
	if len(categoryIDs) == 0 {
		_, err := r.db.Exec(ctx, "DELETE FROM product_categories WHERE product_id = $1", productID)
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM product_categories WHERE product_id = $1", productID); err != nil {
		return err
	}

	for _, catID := range categoryIDs {
		if _, err := tx.Exec(ctx,
			"INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2)",
			productID, catID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *productRepository) GetCategories(ctx context.Context, productID int) ([]models.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.name, c.created_at
		 FROM categories c
		 INNER JOIN product_categories pc ON c.id = pc.category_id
		 WHERE pc.product_id = $1
		 ORDER BY c.name ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// Variant CRUD (Story 2)

func (r *productRepository) CreateVariant(ctx context.Context, productID int, v *models.CreateVariantRequest) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.QueryRow(ctx,
		`INSERT INTO product_variants (product_id, stock, sku, weight, length, width, height)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, product_id, stock, sku, weight, length, width, height`,
		productID, v.Stock, v.SKU, v.Weight, v.Length, v.Width, v.Height,
	).Scan(&variant.ID, &variant.ProductID, &variant.Stock, &variant.SKU,
		&variant.Weight, &variant.Length, &variant.Width, &variant.Height)
	if err != nil {
		return nil, err
	}

	for _, ov := range v.OptionValues {
		_, err := r.db.Exec(ctx,
			`INSERT INTO variant_option_values (variant_id, option_value_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			variant.ID, ov.OptionValueID)
		if err != nil {
			return nil, err
		}
	}

	optionValues, err := r.getVariantOptionValues(ctx, variant.ID)
	if err != nil {
		return nil, err
	}
	variant.OptionValues = optionValues
	variant.Label = buildVariantLabel(optionValues)

	return &variant, nil
}

func (r *productRepository) UpdateVariant(ctx context.Context, variantID int, v *models.UpdateVariantRequest) (*models.ProductVariant, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if v.Stock != nil {
		setClauses = append(setClauses, "stock = $"+strconv.Itoa(argIdx))
		args = append(args, *v.Stock)
		argIdx++
	}
	if v.SKU != nil {
		setClauses = append(setClauses, "sku = $"+strconv.Itoa(argIdx))
		args = append(args, *v.SKU)
		argIdx++
	}
	if v.Weight != nil {
		setClauses = append(setClauses, "weight = $"+strconv.Itoa(argIdx))
		args = append(args, *v.Weight)
		argIdx++
	}
	if v.Length != nil {
		setClauses = append(setClauses, "length = $"+strconv.Itoa(argIdx))
		args = append(args, *v.Length)
		argIdx++
	}
	if v.Width != nil {
		setClauses = append(setClauses, "width = $"+strconv.Itoa(argIdx))
		args = append(args, *v.Width)
		argIdx++
	}
	if v.Height != nil {
		setClauses = append(setClauses, "height = $"+strconv.Itoa(argIdx))
		args = append(args, *v.Height)
		argIdx++
	}

	if len(setClauses) > 0 {
		args = append(args, variantID)
		query := `UPDATE product_variants SET ` + strings.Join(setClauses, ", ") +
			` WHERE id = $` + strconv.Itoa(argIdx) +
			` RETURNING id, product_id, stock, sku, weight, length, width, height`

		var variant models.ProductVariant
		err := r.db.QueryRow(ctx, query, args...).Scan(
			&variant.ID, &variant.ProductID, &variant.Stock, &variant.SKU,
			&variant.Weight, &variant.Length, &variant.Width, &variant.Height,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("update variant: %w", err)
		}
	}

	if v.OptionValues != nil {
		tx, err := r.db.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, "DELETE FROM variant_option_values WHERE variant_id = $1", variantID); err != nil {
			return nil, err
		}
		for _, ov := range v.OptionValues {
			if _, err := tx.Exec(ctx,
				`INSERT INTO variant_option_values (variant_id, option_value_id)
				 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				variantID, ov.OptionValueID); err != nil {
				return nil, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}

	var variant models.ProductVariant
	err := r.db.QueryRow(ctx,
		`SELECT id, product_id, stock, sku, weight, length, width, height
		 FROM product_variants WHERE id = $1`, variantID,
	).Scan(&variant.ID, &variant.ProductID, &variant.Stock, &variant.SKU,
		&variant.Weight, &variant.Length, &variant.Width, &variant.Height)
	if err != nil {
		return nil, ErrNotFound
	}

	optionValues, err := r.getVariantOptionValues(ctx, variantID)
	if err != nil {
		return nil, err
	}
	variant.OptionValues = optionValues
	variant.Label = buildVariantLabel(optionValues)

	return &variant, nil
}

func (r *productRepository) DeleteVariant(ctx context.Context, variantID int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM product_variants WHERE id = $1", variantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) ValidateOptionValues(ctx context.Context, productID int, optionValueIDs []int) error {
	// Check all option_value_ids belong to this product's option groups
	if len(optionValueIDs) > 0 {
		rows, err := r.db.Query(ctx,
			`SELECT pov.id FROM product_option_values pov
			 JOIN product_option_groups pog ON pov.group_id = pog.id
			 WHERE pog.product_id = $1 AND pov.id = ANY($2)`,
			productID, optionValueIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		validIDs := make(map[int]bool)
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return err
			}
			validIDs[id] = true
		}

		for _, id := range optionValueIDs {
			if !validIDs[id] {
				return fmt.Errorf("option value %d does not belong to this product", id)
			}
		}
	}

	// Fetch this product's option groups
	groupRows, err := r.db.Query(ctx,
		`SELECT pog.id FROM product_option_groups pog
		 WHERE pog.product_id = $1
		 ORDER BY pog.sort_order`, productID)
	if err != nil {
		return err
	}
	defer groupRows.Close()

	var groupIDs []int
	for groupRows.Next() {
		var gid int
		if err := groupRows.Scan(&gid); err != nil {
			return err
		}
		groupIDs = append(groupIDs, gid)
	}

	if len(groupIDs) == 0 {
		if len(optionValueIDs) > 0 {
			return fmt.Errorf("this product has no option groups")
		}
		return nil
	}

	// Map each option value to its group
	valueToGroup := make(map[int]int)
	valRows, err := r.db.Query(ctx,
		`SELECT id, group_id FROM product_option_values WHERE group_id = ANY($1)`,
		groupIDs)
	if err != nil {
		return err
	}
	defer valRows.Close()

	for valRows.Next() {
		var vid, gid int
		if err := valRows.Scan(&vid, &gid); err != nil {
			return err
		}
		valueToGroup[vid] = gid
	}

	// Check exactly one value per option group: no group gets two values...
	selectedGroups := make(map[int]bool)
	for _, vid := range optionValueIDs {
		gid := valueToGroup[vid]
		if selectedGroups[gid] {
			return fmt.Errorf("multiple values selected for the same option group")
		}
		selectedGroups[gid] = true
	}

	// ...and every group gets exactly one value.
	if len(selectedGroups) != len(groupIDs) {
		return fmt.Errorf("must select exactly one value for each of this product's %d option group(s)", len(groupIDs))
	}

	return nil
}

func (r *productRepository) HasVariantCombination(ctx context.Context, productID int, optionValueIDs []int, excludeVariantID int) (bool, error) {
	if len(optionValueIDs) == 0 {
		return false, nil
	}

	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT pv.id)
		 FROM product_variants pv
		 JOIN variant_option_values vov ON pv.id = vov.variant_id
		 WHERE pv.product_id = $1 AND pv.id != $4 AND vov.option_value_id = ANY($2)
		 GROUP BY pv.id
		 HAVING COUNT(DISTINCT vov.option_value_id) = $3`,
		productID, optionValueIDs, len(optionValueIDs), excludeVariantID,
	).Scan(&count)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

// Tags (Story 3)

func (r *productRepository) GetTags(ctx context.Context, productID int) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tag FROM product_tags WHERE product_id = $1 ORDER BY tag`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *productRepository) AddTags(ctx context.Context, productID int, tags []string) error {
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		_, err := r.db.Exec(ctx,
			`INSERT INTO product_tags (product_id, tag) VALUES ($1, $2)
			 ON CONFLICT (product_id, tag) DO NOTHING`,
			productID, strings.ToLower(tag))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *productRepository) RemoveTag(ctx context.Context, productID int, tag string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM product_tags WHERE product_id = $1 AND tag = $2`,
		productID, strings.ToLower(tag))
	return err
}

func (r *productRepository) syncTags(ctx context.Context, productID int, tags []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM product_tags WHERE product_id = $1", productID); err != nil {
		return err
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_tags (product_id, tag) VALUES ($1, $2)
			 ON CONFLICT (product_id, tag) DO NOTHING`,
			productID, strings.ToLower(tag)); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Image ordering (Story 5)

func (r *productRepository) GetImageOrder(ctx context.Context, productID int) ([]models.ProductImageOrder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, image_path, sort_order, is_primary
		 FROM product_image_order WHERE product_id = $1
		 ORDER BY sort_order, id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.ProductImageOrder
	for rows.Next() {
		var o models.ProductImageOrder
		if err := rows.Scan(&o.ID, &o.ProductID, &o.ImagePath, &o.SortOrder, &o.IsPrimary); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *productRepository) UpdateImageOrder(ctx context.Context, productID int, images []models.ImageOrderEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM product_image_order WHERE product_id = $1", productID); err != nil {
		return err
	}

	for _, img := range images {
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_image_order (product_id, image_path, sort_order, is_primary)
			 VALUES ($1, $2, $3, $4)`,
			productID, img.ImagePath, img.SortOrder, img.IsPrimary); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Option Groups & Values (Story 2)

func (r *productRepository) GetOptionGroupsByProduct(ctx context.Context, productID int) ([]models.ProductOptionGroup, error) {
	groups, err := r.getOptionGroups(ctx, productID)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		values, err := r.getOptionValuesByGroup(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Values = values
	}
	return groups, nil
}

func (r *productRepository) getOptionGroups(ctx context.Context, productID int) ([]models.ProductOptionGroup, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, name, sort_order
		 FROM product_option_groups WHERE product_id = $1
		 ORDER BY sort_order, id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.ProductOptionGroup
	for rows.Next() {
		var g models.ProductOptionGroup
		if err := rows.Scan(&g.ID, &g.ProductID, &g.Name, &g.SortOrder); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *productRepository) getOptionValuesByGroup(ctx context.Context, groupID int) ([]models.ProductOptionValue, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, group_id, value, price_modifier
		 FROM product_option_values WHERE group_id = $1
		 ORDER BY id`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []models.ProductOptionValue
	for rows.Next() {
		var v models.ProductOptionValue
		if err := rows.Scan(&v.ID, &v.GroupID, &v.Value, &v.PriceModifier); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func (r *productRepository) CreateOptionGroup(ctx context.Context, productID int, req *models.CreateOptionGroupRequest) (*models.ProductOptionGroup, error) {
	var group models.ProductOptionGroup
	err := r.db.QueryRow(ctx,
		`INSERT INTO product_option_groups (product_id, name, sort_order)
		 VALUES ($1, $2, 0)
		 RETURNING id, product_id, name, sort_order`,
		productID, req.Name,
	).Scan(&group.ID, &group.ProductID, &group.Name, &group.SortOrder)
	if err != nil {
		return nil, err
	}

	var values []models.ProductOptionValue
	for _, v := range req.Values {
		var ov models.ProductOptionValue
		pm := 0
		if v.PriceModifier != nil {
			pm = *v.PriceModifier
		}
		err := r.db.QueryRow(ctx,
			`INSERT INTO product_option_values (group_id, value, price_modifier)
			 VALUES ($1, $2, $3)
			 RETURNING id, group_id, value, price_modifier`,
			group.ID, v.Value, pm,
		).Scan(&ov.ID, &ov.GroupID, &ov.Value, &ov.PriceModifier)
		if err != nil {
			return nil, err
		}
		values = append(values, ov)
	}
	group.Values = values
	return &group, nil
}

func (r *productRepository) UpdateOptionGroup(ctx context.Context, groupID int, req *models.UpdateOptionGroupRequest) (*models.ProductOptionGroup, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.SortOrder != nil {
		setClauses = append(setClauses, "sort_order = $"+strconv.Itoa(argIdx))
		args = append(args, *req.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.getOptionGroupByID(ctx, groupID)
	}

	args = append(args, groupID)
	query := `UPDATE product_option_groups SET ` + strings.Join(setClauses, ", ") +
		` WHERE id = $` + strconv.Itoa(argIdx) +
		` RETURNING id, product_id, name, sort_order`

	var group models.ProductOptionGroup
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&group.ID, &group.ProductID, &group.Name, &group.SortOrder,
	)
	if err != nil {
		return nil, ErrNotFound
	}

	values, err := r.getOptionValuesByGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.Values = values
	return &group, nil
}

func (r *productRepository) getOptionGroupByID(ctx context.Context, groupID int) (*models.ProductOptionGroup, error) {
	var group models.ProductOptionGroup
	err := r.db.QueryRow(ctx,
		`SELECT id, product_id, name, sort_order
		 FROM product_option_groups WHERE id = $1`, groupID,
	).Scan(&group.ID, &group.ProductID, &group.Name, &group.SortOrder)
	if err != nil {
		return nil, ErrNotFound
	}
	values, err := r.getOptionValuesByGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.Values = values
	return &group, nil
}

func (r *productRepository) DeleteOptionGroup(ctx context.Context, groupID int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM product_option_groups WHERE id = $1", groupID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) CreateOptionValue(ctx context.Context, groupID int, req *models.CreateOptionValueRequest) (*models.ProductOptionValue, error) {
	pm := 0
	if req.PriceModifier != nil {
		pm = *req.PriceModifier
	}
	var v models.ProductOptionValue
	err := r.db.QueryRow(ctx,
		`INSERT INTO product_option_values (group_id, value, price_modifier)
		 VALUES ($1, $2, $3)
		 RETURNING id, group_id, value, price_modifier`,
		groupID, req.Value, pm,
	).Scan(&v.ID, &v.GroupID, &v.Value, &v.PriceModifier)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *productRepository) UpdateOptionValue(ctx context.Context, valueID int, req *models.UpdateOptionValueRequest) (*models.ProductOptionValue, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Value != nil {
		setClauses = append(setClauses, "value = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Value)
		argIdx++
	}
	if req.PriceModifier != nil {
		setClauses = append(setClauses, "price_modifier = $"+strconv.Itoa(argIdx))
		args = append(args, *req.PriceModifier)
		argIdx++
	}
	if req.SortOrder != nil {
		setClauses = append(setClauses, "sort_order = $"+strconv.Itoa(argIdx))
		args = append(args, *req.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		var v models.ProductOptionValue
		err := r.db.QueryRow(ctx,
			`SELECT id, group_id, value, price_modifier
			 FROM product_option_values WHERE id = $1`, valueID,
		).Scan(&v.ID, &v.GroupID, &v.Value, &v.PriceModifier)
		if err != nil {
			return nil, ErrNotFound
		}
		return &v, nil
	}

	args = append(args, valueID)
	query := `UPDATE product_option_values SET ` + strings.Join(setClauses, ", ") +
		` WHERE id = $` + strconv.Itoa(argIdx) +
		` RETURNING id, group_id, value, price_modifier`

	var v models.ProductOptionValue
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &v.GroupID, &v.Value, &v.PriceModifier,
	)
	if err != nil {
		return nil, ErrNotFound
	}
	return &v, nil
}

func (r *productRepository) DeleteOptionValue(ctx context.Context, valueID int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM product_option_values WHERE id = $1", valueID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) ReorderOptionValues(ctx context.Context, groupID int, valueIDs []int) error {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_option_values WHERE id = ANY($1) AND group_id = $2`,
		valueIDs, groupID).Scan(&count)
	if err != nil {
		return fmt.Errorf("validate option values: %w", err)
	}
	if count != len(valueIDs) {
		return fmt.Errorf("one or more option values do not belong to this group")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, valueID := range valueIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE product_option_values SET sort_order = $1 WHERE id = $2 AND group_id = $3`,
			i, valueID, groupID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Customization Fields (Story 5)

func (r *productRepository) GetCustomizationFieldsByProduct(ctx context.Context, productID int) ([]models.ProductCustomizationField, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, name, field_type, required, default_value, sort_order
		 FROM product_customization_fields WHERE product_id = $1
		 ORDER BY sort_order, id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.ProductCustomizationField
	for rows.Next() {
		var f models.ProductCustomizationField
		if err := rows.Scan(&f.ID, &f.ProductID, &f.Name, &f.FieldType, &f.Required, &f.DefaultValue, &f.SortOrder); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func (r *productRepository) CreateCustomizationField(ctx context.Context, productID int, req *models.CreateCustomizationFieldRequest) (*models.ProductCustomizationField, error) {
	required := false
	if req.Required != nil {
		required = *req.Required
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	var f models.ProductCustomizationField
	err := r.db.QueryRow(ctx,
		`INSERT INTO product_customization_fields (product_id, name, field_type, required, default_value, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, product_id, name, field_type, required, default_value, sort_order`,
		productID, req.Name, req.FieldType, required, req.DefaultValue, sortOrder,
	).Scan(&f.ID, &f.ProductID, &f.Name, &f.FieldType, &f.Required, &f.DefaultValue, &f.SortOrder)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *productRepository) UpdateCustomizationField(ctx context.Context, fieldID int, req *models.UpdateCustomizationFieldRequest) (*models.ProductCustomizationField, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.FieldType != nil {
		setClauses = append(setClauses, "field_type = $"+strconv.Itoa(argIdx))
		args = append(args, *req.FieldType)
		argIdx++
	}
	if req.Required != nil {
		setClauses = append(setClauses, "required = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Required)
		argIdx++
	}
	if req.DefaultValue != nil {
		setClauses = append(setClauses, "default_value = $"+strconv.Itoa(argIdx))
		args = append(args, *req.DefaultValue)
		argIdx++
	}
	if req.SortOrder != nil {
		setClauses = append(setClauses, "sort_order = $"+strconv.Itoa(argIdx))
		args = append(args, *req.SortOrder)
		argIdx++
	}

	if len(setClauses) == 0 {
		var f models.ProductCustomizationField
		err := r.db.QueryRow(ctx,
			`SELECT id, product_id, name, field_type, required, default_value, sort_order
			 FROM product_customization_fields WHERE id = $1`, fieldID,
		).Scan(&f.ID, &f.ProductID, &f.Name, &f.FieldType, &f.Required, &f.DefaultValue, &f.SortOrder)
		if err != nil {
			return nil, ErrNotFound
		}
		return &f, nil
	}

	args = append(args, fieldID)
	query := `UPDATE product_customization_fields SET ` + strings.Join(setClauses, ", ") +
		` WHERE id = $` + strconv.Itoa(argIdx) +
		` RETURNING id, product_id, name, field_type, required, default_value, sort_order`

	var f models.ProductCustomizationField
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&f.ID, &f.ProductID, &f.Name, &f.FieldType, &f.Required, &f.DefaultValue, &f.SortOrder,
	)
	if err != nil {
		return nil, ErrNotFound
	}
	return &f, nil
}

func (r *productRepository) DeleteCustomizationField(ctx context.Context, fieldID int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM product_customization_fields WHERE id = $1", fieldID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *productRepository) ReorderCustomizationFields(ctx context.Context, productID int, fieldIDs []int) error {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_customization_fields WHERE id = ANY($1) AND product_id = $2`,
		fieldIDs, productID).Scan(&count)
	if err != nil {
		return fmt.Errorf("validate customization fields: %w", err)
	}
	if count != len(fieldIDs) {
		return fmt.Errorf("one or more fields do not belong to this product")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, fieldID := range fieldIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE product_customization_fields SET sort_order = $1 WHERE id = $2 AND product_id = $3`,
			i, fieldID, productID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
