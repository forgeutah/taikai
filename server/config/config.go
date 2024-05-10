package config

import (
	"github.com/forgeutah/taikai/server/storage"

	"github.com/catalystsquad/app-utils-go/env"
	"github.com/catalystsquad/grpc-base-go/pkg"
	"google.golang.org/grpc"
)

const CorndogsQueueType = "corndogs"

var GrpcMaxMsgSizeBytes = env.GetEnvAsIntOrDefault("GRPC_MAX_MSG_SIZE_BYTES", "52428800") // 50MB
var port = env.GetEnvAsIntOrDefault("PORT", "6000")
var GatewayPort = env.GetEnvAsIntOrDefault("GATEWAY_PORT", "6001")
var prometheusEnabled = env.GetEnvAsBoolOrDefault("PROMETHEUS_ENABLED", "false")
var promethusPath = env.GetEnvOrDefault("PROMETHEUS_PATH", "/metrics")
var prometheusPort = env.GetEnvAsIntOrDefault("PROMETHEUS_PORT", "9090")
var enablePrometheusLatencyHistograsm = env.GetEnvAsBoolOrDefault("PROMETHEUS_LATENCY_HISTOGRAMS_ENABLED", "true")
var GatewayRestartDelay = env.GetEnvAsDurationOrDefault("GATEWAY_RESTART_DELAY", "5s")
var StorageType = env.GetEnvOrDefault("STORAGE_TYPE", storage.PostgresStorageType)
var CorndogsUri = env.GetEnvOrDefault("CORNDOGS_URI", "localhost:5080")
var CorndogsIndexTaskTimeout = int64(env.GetEnvAsIntOrDefault("CORNDOGS_INDEX_TASK_TIMEOUT", "30"))
var AssemblyLineRunnerInterval = env.GetEnvAsDurationOrDefault("ASSEMBLYLINE_RUNNER_INTERVAL", "1s")
var AssemblyLineRunnerPoolSize = env.GetEnvAsIntOrDefault("ASSEMBLYLINE_RUNNER_POOL_SIZE", "10")
var AssemblyLineRunnerTaskExpiration = env.GetEnvAsDurationOrDefault("ASSEMBLYLINE_RUNNER_TASK_EXPIRATION", "30s")
var AssemblyLineQueueTaskTimeoutSeconds = env.GetEnvAsIntOrDefault("ASSEMBLYLINE_QUEUE_TASK_TIMEOUT_SECONDS", "30")
var AssemblyLineInstanceRunnerInterval = env.GetEnvAsDurationOrDefault("ASSEMBLYLINE_INSTANCE_RUNNER_INTERVAL", "1s")
var AssemblyLineInstanceRunnerPoolSize = env.GetEnvAsIntOrDefault("ASSEMBLYLINE_INSTANCE_RUNNER_POOL_SIZE", "10")
var AssemblyLineInstanceRunnerTaskExpiration = env.GetEnvAsDurationOrDefault("ASSEMBLYLINE_INSTANCE_RUNNER_TASK_EXPIRATION", "30s")
var AssemblyLineInstanceQueueTaskTimeoutSeconds = env.GetEnvAsIntOrDefault("ASSEMBLYLINE_INSTANCE_QUEUE_TASK_TIMEOUT_SECONDS", "30")
var StepInstanceRunnerInterval = env.GetEnvAsDurationOrDefault("STEP_INSTANCE_RUNNER_INTERVAL", "1s")
var StepInstanceRunnerPoolSize = env.GetEnvAsIntOrDefault("STEP_INSTANCE_RUNNER_POOL_SIZE", "10")
var StepInstanceRunnerTaskExpiration = env.GetEnvAsDurationOrDefault("STEP_INSTANCE_RUNNER_TASK_EXPIRATION", "30s")
var StepInstanceQueueTaskTimeoutSeconds = env.GetEnvAsIntOrDefault("STEP_INSTANCE_QUEUE_TASK_TIMEOUT_SECONDS", "30")
var ZitadelIssuer = env.GetEnvOrDefault("ZITADEL_ISSUER", "")
var ZitadelAuthUrl = env.GetEnvOrDefault("ZITADEL_AUTH_URL", "")
var ZitadelPrivateKey = env.GetEnvOrDefault("ZITADEL_PRIVATE_KEY", "")
var AuthEnabled = env.GetEnvAsBoolOrDefault("AUTH_ENABLED", "true")

func InitAPIConfig() pkg.GrpcServerConfig {
	grpcConfig := pkg.GrpcServerConfig{
		Port:                              port,
		SentryEnabled:                     false,
		PrometheusEnabled:                 prometheusEnabled,
		PrometheusPath:                    promethusPath,
		PrometheusPort:                    prometheusPort,
		PrometheusEnableLatencyHistograms: enablePrometheusLatencyHistograsm,
		GetErrorToReturn: func(err error) error {
			return err
		},
		CaptureRecoveredErr: func(err error) bool {
			return true
		},
		CaptureErrormessage: "error handling request",
		Opts: []grpc.ServerOption{
			grpc.MaxMsgSize(GrpcMaxMsgSizeBytes),
			grpc.MaxRecvMsgSize(GrpcMaxMsgSizeBytes),
			grpc.MaxSendMsgSize(GrpcMaxMsgSizeBytes),
		},
	}

	return grpcConfig
}
