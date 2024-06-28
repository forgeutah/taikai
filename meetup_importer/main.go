package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/catalystsquad/app-utils-go/logging"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	pg "github.com/forgeutah/taikai/server/storage/postgres"
)

// TODO: add users to db

type Config struct {
	Key         string `json:"kolla_key"`
	ConsumerID  string `json:"consumer_id"`
	ConnectorID string `json:"connector_id"`
	ProAccount  string `json:"pro_account"`
}

func parseConfig(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return Config{}, fmt.Errorf("unable to open config file: %s", err)
	}
	defer f.Close()
	jsonParser := json.NewDecoder(f)
	config := Config{}
	if err = jsonParser.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("unable to parse config file: %s", err)
	}
	return config, nil
}

// go run main.go -config-path=../../Meetup-Go-Graphql-Scraper/forge.json --file-dir=../../Meetup-Go-Graphql-Scraper/csv/
func main() {
	var configPath string
	var filesDir string
	flag.StringVar(&configPath, "config-path", "config.json", "config file path")
	flag.StringVar(&filesDir, "file-dir", "", "directory to save csv files. ex: /tmp")
	flag.Parse()

	// parse config file
	parsedConfig, err := parseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// connect to db
	db := &pg.PostgresStorage{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	storageShutdown, err := db.Initialize(ctx)
	if err != nil {
		logging.Log.WithError(err).Error("error initializing storage")
	}
	if storageShutdown != nil {
		defer storageShutdown()
	}

	if !db.Ready(ctx) {
		logging.Log.Error("storage not ready")
	}

	// create org
	// err = CreateOrg(ctx, db, parsedConfig.ProAccount)
	// if err != nil {
	// 	logging.Log.WithError(err).Error("error creating org")
	// }

	// add groups
	// err = AddGroups(ctx, db, filesDir, parsedConfig.ProAccount)
	// if err != nil {
	// 	logging.Log.WithError(err).Error("error adding groups")
	// }

	// add users
	err = AddUsers(ctx, db, filesDir, parsedConfig.ProAccount)
	if err != nil {
		logging.Log.WithError(err).Error("error adding users")
	}

}

func CreateOrg(ctx context.Context, db *pg.PostgresStorage, orgName string) error {
	// create an upsert org request
	// replace this with a grpc call
	req := taikaiv1.UpsertOrgRequest{
		Org: &taikaiv1.Org{
			Name: orgName,
		},
	}
	_, err := db.UpsertOrgs(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create meetup %w", err)
	}
	return nil
}

func AddGroups(ctx context.Context, db *pg.PostgresStorage, filesDir, orgName string) error {
	// open csv file
	f, err := os.Open(filesDir + "/meetup_groups.csv")
	if err != nil {
		return fmt.Errorf("failed to open file %w", err)
	}

	// parse csv file
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv file %w", err)
	}
	var groups []taikaiv1.Group
	for i, record := range records {
		if i == 0 {
			continue
		}
		groups = append(groups, taikaiv1.Group{
			Name:     &record[1],
			MeetupId: &record[0],
		})
	}
	err = db.UpsertMeetupGroups(ctx, orgName, groups)
	if err != nil {
		return fmt.Errorf("failed to add groups to org: %s. %w", orgName, err)
	}
	return nil
}

func AddUsers(ctx context.Context, db *pg.PostgresStorage, filesDir, orgName string) error {
	// open csv file
	f, err := os.Open(filesDir + "/meetup_users.csv")
	if err != nil {
		return fmt.Errorf("failed to open file %w", err)
	}

	// parse csv file
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv file %w", err)
	}

	// TODO: query db for org id

	for i, record := range records {
		if i == 0 {
			continue
		}

		groupUUID, err := db.GetGroupIDFromMeetup(ctx, record[0])
		if err != nil {
			return fmt.Errorf("failed to get group id %w", err)
		}
		log.Printf("groupName: %s, groupUUID: %s\n", record[1], groupUUID)

		orgUUID, err := db.GetOrgID(ctx, orgName)
		if err != nil {
			return fmt.Errorf("failed to get org id %w", err)
		}

		// Column headers
		// err = csvWriter.Write([]string{"group_id", "group_name", "member_id", "member_name", "member_email", "member_username", "state", "city", "zip", "isOrganizer"})

		fullname := record[3]
		splitName := strings.Split(fullname, " ")

		isAdmin, _ := strconv.ParseBool(record[9])

		user := taikaiv1.User{
			FirstName: &splitName[0],
			LastName:  &splitName[1],
			Username:  &record[5],
			Email:     &record[4],
			State:     &record[6],
			City:      &record[7],
			Zip:       &record[8],
			MeetupId:  &record[2],
			OrgId:     &orgUUID,
		}
		err = db.AddUser(ctx, &user, groupUUID)
		if err != nil {
			return fmt.Errorf("failed to add user %w", err)
		}

		if isAdmin {
			log.Printf("user %s is an admin of group: %s", record[3], record[1])
		}
	}

	log.Printf("added %d users to db", len(records)-1)
	return nil
}
