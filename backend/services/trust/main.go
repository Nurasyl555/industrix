package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	trustpb "github.com/industrix/gen/go/backend/proto/trust/v1"
	"github.com/industrix/pkg/jwt"
	"github.com/industrix/pkg/kafka"
	"github.com/industrix/pkg/logger"
	"github.com/industrix/pkg/postgres"
	"github.com/industrix/pkg/redis"
	"github.com/industrix/services/trust/internal/auth"
	"github.com/industrix/services/trust/internal/company"
	trustgrpc "github.com/industrix/services/trust/internal/grpc"
	"github.com/industrix/services/trust/internal/profile"
	"github.com/industrix/services/trust/internal/repository"
	"github.com/industrix/services/trust/internal/review"
)

func main() {
	serviceName := "trust-service"
	l := logger.New(serviceName)
	logger.SetGlobal(l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgCfg := postgres.DefaultConfig()
	pgCfg.DSN = os.Getenv("DB_DSN")

	pgClient, err := postgres.NewClient(ctx, pgCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer pgClient.Close()

	// Run migrations
	if err := pgClient.RunMigrations("migrations"); err != nil {
		l.Fatal().Err(err).Msg("Failed to run migrations")
	}

	redisCfg := redis.DefaultConfig()
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		redisCfg.Host = addr // Assumes host or host:port, simple implementation
	}
	redisClient, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Kafka Producer
	kafkaProducer, err := kafka.NewProducer(kafka.DefaultConfig())
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect to Kafka, events will be disabled")
		// Not fatal for now, but in prod should probably block
	} else {
		defer kafkaProducer.Close()
	}

	// Load or generate RSA keys for JWT
	privateKey, publicKey := loadRSAKeys(l)
	jwtClient := jwt.NewClient(privateKey, publicKey)

	// Initialize repositories
	repo := repository.NewRepository(pgClient, redisClient)

	// Initialize services
	authSvc := auth.NewService(repo, jwtClient)
	profileSvc := profile.NewService(repo)
	companySvc := company.NewService(repo, kafkaProducer) // Inject producer
	reviewSvc := review.NewService(repo)

	// gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to listen for gRPC")
	}

	grpcServer := grpc.NewServer()
	trustSvcServer := trustgrpc.NewServer(repo, repo, repo, repo, jwtClient)
	trustpb.RegisterTrustServiceServer(grpcServer, trustSvcServer)
	reflection.Register(grpcServer)

	go func() {
		l.Info().Str("port", grpcPort).Msg("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	// HTTP Server (Fiber)
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Industrix Trust Service",
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

	reviewHandler := review.NewHandler(reviewSvc)
	reviewHandler.RegisterRoutes(api)

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

func loadRSAKeys(l *logger.Logger) (*rsa.PrivateKey, *rsa.PublicKey) {
	privKeyPEM := os.Getenv("JWT_PRIVATE_KEY")
	pubKeyPEM := os.Getenv("JWT_PUBLIC_KEY")

	if privKeyPEM != "" && pubKeyPEM != "" {
		l.Info().Msg("Loading JWT keys from environment variables")

		block, _ := pem.Decode([]byte(privKeyPEM))
		if block == nil {
			l.Fatal().Msg("Failed to parse JWT private key PEM")
		}
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			// Try PKCS8
			k, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err2 != nil {
				l.Fatal().Err(err).Msg("Failed to parse JWT private key")
			}
			privKey = k.(*rsa.PrivateKey)
		}

		blockPub, _ := pem.Decode([]byte(pubKeyPEM))
		if blockPub == nil {
			l.Fatal().Msg("Failed to parse JWT public key PEM")
		}
		pubKeyInterface, err := x509.ParsePKIXPublicKey(blockPub.Bytes)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to parse JWT public key")
		}
		pubKey := pubKeyInterface.(*rsa.PublicKey)

		return privKey, pubKey
	}

	l.Warn().Msg("JWT keys not provided in environment, generating ephemeral keys (NOT FOR PRODUCTION)")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to generate RSA keys")
	}
	return privateKey, &privateKey.PublicKey
}
