package product

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedProductSortFields = map[string]string{
	"id":         "p.id",
	"public_id":  "p.public_id",
	"title":      "p.title",
	"status":     "p.status",
	"created_at": "p.created_at",
	"updated_at": "p.updated_at",
}

type PublicProductBrandReadModel struct {
	ID   uuid.UUID
	Name string
	Link *string
}

type PublicProductCategoryReadModel struct {
	ID   uuid.UUID
	Name string
}


type ProductCardReadModel struct {
	ID              uuid.UUID
	Slug            string
	Title           string
	Description     *string
	Status          model.PublicationStatus
	Brand           *PublicProductBrandReadModel
	Category        *PublicProductCategoryReadModel
	ProductType     model.ProductType
	Thumbnail       *model.Object
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	TagsRaw         []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type CreateProductMediaInput struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	ObjectID  uuid.UUID
	MediaType model.MediaType
	SortOrder int
}

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (r *ProductRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	p *model.Product,
) error {
	query := `
		INSERT INTO products (
			id,
			brand_id,
			category_id,
			slug,
			title,
			status,
			product_type,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		p.ID,
		p.BrandID,
		p.CategoryID,
		p.Slug,
		p.Title,
		p.Status,
		p.ProductType,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, "id = $1 AND deleted_at IS NULL", id)
}

func (r *ProductRepository) GetByIDWithDeleted(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, "id = $1", id)
}

func (r *ProductRepository) GetBySlug(
	ctx context.Context,
	qe database.QueryExecutor,
	slug string,
) (*model.Product, error) {
	return r.get(ctx, qe, "slug = $1 AND deleted_at IS NULL", slug)
}

func (r *ProductRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) (*model.Product, error) {
	if whereClause == "" {
		return nil, errors.New("get product: whereClause is required")
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			brand_id,
			category_id,
			slug,
			title,
			description,
			highlights,
			status,
			product_type,
			thumbnail_object_id,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE %s
		LIMIT 1
	`, whereClause)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}

	product, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Product])
	if err != nil {
		return nil, fmt.Errorf("scan product: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	p *model.Product,
) error {
	if p.ID == uuid.Nil {
		return errors.New("update product: product ID is required")
	}

	query := `
		UPDATE products
		SET
			brand_id = $1,
			category_id = $2,
			slug = $3,
			title = $4,
			description = $5,
			highlights = $6,
			status = $7,
			product_type = $8,
			thumbnail_object_id = $9,
			updated_at = now()
		WHERE id = $10 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(
		ctx,
		query,
		p.BrandID,
		p.CategoryID,
		p.Slug,
		p.Title,
		p.Description,
		p.Highlights,
		p.Status,
		p.ProductType,
		p.ThumbnailObjectID,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
		whereClauses = append(whereClauses, "p.status = 'published'")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.public_id ILIKE $%d OR b.name ILIKE $%d OR c.name ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN product_brands b ON p.brand_id = b.id LEFT JOIN product_categories c ON p.category_id = c.id %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*ProductCardReadModel{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Sorting
	orderBy := "ORDER BY p.created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedProductSortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			direction := "ASC"
			if sort.Order == pagination.SortDesc {
				direction = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", dbField, direction))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	// 3. Paginated Select
	selectQuery := fmt.Sprintf(`
		SELECT
			p.id,
			p.slug,
			p.title,
			p.description,
			p.status,
			p.product_type,
			b.id AS brand_id,
			b.name AS brand_name,
			b.link AS brand_link,
			c.id AS category_id,
			c.name AS category_name,
			so.id as so_id,
			so.bucket as so_bucket,
			so.object_key as so_object_key,
			so.content_type as so_content_type,
			so.file_size as so_file_size,
			pv.price as price,
			pv.crossed_out_price as crossed_out_price,
			pv.currency as currency,
			(
				SELECT coalesce(json_agg(json_build_object('id', pt.public_id, 'name', pt.name)), '[]'::json)
				FROM product_tag_assignments pta
				JOIN product_tags pt ON pta.tag_id = pt.id
				WHERE pta.product_id = p.id
			) as tags_raw,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN storage_objects so ON p.thumbnail_object_id = so.id
		LEFT JOIN product_variants pv ON p.id = pv.product_id AND pv.is_default = true AND pv.deleted_at IS NULL
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list products select: %w", err)
	}
	defer rows.Close()

	items := make([]*ProductCardReadModel, 0, q.PageSize)
	for rows.Next() {
		var (
			item                           ProductCardReadModel
			brandID, catID                 *uuid.UUID
			brandName, brandLink, catName  *string
			soID                           *uuid.UUID
			soBucket, soKey, soContentType *string
			soFileSize                     *int64
		)
		err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.ProductType,
			&brandID,
			&brandName,
			&brandLink,
			&catID,
			&catName,
			&soID,
			&soBucket,
			&soKey,
			&soContentType,
			&soFileSize,
			&item.Price,
			&item.CrossedOutPrice,
			&item.Currency,
			&item.TagsRaw,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product list item: %w", err)
		}

		if brandID != nil && brandName != nil {
			item.Brand = &PublicProductBrandReadModel{
				ID:   *brandID,
				Name: *brandName,
				Link: brandLink,
			}
		}

		if catID != nil && catName != nil {
			item.Category = &PublicProductCategoryReadModel{
				ID:   *catID,
				Name: *catName,
			}
		}

		if soID != nil {
			item.Thumbnail = &model.Object{
				ID:          *soID,
				Bucket:      *soBucket,
				Key:         *soKey,
				ContentType: *soContentType,
				FileSize:    *soFileSize,
			}
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product list items: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

func (r *ProductRepository) AdminList(
	ctx context.Context,
	qe database.QueryExecutor,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.Product], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.slug ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin products: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.Product{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Sorting
	orderBy := "ORDER BY p.created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedProductSortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			direction := "ASC"
			if sort.Order == pagination.SortDesc {
				direction = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", dbField, direction))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	// 3. Paginated Select
	selectQuery := fmt.Sprintf(`
		SELECT
			p.id,
			p.brand_id,
			p.category_id,
			p.slug,
			p.title,
			p.description,
			p.highlights,
			p.status,
			p.product_type,
			p.thumbnail_object_id,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list admin products select: %w", err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.Product])
	if err != nil {
		return nil, fmt.Errorf("collect admin products: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

func (r *ProductRepository) SoftDeleteByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) error {
	return r.softDelete(ctx, qe, "id = $1", id)
}

func (r *ProductRepository) softDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) error {
	if whereClause == "" {
		return errors.New("soft delete product: whereClause is required")
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET deleted_at = now(), updated_at = now()
		WHERE %s
	`, whereClause)

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("soft delete product: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *ProductRepository) UpdateStatus(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	status model.PublicationStatus,
) error {
	if productID == uuid.Nil {
		return errors.New("update product status: productID is required")
	}

	query := `
		UPDATE products
		SET status = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, productID, status)
	if err != nil {
		return fmt.Errorf("update product status: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
