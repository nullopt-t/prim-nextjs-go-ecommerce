package tag

import (
	"context"
	"fmt"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type TagRepository struct {
}

func NewRepository() *TagRepository {
	return &TagRepository{}
}

func (tr *TagRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	tag *model.Tag,
) error {
	query := `
	INSERT INTO tags(
		id,
		name,
		created_at,
		updated_at
	)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
	`
	_, err := qe.Exec(
		ctx,
		query,
		tag.ID,
		tag.Name,
		tag.CreatedAt,
		tag.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create tag:%w", err)
	}

	return nil
}
