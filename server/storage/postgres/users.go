package postgres

import (
	"context"
	"fmt"

	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (s *PostgresStorage) AddUser(ctx context.Context, user *taikaiv1.User, groupUUID uuid.UUID) error {
	// insert user into db
	query := `INSERT INTO users 
	(meetup_id, first_name, last_name, username, email, state, city, zip, group_id, org_id) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (meetup_id) 
	DO UPDATE SET group_id = array_append(users.group_id, $11)
	`
	gArray := pq.Array([]string{groupUUID.String()})
	_, err := s.db.ExecContext(
		ctx,
		query,
		user.MeetupId,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Email,
		user.State,
		user.City,
		user.Zip,
		gArray,
		user.OrgId,
		groupUUID)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to insert user, %s: %w", *user.MeetupId, err))
	}
	return nil
}
