package storage

import (
	"context"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
)

var Storage StorageInterface

type StorageInterface interface {
	Initialize() (shutdown func(), err error)
	Ready() bool
	UpsertHellos(ctx context.Context, request *taikaiv1.UpsertHellosRequest) ([]*taikaiv1.Hello, error)
	DeleteHellos(ctx context.Context, ids []string) error
	ListHellos(ctx context.Context, request *taikaiv1.ListRequest) ([]*taikaiv1.Hello, error)
	GetHellos(ctx context.Context, request *taikaiv1.GetRequest) ([]*taikaiv1.Hello, error)
	GetHello(ctx context.Context, id string) (*taikaiv1.Hello, error)
}
