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
	(meetup_id, first_name, last_name, username, email, state, city, zip, groupIds, org_id) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (meetup_id) 
	DO UPDATE SET groupIds = array_append(users.groupIds, $10)
	`
	gArray := pq.GenericArray{}
	err := gArray.Scan(groupUUID)
	_, err = s.db.ExecContext(
		ctx,
		query,
		user.MeetupId,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Email,
		user.State,
		user.Zip,
		gArray,
		user.OrgId,
		groupUUID)
	if err != nil {
		return fmt.Errorf("failed to insert user into db %v.", err)
	}
	return nil
}
