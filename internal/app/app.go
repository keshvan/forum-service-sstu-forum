package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/keshvan/forum-service-sstu-forum/config"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/go-common-forum/httpserver"
	"github.com/keshvan/go-common-forum/jwt"
	"github.com/keshvan/go-common-forum/postgres"
)

func Run(cfg *config.Config) {
	//Database
	pg, err := postgres.New(cfg.PG_URL)
	if err != nil {
		log.Fatalf("app - Run - postgres.New")
	}
	defer pg.Close()

	//Repos
	categoryRepo := repo.NewCategoryRepository(pg)
	topicRepo := repo.NewTopicRepository(pg)
	postRepo := repo.NewPostRepository(pg)

	//Usecase
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	topicUsecase := usecase.NewTopicUsecase(topicRepo, categoryRepo)
	postUsecase := usecase.NewPostUsecase(postRepo, topicRepo)

	//JWT
	jwt := jwt.New(cfg.Secret, cfg.AccessTTL, cfg.RefreshTTL)

	//HTTP-Server
	httpServer := httpserver.New(cfg.Server)
	controller.SetRoutes(httpServer.Engine, categoryUsecase, topicUsecase, postUsecase, jwt)

	httpServer.Run()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt
}
