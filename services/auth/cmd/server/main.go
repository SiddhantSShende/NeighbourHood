package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "neighbourhood/proto/gen/go/auth"
	"neighbourhood/services/auth/internal/config"
	"neighbourhood/services/auth/internal/delivery/grpc/handler"
	"neighbourhood/services/auth/internal/repository/postgres"
	"neighbourhood/services/auth/internal/repository/redis"
	"neighbourhood/services/auth/internal/usecase"
	"neighbourhood/services/auth/pkg/logger"
	"neighbourhood/services/auth/pkg/metrics"
	"neighbourhood/services/auth/pkg/tracing"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/auth.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	defer log.Sync()

	log.Info("Starting Auth Service",
		"port", cfg.Server.Port,
		"environment", os.Getenv("ENV"),
	)

	// Initialize tracing
	if cfg.Tracing.Enabled {
		shutdown, err := tracing.InitTracer(cfg.Tracing.ServiceName, cfg.Tracing.Endpoint, cfg.Tracing.SampleRate)
		if err != nil {
			log.Error("Failed to initialize tracer", "error", err)
		} else {
			defer shutdown()
			log.Info("Tracing initialized", "endpoint", cfg.Tracing.Endpoint)
		}
	}

	// Initialize metrics
	metricsServer := metrics.NewServer(cfg.Metrics.Port, cfg.Metrics.Path)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Error("Failed to start metrics server", "error", err)
		}
	}()
	defer metricsServer.Shutdown()

	// Initialize PostgreSQL repository
	pgRepo, err := postgres.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", "error", err)
	}
	defer pgRepo.Close()
	log.Info("Connected to PostgreSQL", "host", cfg.Database.Host)

	// Initialize Redis repository
	redisRepo, err := redis.New(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redisRepo.Close()
	log.Info("Connected to Redis", "host", cfg.Redis.Host)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		pgRepo,
		redisRepo,
		cfg.JWT,
		cfg.OAuth,
		cfg.Security,
		log,
	)

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.Server.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.Server.GRPC.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.Server.GRPC.Keepalive.Time,
			Timeout: cfg.Server.GRPC.Keepalive.Timeout,
		}),
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			tracing.UnaryServerInterceptor(),
			log.UnaryServerInterceptor(),
		),
	)

	// Register auth service
	authHandler := handler.NewAuthHandler(authUseCase, log)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("auth-service", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection (for grpcurl and debugging)
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatal("Failed to listen", "error", err)
	}

	go func() {
		log.Info("gRPC server listening", "port", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Auth Service...")

	// Set health check to NOT_SERVING
	healthServer.SetServingStatus("auth-service", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Graceful stop with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("Auth Service stopped gracefully")
	case <-ctx.Done():
		log.Warn("Shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	}
}
