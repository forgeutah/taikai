module github.com/forgeutah/taikai/meetup_importer

go 1.22.0

require (
	github.com/catalystsquad/app-utils-go v1.0.8
	github.com/davecgh/go-spew v1.1.1
	github.com/forgeutah/taikai/protos v0.0.0-20240612235959-5045e9cd0437
	github.com/forgeutah/taikai/server v0.0.0-20240612235959-5045e9cd0437
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.20.0 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/forgeutah/taikai/server => ./../server

replace github.com/forgeutah/taikai/protos => ./../protos
