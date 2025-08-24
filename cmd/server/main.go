package main

import (
	"context"
	"fmt"
	"github.com/tinyurl-pestebani/go-otel-setup/pkg/otel"
	pb "github.com/tinyurl-pestebani/go-proto-pkg/pkg/pb/v1"
	"github.com/tinyurl-pestebani/statistics-database/pkg/db_layer"
	"github.com/tinyurl-pestebani/statistics-service/pkg/config"
	"github.com/tinyurl-pestebani/statistics-service/pkg/service"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	name = "main"
)

var (
	logger = otelslog.NewLogger(name)
)

func main() {
	// 1. Create a cancellable parent context for the entire application's lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.NewServiceConfigFromEnv()
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("[::]:%d", conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Setup otel
	serviceName := "statistics-service"
	serviceVersion := "0.0.1"
	otelShutdown, err := otel.SetOTelSDK(ctx, serviceName, serviceVersion)
	if err != nil {
		log.Fatalf("failed to set up OpenTelemetry: %v", err)
	}

	// Create dependencies
	statsDB, err := db_layer.NewDBLayer()
	if err != nil {
		log.Fatalf("failed creating statistics database: %s", err)
	}

	srv, err := service.NewStatisticsService(statsDB)
	if err != nil {
		log.Fatalf("failed creating service: %s", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStatisticsServiceServer(grpcServer, srv)

	// 2. Set up a channel to listen for OS shutdown signals.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 3. Start the gRPC server in a separate goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Server starting", "port", conf.Port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed to serve", "error", err)
			// If err, terminate the server
			signalChan <- syscall.SIGTERM
		}
	}()

	// 4. Block until a signal is received or the context is cancelled.
	select {
	case <-signalChan:
		logger.Info("Shutdown signal received.")
	case <-ctx.Done():
		logger.Info("Context cancelled.")
	}

	// 5. Initiate graceful shutdown of the gRPC server.
	logger.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	// 6. Call the cancel function to propagate the shutdown signal
	cancel()

	// 7. Create a context with a timeout for the final cleanup.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// 8. Wait for the server goroutine to finish.
	waitGroupDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitGroupDone)
	}()

	select {
	case <-waitGroupDone:
		logger.Info("gRPC server goroutine finished gracefully.")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timed out waiting for server goroutine.")
	}

	// 9. Perform final cleanup of other resources.
	logger.Info("Closing service resources...")
	if err := srv.Close(); err != nil {
		logger.Error("failed to close service", "error", err)
	}

	if err := otelShutdown(shutdownCtx); err != nil {
		logger.Error("failed to shut down OpenTelemetry", "error", err)
	}

	logger.Info("Shutdown complete.")
}
