package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	identityv1 "github.com/industrix/gen/go/identity/v1"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/pkg/logger"
	"github.com/industrix/pkg/postgres"
	"github.com/industrix/pkg/redis"
	"github.com/industrix/services/identity/internal/auth"
	"github.com/industrix/services/identity/internal/company"
	identitygrpc "github.com/industrix/services/identity/internal/grpc"
	"github.com/industrix/services/identity/internal/profile"
)

func main() {
	serviceName := "identity-service"
	l := logger.New(serviceName)
	logger.SetGlobal(l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgCfg := postgres.DefaultConfig()
	pgCfg.Database = "identity_db"
	pgClient, err := postgres.NewClient(ctx, pgCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer pgClient.Close()

	redisCfg := redis.DefaultConfig()
	redisClient, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	jwtClient := jwt.NewClient(nil)

	authRepo := auth.NewRepository(pgClient, redisClient)
	profileRepo := profile.NewRepository(pgClient)
	companyRepo := company.NewRepository(pgClient)

	authSvc := auth.NewService(authRepo, jwtClient)
	profileSvc := profile.NewService(profileRepo, nil, nil)
	companySvc := company.NewService(companyRepo, nil, nil)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to listen for gRPC")
	}

	grpcServer := grpc.NewServer()
	identitySvcServer := identitygrpc.NewServer(authRepo, profileRepo, jwtClient)
	identityv1.RegisterIdentityServiceServer(grpcServer, identitySvcServer)
	reflection.Register(grpcServer)

	go func() {
		l.Info().Str("port", grpcPort).Msg("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Industrix Identity Service",
	})
	app.Use(recover.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterRoutes(api)

	profileHandler := profile.NewHandler(profileSvc)
	profileHandler.RegisterRoutes(api)

	companyHandler := company.NewHandler(companySvc)
	companyHandler.RegisterRoutes(api)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	go func() {
		l.Info().Str("port", httpPort).Msg("Starting HTTP server")
		if err := app.Listen(fmt.Sprintf(":%s", httpPort)); err != nil {
			l.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	l.Info().Msg("Shutting down gracefully...")
	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		l.Error().Err(err).Msg("HTTP server shutdown error")
	}

	l.Info().Msg("Service stopped")
}
