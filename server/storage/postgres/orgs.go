package postgres

import (
	"context"

	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *PostgresStorage) UpsertOrgs(ctx context.Context, request taikaiv1.UpsertOrgRequest) (*emptypb.Empty, error) {
	pb := taikaiv1.UpsertOrgRequest(request)
	query := `INSERT INTO orgs (org_name)
						VALUES ($1)
						ON CONFLICT (org_name) DO NOTHING 
						`
	_, err := p.db.ExecContext(ctx, query, pb.Org.Name)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (p *PostgresStorage) ListOrgs(ctx context.Context, request *taikaiv1.ListOrgRequest) (*taikaiv1.ListOrgResponse, error) {
	query := `SELECT id, org_name FROM orgs where org_name = $1`
	filterParam := request.GetName()

	if request.GetName() == "" {
		query = `SELECT id, org_name FROM orgs where id = $1`
		filterParam = request.GetId()
	}
	rows, err := p.db.QueryContext(ctx, query, filterParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orgs := []*taikaiv1.Org{}
	for rows.Next() {
		org := &taikaiv1.Org{}
		err := rows.Scan(&org.Id, &org.Name)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return &taikaiv1.ListOrgResponse{Orgs: orgs}, nil
}
