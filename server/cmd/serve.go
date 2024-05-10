package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/catalystsquad/app-utils-go/errorutils"
	"github.com/catalystsquad/app-utils-go/logging"
	"github.com/catalystsquad/grpc-base-go/pkg"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/forgeutah/taikai/server/config"
	"github.com/forgeutah/taikai/server/handlers"
	"github.com/forgeutah/taikai/server/storage"
	"github.com/forgeutah/taikai/server/storage/postgres"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)
// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start and serve the taikai api",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

var Server *pkg.GrpcServer
var grpcConfig pkg.GrpcServerConfig

func serve() {
	var storageShutdown func()
	var err error
	// init grpc server config
	grpcConfig = config.InitAPIConfig()
	// init grpc server
	Server, err = pkg.NewGrpcServer(grpcConfig)
	if err != nil {
		errorutils.LogOnErr(nil, "error initializing grpc server", err)
		return
	}
	// init storage
	if storageShutdown, err = initializeStorage(); err != nil {
		return
	}
	if storageShutdown != nil {
		defer storageShutdown()
	}
	// register service implementations
	registerServices()
	// run the gateway in the background
	runGateway()
	// run the grpc server
	err = Server.Run()
	errorutils.LogOnErr(nil, "error running grpc server", err)
}

func initializeStorage() (func(), error) {
	switch config.StorageType {
	case storage.PostgresStorageType:
		storage.Storage = postgres.PostgresStorage{}
	}
	return storage.Storage.Initialize()
}

func registerServices() {
	var apiServer taikaiv1.ApiServer = &handlers.ApiServer{}
	taikaiv1.RegisterApiServer(Server.Server, apiServer)
}

func runGateway() {
	grpcAddress := fmt.Sprintf("localhost:%d", grpcConfig.Port)
	httpAddress := fmt.Sprintf(":%d", config.GatewayPort)
	mux := runtime.NewServeMux(runtime.WithMetadata(func(_ context.Context, req *http.Request) metadata.MD {
		return metadata.New(map[string]string{
			"grpcgateway-http-path": req.URL.Path,
		})
	}))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := taikaiv1.RegisterApiHandlerFromEndpoint(context.Background(), mux, grpcAddress, opts)
	errorutils.PanicOnErr(nil, "error registering grpc gateway api handler", err)
	// forever loop to restart on crash
	go func(httpAddress string, mux *runtime.ServeMux) {
		for {
			logging.Log.WithFields(logrus.Fields{"address": httpAddress}).Info("http gateway started")
			err = http.ListenAndServe(httpAddress, cors.AllowAll().Handler(mux))
			errorutils.LogOnErr(nil, "error running grpc gateway", err)
			time.Sleep(config.GatewayRestartDelay)
		}
	}(httpAddress, mux)
}
