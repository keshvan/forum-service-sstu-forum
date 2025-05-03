package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/keshvan/forum-service-sstu-forum/config"
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/go-common-forum/httpserver"
	"github.com/keshvan/go-common-forum/jwt"
	"github.com/keshvan/go-common-forum/logger"
	"github.com/keshvan/go-common-forum/postgres"
)

func Run(cfg *config.Config) {
	//Logger
	logger := logger.New("forum-service", cfg.LogLevel)

	//Database
	pg, err := postgres.New(cfg.PG_URL)
	if err != nil {
		log.Fatalf("app - Run - postgres.New")
	}
	defer pg.Close()

	//Repos
	categoryRepo := repo.NewCategoryRepository(pg, logger)
	topicRepo := repo.NewTopicRepository(pg, logger)
	postRepo := repo.NewPostRepository(pg, logger)

	//CLient
	userClient, err := client.New(cfg.GrpcAddress, logger)
	if err != nil {
		log.Fatalf("app - Run - client.New: %v", err)
	}
	defer userClient.Close()

	//Usecase
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, logger)
	topicUsecase := usecase.NewTopicUsecase(topicRepo, categoryRepo, userClient, logger)
	postUsecase := usecase.NewPostUsecase(postRepo, topicRepo, userClient, logger)

	//JWT
	jwt := jwt.New(cfg.Secret, cfg.AccessTTL, cfg.RefreshTTL)

	//HTTP-Server
	httpServer := httpserver.New(cfg.Server)
	controller.SetRoutes(httpServer.Engine, categoryUsecase, topicUsecase, postUsecase, jwt, logger)

	httpServer.Run()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt
}
