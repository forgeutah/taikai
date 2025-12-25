package postgres

import (
	"context"
	"fmt"

	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/google/uuid"
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

// GetGroupIDFromMeetup returns the group id for the given group name. this is a helper function for importing proMeetupGroups to Taikai
func (p *PostgresStorage) GetGroupIDFromMeetup(ctx context.Context, groupMeetupID string) (uuid.UUID, error) {
	var groupID uuid.UUID
	query := `SELECT id FROM groups WHERE meetup_id = $1`
	err := p.db.QueryRowContext(ctx, query, groupMeetupID).Scan(&groupID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return groupID, nil
}
