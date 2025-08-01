package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"dployr.io/pkg/models"
)

type EventRepo struct {
	*Repository[models.Event]
}

func NewEventRepo(db *sqlx.DB) *EventRepo {
	return &EventRepo{
		Repository: NewRepository[models.Event](db, "events"),
	}
}

func (r *EventRepo) Create(ctx context.Context, e *models.Event) error {
	return r.Repository.Create(ctx, e)
}

func (r *EventRepo) Update(ctx context.Context, e models.Event) error {
	return r.Repository.Update(ctx, e)
}

func (r *EventRepo) Upsert(ctx context.Context, e models.Event, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, e, conflictCols, updateCols)
}

func (r *EventRepo) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *EventRepo) GetByID(ctx context.Context, id any) (models.Event, error) {
	return r.Repository.GetByID(ctx, id)
}
