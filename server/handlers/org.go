package handlers

import (
	"context"

	"github.com/forgeutah/taikai/server/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (a ApiServer) UpsertOrg(ctx context.Context, request *taikaiv1.UpsertOrgRequest) (*emptypb.Empty, error) {
	createdOrg, err := storage.Storage.UpsertOrg(ctx, request)
	return &taikaiv1.Org{Org: createdOrg}, err
}

func (a ApiServer) ListOrg(ctx context.Context, request *taikaiv1.ListOrgRequest) (*taikaiv1.ListOrgResponse, error) {
	orgs, err := storage.Storage.ListOrg(ctx, request)
	return &taikaiv1.Orgs{Orgs: orgs}, err
}
