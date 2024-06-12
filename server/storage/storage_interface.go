package storage

import (
	"context"

	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StorageInterface interface {
	Initialize(ctx context.Context) (shutdown func(), err error)
	Ready(ctx context.Context) bool
}

type OrgEditor interface {
	UpsertOrg(ctx context.Context, request *taikaiv1.UpsertOrgRequest) (*emptypb.Empty, error)
	ListOrg(ctx context.Context, request *taikaiv1.ListOrgRequest) (*taikaiv1.ListOrgResponse, error)
}
