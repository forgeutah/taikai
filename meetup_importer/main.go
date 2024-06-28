package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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

	// create an upsert org request
	// replace this with a grpc call
	req := taikaiv1.UpsertOrgRequest{
		Org: &taikaiv1.Org{
			Name: parsedConfig.ProAccount,
		},
	}
	_, err = db.UpsertOrgs(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// open csv file
	f, err := os.Open(filesDir + "/meetup_groups.csv")
	if err != nil {
		log.Fatal(err)
	}

	// parse csv file
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
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
	err = db.UpsertMeetupGroups(ctx, parsedConfig.ProAccount, groups)
	if err != nil {
		log.Fatal(err)
	}
}
