package main

import (
	"context"
	"time"

	"github.com/catalystsquad/app-utils-go/logging"
	pg "github.com/forgeutah/taikai/server/storage/postgres"
)

func main() {
	// init grpc server
	// grpcConfig = config.InitAPIConfig()
	// Server, err = pkg.NewGrpcServer(grpcConfig)
	// if err != nil {
	// 	errorutils.LogOnErr(nil, "error initializing grpc server", err)
	// 	return
	// }
	// init storage
	db := &pg.PostgresStorage{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	storageShutdown, err := db.Initialize(ctx)
	if err != nil {
		logging.Log.WithError(err).Error("error initializing storage")
	}
	if storageShutdown != nil {
		defer storageShutdown()
	}
	// register service implementations
	// registerServices()
	// // run the gateway in the background
	// runGateway()
	// // run the grpc server
	// err = Server.Run()
	// errorutils.LogOnErr(nil, "error running grpc server", err)
}

// func registerServices() {
// 	var apiServer taikaiv1.ApiServer = &handlers.ApiServer{}
// 	taikaiv1.RegisterApiServer(Server.Server, apiServer)
// }

// func runGateway() {
// 	grpcAddress := fmt.Sprintf("localhost:%d", grpcConfig.Port)
// 	httpAddress := fmt.Sprintf(":%d", config.GatewayPort)
// 	mux := runtime.NewServeMux(runtime.WithMetadata(func(_ context.Context, req *http.Request) metadata.MD {
// 		return metadata.New(map[string]string{
// 			"grpcgateway-http-path": req.URL.Path,
// 		})
// 	}))
// 	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
// 	err := taikaiv1.RegisterApiHandlerFromEndpoint(context.Background(), mux, grpcAddress, opts)
// 	errorutils.PanicOnErr(nil, "error registering grpc gateway api handler", err)
// 	// forever loop to restart on crash
// 	go func(httpAddress string, mux *runtime.ServeMux) {
// 		for {
// 			logging.Log.WithFields(logrus.Fields{"address": httpAddress}).Info("http gateway started")
// 			err = http.ListenAndServe(httpAddress, cors.AllowAll().Handler(mux))
// 			errorutils.LogOnErr(nil, "error running grpc gateway", err)
// 			time.Sleep(config.GatewayRestartDelay)
// 		}
// 	}(httpAddress, mux)
// }
