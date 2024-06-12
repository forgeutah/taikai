package handlers

import (
	"context"
	"errors"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/forgeutah/taikai/server/storage"
)

type ApiServer struct{}

func (a ApiServer) Healthy(ctx context.Context, empty *taikaiv1.Empty) (*taikaiv1.Empty, error) {
	return &taikaiv1.Empty{}, nil
}

func (a ApiServer) Ready(ctx context.Context, empty *taikaiv1.Empty) (*taikaiv1.Empty, error) {
	if storage.Storage.Ready() {
		return &taikaiv1.Empty{}, nil
	}
	return nil, errors.New("")
}

func 
