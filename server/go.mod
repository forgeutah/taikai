module github.com/forgeutah/taikai/server

go 1.22

require (
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/catalystsquad/app-utils-go v1.0.7
	github.com/catalystsquad/grpc-base-go v1.0.5
	github.com/davecgh/go-spew v1.1.1
	github.com/emirpasic/gods v1.18.1
	github.com/forgeutah/taikai/protos v0.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.20.0
	github.com/samber/lo v1.39.0
	github.com/stretchr/testify v1.8.3
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/joomcode/errorx v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/forgeutah/taikai/protos => ./../protos
