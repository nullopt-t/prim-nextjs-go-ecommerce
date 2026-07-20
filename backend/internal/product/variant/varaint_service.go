package variant

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	fileUtil "github.com/m-mahmoud-alsaid/prim-backend/pkg/file"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/minio/minio-go/v7"
)

const (
	variantMediaBucket = "variant-media"
)

type ObjectService interface {
	CreateObjectWithTx(
		ctx context.Context,
		tx database.QueryExecutor,
		bucket string,
		key string,
		size int64,
		contentType string,
		status model.ObjectStatus,
	) (*model.Object, error)
}

type VariantService struct {
	logger      log.Logger
	dr          database.Runner
	minioClient *minio.Client
	vr          *VariantRepository
	mr          *MediaRepository
	os          ObjectService
	minCfg      *config.MinioConfig
}

func NewService(
	logger log.Logger,
	r database.Runner,
	minioClient *minio.Client,
	vr *VariantRepository,
	mr *MediaRepository,
	os ObjectService,
	minCfg *config.MinioConfig,
) *VariantService {
	return &VariantService{
		logger:      logger,
		dr:          r,
		minioClient: minioClient,
		vr:          vr,
		mr:          mr,
		os:          os,
		minCfg:      minCfg,
	}
}

func (vs *VariantService) GenPublicURL(bucket, key string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		strings.TrimRight(vs.minCfg.PublicURL, "/"),
		bucket,
		strings.TrimLeft(key, "/"),
	)
}

func (vs *VariantService) GenObjectKey(
	variantID uuid.UUID,
	contentType string,
) string {
	return fmt.Sprintf(
		"%s/%s%s",
		variantID,
		uuid.NewString(),
		fileUtil.MimeExtension(contentType),
	)
}

type CreateVariantInput struct {
	ProductID uuid.UUID
	SKU       *string
	Price     int64
	Currency  string
}

func (vs *VariantService) CreateVariant(
	ctx context.Context,
	in CreateVariantInput,
) (*model.ProductVariant, error) {

	now := time.Now().UTC()
	variant := &model.ProductVariant{
		ID:        uuid.New(),
		ProductID: in.ProductID,
		SKU:       in.SKU,
		Price:     in.Price,
		Currency:  in.Currency,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return vs.vr.Create(
				ctx,
				db,
				variant,
			)
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new variant",
			err,
		)
	}
	return variant, nil
}

func (vs *VariantService) AttachMediaObject(
	ctx context.Context,
	VariantID uuid.UUID,
	ObjectID uuid.UUID,
	MediaType model.MediaType,
	SortOrder int,
) (*model.VariantMedia, error) {

	now := time.Now().UTC()
	media := &model.VariantMedia{
		ID:        uuid.New(),
		VariantID: VariantID,
		ObjectID:  ObjectID,
		Type:      MediaType,
		SortOrder: SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return vs.mr.CreateVariantMedia(
				ctx,
				db,
				media,
			)
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create variant media",
			err,
		)
	}
	return media, nil
}

func (vs *VariantService) ReorderVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
	media []uuid.UUID,
) error {
	err := vs.dr.WithTx(
		ctx,
		func(tx database.QueryExecutor) error {
			for i, mediaID := range media {
				err := vs.mr.ReorderVariantMedia(
					ctx,
					tx,
					variantID,
					mediaID,
					i,
				)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to reorder variant media",
			err,
		)
	}
	return nil
}

func (vs *VariantService) AttachMediaObjectWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	variantID uuid.UUID,
	objectID uuid.UUID,
	mediaType model.MediaType,
	sortOrder int,
) error {
	now := time.Now().UTC()
	media := &model.VariantMedia{
		ID:        uuid.New(),
		VariantID: variantID,
		ObjectID:  objectID,
		Type:      mediaType,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := vs.mr.CreateVariantMedia(
		ctx,
		tx,
		media,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VariantService) GetVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	var media []*model.VariantMedia
	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			m, err := vs.mr.GetVariantMedia(
				ctx,
				db,
				variantID,
			)
			if err != nil {
				return err
			}
			media = m
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new variant",
			err,
		)
	}

	for _, m := range media {
		m.Object.PublicURL = vs.GenPublicURL(
			m.Object.Bucket,
			m.Object.Key,
		)
	}
	return media, nil
}

func (vs *VariantService) GetVariant(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.ProductVariant, error) {
	var variant *model.ProductVariant
	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			v, err := vs.vr.GetVariant(
				ctx,
				db,
				variantID,
			)
			if err != nil {
				return err
			}
			variant = v
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return variant, nil
}

func (vs *VariantService) UploadVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
	fileHeader *multipart.FileHeader,
) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return err
	}

	mediaType, err := model.ParseMediaType(detectedContentType)
	if err != nil {
		return security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"invalid media type",
			err,
		)
	}

	vs.logger.Debug("detected content type", log.Meta{
		"ContentType": detectedContentType,
	})

	variant, err := vs.GetVariant(ctx, variantID)
	if err != nil {
		return err
	}

	key := vs.GenObjectKey(variant.ID, detectedContentType)
	vs.logger.Debug("uploading variant media", log.Meta{
		"VariantID": variantID.String(),
		"Key":       key,
	})

	exists, err := vs.minioClient.BucketExists(ctx, variantMediaBucket)
	if err != nil {
		return err
	}

	if !exists {
		err := vs.minioClient.MakeBucket(
			ctx,
			variantMediaBucket,
			minio.MakeBucketOptions{},
		)
		if err != nil {
			return err
		}
	}

	cleanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	info, err := vs.minioClient.PutObject(
		cleanCtx,
		variantMediaBucket,
		key,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{
			ContentType: detectedContentType,
		},
	)
	if err != nil {
		return err
	}

	err = vs.dr.WithTx(
		ctx,
		func(tx database.QueryExecutor) error {
			obj, err := vs.os.CreateObjectWithTx(
				ctx,
				tx,
				info.Bucket,
				key,
				info.Size,
				detectedContentType,
				model.ObjectStatusUploaded,
			)
			if err != nil {
				return err
			}

			err = vs.AttachMediaObjectWithTx(
				ctx,
				tx,
				variantID,
				obj.ID,
				mediaType,
				0,
			)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		_ = vs.minioClient.RemoveObject(
			ctx,
			info.Bucket,
			info.Key,
			minio.RemoveObjectOptions{
				ForceDelete: false,
			},
		)
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to upload the media",
			err,
		)
	}

	return nil
}
