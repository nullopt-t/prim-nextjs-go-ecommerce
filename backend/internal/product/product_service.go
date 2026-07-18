package product

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/tag"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/types"

	"github.com/google/uuid"
)

type ProductService struct {
	dbExecuter      database.Runner
	productRepo     *ProductRepository
	brandService    *brand.BrandService
	categoryService *category.CategoryService
	tagService      *tag.TagService
	variantService  *variant.VariantService
}

func NewService(r database.Runner,
	productRepo *ProductRepository,
	brandService *brand.BrandService,
	categoryService *category.CategoryService,
	tagService *tag.TagService,
	variantService *variant.VariantService,
) *ProductService {
	return &ProductService{
		dbExecuter:      r,
		productRepo:     productRepo,
		brandService:    brandService,
		categoryService: categoryService,
		tagService:      tagService,
		variantService:  variantService,
	}
}

type CreateProductInput struct {
	BrandID          *uuid.UUID
	Title            string
	ShortDescription string
	Description      string
	Slug             string
	Status           model.ProductStatus
	CreatedBy        uuid.UUID
}

func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {

	now := time.Now().UTC()
	product := &model.Product{
		ID:               uuid.New(),
		BrandID:          input.BrandID,
		Title:            input.Title,
		ShortDescription: input.ShortDescription,
		Description:      input.Description,
		Slug:             input.Slug,
		Status:           model.ProductStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			err := s.productRepo.Create(ctx, db, product)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrConflict,
		):
			return nil, security.NewSecureError(
				http.StatusConflict,
				security.CodeConflict,
				"product already exists",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to create a new product",
				err,
			)
		}
	}

	return product, nil
}

type ProductDetails struct {
	Product *model.Product
	Brand   *model.ProductBrand
}

func (s *ProductService) GetByID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.Product, error) {
	product := &model.Product{}
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			prod, err := s.productRepo.Get(ctx, db, Filter{
				ID: types.Ptr(productID),
			})
			if err != nil {
				return err
			}
			product = prod
			return nil
		},
	)
	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(
			mappedError,
			database.ErrNotFound,
		):
			return nil, security.NewSecureError(
				http.StatusNotFound,
				security.CodeNotFound,
				"product not found",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch a product",
				err,
			)
		}
	}

	return product, nil
}

func (s *ProductService) GetBySlug(
	ctx context.Context,
	slug string,
) (*ProductDetails, error) {
	productDetails := &ProductDetails{}
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			prod, err := s.productRepo.Get(ctx, db, Filter{
				Slug: types.Ptr(slug),
			})
			if err != nil {
				return err
			}
			productDetails.Product = prod
			return nil
		},
	)
	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(
			mappedError,
			database.ErrNotFound,
		):
			return nil, security.NewSecureError(
				http.StatusNotFound,
				security.CodeNotFound,
				"product not found",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch a product",
				err,
			)
		}
	}

	if productDetails.Product.BrandID != nil {
		brand, err := s.brandService.GetBrandByID(ctx, *productDetails.Product.BrandID)
		if err != nil {
			return nil, err
		}
		productDetails.Brand = brand
	}

	return productDetails, nil
}

func (s *ProductService) List(
	ctx context.Context,
	q *api.ListQuery,
) (*ProductList, error) {
	var res *ProductList
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			list, err := s.productRepo.List(ctx, db, q)
			if err != nil {
				return err
			}
			res = list
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch products",
			err,
		)
	}
	return res, nil
}

func (s *ProductService) GetProductVariants(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductVariant, error) {
	var variants []*model.ProductVariant
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			variants, err = s.productRepo.GetVariants(
				ctx,
				db,
				productID,
			)

			return err
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch product variants",
			err,
		)
	}
	return variants, nil
}

func (s *ProductService) GetProductCategories(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductCategory, error) {
	var categories []*model.ProductCategory
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			categories, err = s.categoryService.ListProductCategories(
				ctx,
				productID,
			)
			return err
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch product categories",
			err,
		)
	}
	return categories, nil
}

func (s *ProductService) PutProductCategories(
	ctx context.Context,
	productID uuid.UUID,
	categoryIDs []uuid.UUID,
) error {
	err := s.categoryService.PutProductCategories(
		ctx,
		productID,
		categoryIDs,
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to put product categories",
			err,
		)
	}
	return nil
}

func (s *ProductService) GetProductTags(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	tags, err := s.tagService.ListProductTags(ctx, productID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *ProductService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return s.productRepo.SetDefaultVariant(
				ctx,
				db,
				productID,
				variantID,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to set default variant",
			err,
		)
	}
	return nil
}

func (s *ProductService) PublishProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return s.productRepo.PublishProduct(
				ctx,
				db,
				productID,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to publish product",
			err,
		)
	}
	return nil
}

func (s *ProductService) ArchiveProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return s.productRepo.ArchiveProduct(
				ctx,
				db,
				productID,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to archive product",
			err,
		)
	}
	return nil
}

func (s *ProductService) PutProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagsIDs []uuid.UUID,
) error {
	err := s.tagService.PutProductTags(
		ctx,
		productID,
		tagsIDs,
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to put product tags",
			err,
		)
	}
	return nil
}

type CreateProductVariantInput struct {
	SKU      *string
	Price    int64
	Currency string
}

func (s *ProductService) CreateProductVariant(
	ctx context.Context,
	productID uuid.UUID,
	input CreateProductVariantInput,
) (*model.ProductVariant, error) {
	return s.variantService.CreateVariant(ctx, variant.CreateVariantInput{
		ProductID: productID,
		SKU:       input.SKU,
		Price:     input.Price,
		Currency:  input.Currency,
	})
}
