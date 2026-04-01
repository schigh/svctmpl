package service

import (
	"context"

	"github.com/example/myservice/internal/model"
)

// ResourceRepository defines the persistence operations the service layer
// requires. This interface is defined by the consumer (service) rather than
// the implementor (repository), following the Go convention for
// consumer-defined interfaces.
type ResourceRepository interface {
	Create(ctx context.Context, r *model.Resource) (*model.Resource, error)
	GetByID(ctx context.Context, id string) (*model.Resource, error)
	List(ctx context.Context) ([]*model.Resource, error)
	Update(ctx context.Context, r *model.Resource) (*model.Resource, error)
	Delete(ctx context.Context, id string) error
}
