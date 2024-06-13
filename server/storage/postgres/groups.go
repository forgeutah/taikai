package postgres

import (
	"context"
	"fmt"

	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
)

func (p *PostgresStorage) UpsertMeetupGroups(ctx context.Context, orgName string, groups []taikaiv1.Group) error {
	orgID, err := p.GetOrgID(ctx, orgName)
	if err != nil {
		return fmt.Errorf("failed to get groupID %w", err)
	}
	for _, group := range groups {
		query := `
		INSERT INTO groups (name, org_id, meetup_id)
					VALUES ($1, $2, $3);
					`
		_, err := p.db.ExecContext(ctx, query, group.Name, orgID, group.MeetupId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) GetOrgID(ctx context.Context, orgName string) (string, error) {
	var orgID string
	query := `SELECT id FROM orgs WHERE org_name = $1`
	err := p.db.QueryRowContext(ctx, query, orgName).Scan(&orgID)
	if err != nil {
		return "", err
	}
	return orgID, nil
}
