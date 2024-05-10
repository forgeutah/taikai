package handlers

import (
	"context"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/forgeutah/taikai/server/storage"
)

func (a ApiServer) UpsertHellos(ctx context.Context, request *taikaiv1.UpsertHellosRequest) (*taikaiv1.Hellos, error) {
	upsertedHellos, err := storage.Storage.UpsertHellos(ctx, request)
	return &taikaiv1.Hellos{Hellos: upsertedHellos}, err
}

func (a ApiServer) DeleteHellos(ctx context.Context, request *taikaiv1.DeleteRequest) (*taikaiv1.DeleteResponse, error) {
	return &taikaiv1.DeleteResponse{}, storage.Storage.DeleteHellos(ctx, request.Ids)
}

func (a ApiServer) ListHellos(ctx context.Context, request *taikaiv1.ListRequest) (*taikaiv1.Hellos, error) {
	if request.Limit == 0 {
		request.Limit = 100
	}
	hellos, err := storage.Storage.ListHellos(ctx, request)
	return &taikaiv1.Hellos{Hellos: hellos}, err
}

func (a ApiServer) GetHellos(ctx context.Context, request *taikaiv1.GetRequest) (*taikaiv1.Hellos, error) {
	hellos, err := storage.Storage.GetHellos(ctx, request)
	return &taikaiv1.Hellos{Hellos: hellos}, err
}
