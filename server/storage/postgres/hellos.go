package postgres

import (
	"context"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/joomcode/errorx"
)

func (p PostgresStorage) UpsertHellos(ctx context.Context, request *taikaiv1.UpsertHellosRequest) ([]*taikaiv1.Hello, error) {
	helloProtos := taikaiv1.HelloProtos(request.Hellos)
	_, err := helloProtos.Upsert(ctx, db)
	return helloProtos, err
}

func (p PostgresStorage) DeleteHellos(ctx context.Context, ids []string) error {
	return taikaiv1.DeleteHelloGormModels(ctx, db, ids)
}

func (p PostgresStorage) ListHellos(ctx context.Context, request *taikaiv1.ListRequest) ([]*taikaiv1.Hello, error) {
	protos := taikaiv1.HelloProtos{}
	var orderBy string
	if orderBy = request.OrderBy; orderBy == "" {
		orderBy = "created_at"
	}
	err := protos.List(ctx, db, int(request.Limit), int(request.Offset), orderBy)
	return protos, err
}

func (p PostgresStorage) GetHellos(ctx context.Context, request *taikaiv1.GetRequest) ([]*taikaiv1.Hello, error) {
	protos := taikaiv1.HelloProtos{}
	err := protos.GetByIds(ctx, db, request.Ids)
	return protos, err
}

func (p PostgresStorage) GetHello(ctx context.Context, id string) (*taikaiv1.Hello, error) {
	var err error
	protos := taikaiv1.HelloProtos{}
	if err = protos.GetByIds(ctx, db, []string{id}); err != nil {
		return nil, err
	}
	if len(protos) == 0 {
		return nil, errorx.IllegalArgument.New("hello with id %s not found", id)
	}
	return protos[0], nil
}
