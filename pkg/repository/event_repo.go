package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	
	"dployr.io/pkg/models"
)

type Event struct {
	*Repository[models.Event]
}

func NewEventRepo(db *sqlx.DB) *Event {
	return &Event{
		Repository: NewRepository[models.Event](db, "events"),
	}
}

func (r *Event) Create(ctx context.Context, e *models.Event) (*models.Event, error) {
	return e, r.Repository.Create(ctx, e)
}

func (r *Event) Update(ctx context.Context, e *models.Event) error {
	return r.Repository.Update(ctx, e)
}

func (r *Event) Upsert(ctx context.Context, e *models.Event, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, e, conflictCols, updateCols)
}

func (r *Event) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *Event) GetByID(ctx context.Context, id any) (*models.Event, error) {
	return r.Repository.GetByID(ctx, id)
}

